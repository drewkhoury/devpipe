// Copyright 2025 Andrew Khoury
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// devpipe - Fast, local pipeline runner

// Package main implements the devpipe CLI tool for running local development pipelines.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/drew/devpipe/internal/config"
	"github.com/drew/devpipe/internal/dashboard"
	"github.com/drew/devpipe/internal/git"
	"github.com/drew/devpipe/internal/metrics"
	"github.com/drew/devpipe/internal/model"
	"github.com/drew/devpipe/internal/sarif"
	"github.com/drew/devpipe/internal/ui"
	"golang.org/x/sync/errgroup"
)

// Version information (set via ldflags during build)
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

// sliceFlag allows repeating --skip
type sliceFlag []string

func (s *sliceFlag) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *sliceFlag) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func main() {
	// Check for subcommands first
	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "list":
			listCmd()
			return
		case "validate":
			validateCmd()
			return
		case "generate-reports":
			generateReportsCmd()
			return
		case "sarif":
			sarifCmd()
			return
		case "version", "--version", "-v":
			fmt.Printf("devpipe version %s\n", version)
			return
		case "help", "--help", "-h":
			printHelp()
			return
		default:
			// Check if it looks like a subcommand (doesn't start with -)
			if !strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Unknown command: %s\n", arg)

				// Suggest similar commands
				commands := []string{"list", "validate", "generate-reports", "sarif", "version", "help"}
				if suggestion := findSimilarCommand(arg, commands); suggestion != "" {
					fmt.Fprintf(os.Stderr, "Did you mean '%s'?\n", suggestion)
				}
				fmt.Fprintln(os.Stderr)
				printHelp()
				os.Exit(1)
			}
		}
	}

	// CLI flags
	var (
		flagConfig           string
		flagSince            string
		flagOnly             string
		flagUI               string
		flagFixType          string
		flagNoColor          bool
		flagDashboard        bool
		flagFailFast         bool
		flagDryRun           bool
		flagVerbose          bool
		flagFast             bool
		flagIgnoreWatchPaths bool
		flagSkipVals         sliceFlag
	)

	flag.StringVar(&flagConfig, "config", "", "Path to config file (default: config.toml)")
	flag.StringVar(&flagSince, "since", "", "Git ref to compare against (overrides config)")
	flag.StringVar(&flagOnly, "only", "", "Run only specific task(s) by id (comma-separated)")
	flag.StringVar(&flagUI, "ui", "basic", "UI mode: basic, full")
	flag.StringVar(&flagFixType, "fix-type", "", "Fix type: auto, helper, none (overrides config)")
	flag.BoolVar(&flagDashboard, "dashboard", false, "Show dashboard with live progress")
	flag.BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	flag.Var(&flagSkipVals, "skip", "Skip a task by id (can be specified multiple times)")
	flag.BoolVar(&flagFailFast, "fail-fast", false, "Stop on first task failure")
	flag.BoolVar(&flagDryRun, "dry-run", false, "Do not execute commands, simulate only")
	flag.BoolVar(&flagVerbose, "verbose", false, "Verbose logging")
	flag.BoolVar(&flagFast, "fast", false, "Skip long running tasks")
	flag.BoolVar(&flagIgnoreWatchPaths, "ignore-watch-paths", false, "Ignore watchPaths and run all tasks")
	flag.Parse()

	// Load configuration first to get UI mode
	cfg, configTaskOrder, phaseNames, taskToPhase, err := config.LoadConfig(flagConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	// Merge with defaults
	mergedCfg := config.MergeWithDefaults(cfg)

	// Validate configuration before running
	result, err := config.ValidateConfig(&mergedCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to validate config: %v\n", err)
		os.Exit(1)
	}
	if !result.Valid {
		fmt.Fprintf(os.Stderr, "ERROR: Configuration validation failed:\n")
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", e.Field, e.Message)
		}
		os.Exit(1)
	}
	if len(result.Warnings) > 0 && flagVerbose {
		for _, w := range result.Warnings {
			fmt.Fprintf(os.Stderr, "WARNING: %s: %s\n", w.Field, w.Message)
		}
	}

	// Parse UI mode (CLI flag overrides config)
	uiModeStr := flagUI
	if flagUI == "basic" && cfg != nil && mergedCfg.Defaults.UIMode != "" {
		// Use config value if CLI flag is default
		uiModeStr = mergedCfg.Defaults.UIMode
	}

	var uiMode ui.UIMode
	switch uiModeStr {
	case "basic":
		uiMode = ui.UIModeBasic
	case "full":
		uiMode = ui.UIModeFull
	default:
		uiMode = ui.UIModeBasic
	}

	// Create renderer
	enableColors := !flagNoColor && ui.IsColorEnabled()
	// Determine if we should use dashboard (animated tracker)
	useAnimated := flagDashboard && ui.IsTTY(uintptr(1))
	renderer := ui.NewRenderer(uiMode, enableColors, useAnimated)

	// Determine project root first (for all path resolution)
	// This can be overridden in config, or auto-detected from git/config location
	// We need to do this before git detection to know where to look for git
	cwdGitRoot, cwdInGitRepo := git.DetectProjectRoot()
	projectRoot := determineProjectRoot(flagConfig, mergedCfg, cwdGitRoot, cwdInGitRepo)

	// Now detect git root from the project root location (for git operations)
	gitRoot, inGitRepo := git.DetectProjectRootFrom(projectRoot)

	// Safety check: prevent running in dangerous system directories
	if !git.IsSafeDirectory(projectRoot) {
		fmt.Fprintf(os.Stderr, "ERROR: Refusing to run devpipe in system directory: %s\n", projectRoot)
		fmt.Fprintf(os.Stderr, "This safety check prevents accidental execution in critical system paths.\n")
		fmt.Fprintf(os.Stderr, "Please run devpipe from your project directory, or set projectRoot in your config.\n")
		os.Exit(1)
	}

	// Auto-generate config.toml if it doesn't exist and no custom config specified
	if cfg == nil && flagConfig == "" {
		defaultConfigPath := "config.toml"

		// Prompt user to create config
		fmt.Printf("No config.toml found. Create one with example tasks? (y/n): ")
		var response string
		_, _ = fmt.Scanln(&response) // Best effort user input

		if response == "y" || response == "Y" || response == "yes" || response == "Yes" {
			if err := config.GenerateDefaultConfig(defaultConfigPath, projectRoot); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: Could not generate config.toml: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✓ Created config.toml - edit it to customize your tasks\n")
			fmt.Printf("  Full reference: https://github.com/drewkhoury/devpipe/blob/main/config.example.toml\n\n")

			// Reload config after generating
			cfg, configTaskOrder, phaseNames, taskToPhase, _ = config.LoadConfig(defaultConfigPath)
			mergedCfg = config.MergeWithDefaults(cfg)
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: No config.toml found. Create one or specify with --config flag\n")
			os.Exit(1)
		}
	}

	// Determine which tasks to use
	var tasks map[string]config.TaskConfig
	var taskOrder []string

	if cfg == nil || len(cfg.Tasks) == 0 {
		// No config file or no tasks defined, use built-in
		if flagVerbose {
			fmt.Println("No config file found, using built-in tasks")
		}
		tasks = config.BuiltInTasks(projectRoot)
		taskOrder = config.GetTaskOrder()
	} else {
		// Use tasks from config
		tasks = mergedCfg.Tasks
		// Use the order extracted from the config file
		if len(configTaskOrder) > 0 {
			taskOrder = configTaskOrder
		} else {
			// Fallback: use built-in order if tasks match, otherwise alphabetical
			builtInOrder := config.GetTaskOrder()
			for _, id := range builtInOrder {
				if _, exists := tasks[id]; exists {
					taskOrder = append(taskOrder, id)
				}
			}
			// Add any additional tasks not in built-in order
			for id := range tasks {
				found := false
				for _, existing := range taskOrder {
					if existing == id {
						found = true
						break
					}
				}
				if !found {
					taskOrder = append(taskOrder, id)
				}
			}
		}

		// Filter out "wait*" and "phase-*" pseudo-tasks (they're just phase markers)
		// Keep them in taskOrder for phase detection, but remove from tasks map
		for id := range tasks {
			if id == "wait" || strings.HasPrefix(id, "wait-") || strings.HasPrefix(id, "phase-") {
				delete(tasks, id)
			}
		}
	}

	// Determine git mode and ref
	gitMode := mergedCfg.Defaults.Git.Mode
	gitRef := mergedCfg.Defaults.Git.Ref

	// CLI --since overrides config
	if flagSince != "" {
		gitMode = "ref"
		gitRef = flagSince
	}

	// Get changed files (uses git root)
	gitInfo := git.DetectChangedFiles(gitRoot, inGitRepo, gitMode, gitRef, flagVerbose)

	// Prepare output dir (uses project root for relative paths, respects absolute paths)
	var outputRoot string
	// Clean the configured path first (handles trailing slashes, .., etc.)
	cleanedOutputRoot := filepath.Clean(mergedCfg.Defaults.OutputRoot)

	if filepath.IsAbs(cleanedOutputRoot) {
		// Use absolute path as-is
		outputRoot = cleanedOutputRoot
	} else {
		// Relative path - join with projectRoot
		outputRoot = filepath.Join(projectRoot, cleanedOutputRoot)
	}

	// Clean the final path
	outputRoot = filepath.Clean(outputRoot)

	// Resolve symlinks (best effort) to check actual destination
	// Try to resolve the full path first
	resolvedOutput, err := filepath.EvalSymlinks(outputRoot)
	if err == nil {
		outputRoot = resolvedOutput
	} else {
		// If full path doesn't exist, try resolving parent directories
		dir := outputRoot
		for dir != "/" && dir != "." {
			if resolved, err := filepath.EvalSymlinks(dir); err == nil {
				// Found a resolvable parent, reconstruct the path
				rel, _ := filepath.Rel(dir, outputRoot)
				outputRoot = filepath.Join(resolved, rel)
				break
			}
			dir = filepath.Dir(dir)
		}
	}

	// Safety check on final resolved path (allow /tmp subdirectories)
	if !git.IsSafeDirectory(outputRoot) && !strings.HasPrefix(outputRoot, "/tmp/") {
		fmt.Fprintf(os.Stderr, "ERROR: Output directory resolves to dangerous location: %s\n", outputRoot)
		fmt.Fprintf(os.Stderr, "This safety check prevents accidental execution in critical system paths.\n")
		fmt.Fprintf(os.Stderr, "Use a safe location like /tmp/devpipe or a relative path within your project.\n")
		os.Exit(1)
	}

	// Verbose logging (after all paths are determined)
	if flagVerbose {
		renderer.Verbose(flagVerbose, "devpipe version: %s (commit: %s, built: %s)", version, commit, buildDate)
		configPath := flagConfig
		if configPath == "" {
			configPath = "config.toml"
		}
		if flagConfig != "" {
			renderer.Verbose(flagVerbose, "Config: %s (from --config)", configPath)
		} else {
			renderer.Verbose(flagVerbose, "Config: %s", configPath)
		}
		if mergedCfg.Defaults.ProjectRoot != "" {
			renderer.Verbose(flagVerbose, "Project root: %s (from config)", projectRoot)
		} else {
			if inGitRepo {
				renderer.Verbose(flagVerbose, "Project root: %s (auto-detected from git)", projectRoot)
			} else {
				renderer.Verbose(flagVerbose, "Project root: %s (auto-detected from config location)", projectRoot)
			}
		}
		if inGitRepo {
			renderer.Verbose(flagVerbose, "Git root: %s (detected by running git from project root)", gitRoot)
		} else {
			renderer.Verbose(flagVerbose, "Git root: %s (no git repo found at project root)", gitRoot)
		}
		renderer.Verbose(flagVerbose, "Output directory: %s", outputRoot)
		fmt.Println() // Blank line before run output
	}
	runID := makeRunID()
	runDir := filepath.Join(outputRoot, "runs", runID)
	logDir := filepath.Join(runDir, "logs")

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to create run directories: %v\n", err)
		os.Exit(1)
	}

	// Load historical averages
	historicalAvg := loadHistoricalAverages(outputRoot)

	// Build task list
	var taskDefs []model.TaskDefinition
	for _, id := range taskOrder {
		// Check if this is a "wait" or "wait-*" marker
		if id == "wait" || strings.HasPrefix(id, "wait-") {
			// Mark the previous task as a wait point (end of phase)
			if len(taskDefs) > 0 {
				taskDefs[len(taskDefs)-1].Wait = true
			}
			continue
		}

		taskCfg, ok := tasks[id]
		if !ok {
			continue
		}

		// Resolve with defaults
		resolved := mergedCfg.ResolveTaskConfig(id, taskCfg, projectRoot)

		// Skip if disabled
		if resolved.Enabled != nil && !*resolved.Enabled {
			renderer.Verbose(flagVerbose, "%s DISABLED in config", id)
			continue
		}

		// Use historical average if available, otherwise use 10s default
		estimatedSeconds := 10
		isGuess := true

		if avgSeconds, hasHistory := historicalAvg[id]; hasHistory {
			// Always prefer historical average
			estimatedSeconds = avgSeconds
			isGuess = false // Historical data, not a guess
		}

		// Get phase name from taskToPhase mapping
		phaseName := ""
		if phaseID, ok := taskToPhase[id]; ok {
			// Look up the phase name using the phase ID
			for _, phaseInfo := range phaseNames {
				if phaseInfo.ID == phaseID {
					phaseName = phaseInfo.Name
					break
				}
			}
		}

		taskDef := model.TaskDefinition{
			ID:               id,
			Name:             resolved.Name,
			Desc:             resolved.Desc,
			Phase:            phaseName,
			Type:             resolved.Type,
			Command:          resolved.Command,
			Workdir:          resolved.Workdir,
			EstimatedSeconds: estimatedSeconds,
			IsEstimateGuess:  isGuess,
			Wait:             resolved.Wait,
		}

		// Add metrics config if present
		if resolved.MetricsFormat != "" {
			taskDef.MetricsFormat = resolved.MetricsFormat
			taskDef.MetricsPath = resolved.MetricsPath
			if flagVerbose && !useAnimated {
				// Only print before dashboard starts; during dashboard it goes to output
				fmt.Printf("[%-15s] %s Metrics configured: format=%s, path=%s\n", renderer.Gray("verbose"), id, resolved.MetricsFormat, resolved.MetricsPath)
			}
		}

		// Add fix config if present
		// Apply CLI flag override if specified
		fixType := resolved.FixType
		if flagFixType != "" {
			fixType = flagFixType
		}
		taskDef.FixType = fixType
		taskDef.FixCommand = resolved.FixCommand

		// Add watchPaths if present
		taskDef.WatchPaths = resolved.WatchPaths

		taskDefs = append(taskDefs, taskDef)
	}

	// Apply CLI filters
	filteredTasks := filterTasks(taskDefs, flagOnly, flagSkipVals, flagFast, mergedCfg.Defaults.FastThreshold, flagVerbose)

	// Apply watchPaths filtering based on git changes (unless --ignore-watch-paths is set)
	if !flagIgnoreWatchPaths && gitInfo.InGitRepo && len(gitInfo.ChangedFiles) >= 0 {
		filteredTasks = filterTasksByWatchPaths(filteredTasks, gitInfo.ChangedFiles, projectRoot, flagVerbose)
	}

	// Run tasks
	var (
		results         []model.TaskResult
		overallExitCode int
		anyFailed       bool
	)

	// Create pipeline.log for verbose output
	pipelineLogPath := filepath.Join(runDir, "pipeline.log")
	pipelineLog, err := os.Create(pipelineLogPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: cannot create pipeline.log: %v\n", err)
		pipelineLog = nil
	}
	if pipelineLog != nil {
		defer func() {
			if err := pipelineLog.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close pipeline log: %v\n", err)
			}
		}()
		// Set pipeline log on renderer so verbose output is captured
		renderer.SetPipelineLog(pipelineLog)
	}

	// Render header
	renderer.RenderHeader(runID, projectRoot, gitMode, len(gitInfo.ChangedFiles))

	// Setup animation if enabled
	var tracker *ui.AnimatedTaskTracker

	// Ensure cursor is restored on exit (in case of panic or early exit)
	defer func() {
		if tracker != nil {
			fmt.Print("\033[?25h") // Show cursor
			fmt.Println()          // Add newline to ensure clean exit
		}
	}()

	// Group tasks into phases based on wait markers
	phases := groupTasksIntoPhases(filteredTasks, phaseNames)

	if renderer.IsAnimated() {
		// Build task progress list with phase information
		var taskProgress []ui.TaskProgress
		for phaseIdx, phase := range phases {
			for _, st := range phase.Tasks {
				taskProgress = append(taskProgress, ui.TaskProgress{
					ID:               st.ID,
					Name:             st.Name,
					Type:             st.Type,
					Phase:            phaseIdx + 1, // 1-indexed phases
					PhaseName:        phase.Name,
					Status:           "PENDING",
					EstimatedSeconds: st.EstimatedSeconds,
					IsEstimateGuess:  st.IsEstimateGuess,
					ElapsedSeconds:   0,
					StartTime:        time.Time{},
				})
			}
		}

		// Calculate header lines
		headerLines := 4 // basic mode: run ID, repo, git mode, changed files
		if renderer.IsAnimated() {
			headerLines = 4
		}

		tracker = renderer.CreateAnimatedTracker(taskProgress, headerLines, mergedCfg.Defaults.AnimationRefreshMs, mergedCfg.Defaults.AnimatedGroupBy)
		if tracker != nil {
			if err := tracker.Start(); err != nil {
				// Animation failed, fall back to non-animated
				if flagVerbose {
					fmt.Fprintf(os.Stderr, "WARNING: Animation not supported, using basic mode\n")
				}
				tracker = nil
			} else {
				// Set tracker on renderer so verbose output can be routed
				renderer.SetTracker(tracker)
			}
		}
	}

	if len(phases) > 1 {
		renderer.Verbose(flagVerbose, "Executing %d phases with parallel tasks", len(phases))
	}

	// Track total pipeline duration
	pipelineStart := time.Now()

	// Set git-related environment variables for all tasks
	if gitInfo.InGitRepo {
		_ = os.Setenv("DEVPIPE_GIT_MODE", gitInfo.Mode)
		_ = os.Setenv("DEVPIPE_GIT_REF", gitInfo.Ref)
		_ = os.Setenv("DEVPIPE_CHANGED_FILES_COUNT", fmt.Sprintf("%d", len(gitInfo.ChangedFiles)))

		// Newline-separated list (handles spaces in filenames)
		_ = os.Setenv("DEVPIPE_CHANGED_FILES", strings.Join(gitInfo.ChangedFiles, "\n"))

		// JSON array (language-agnostic)
		changedFilesJSON, _ := json.Marshal(gitInfo.ChangedFiles)
		_ = os.Setenv("DEVPIPE_CHANGED_FILES_JSON", string(changedFilesJSON))
	}

	// Execute phases sequentially, tasks within each phase in parallel
	var resultsMu sync.Mutex
	var outputMu sync.Mutex // For sequential output display

	for phaseIdx, phase := range phases {
		// Log phase start
		if len(phases) > 1 {
			phaseName := phase.Name
			if phaseName == "" {
				phaseName = fmt.Sprintf("Phase %d", phaseIdx+1)
			}
			if tracker == nil {
				fmt.Printf("\n▶ Starting %s (%d tasks)\n", phaseName, len(phase.Tasks))
			}
			renderer.Verbose(flagVerbose, "Phase %d/%d (%d tasks)", phaseIdx+1, len(phases), len(phase.Tasks))
		}

		// Use errgroup for parallel execution within phase
		g := new(errgroup.Group)
		g.SetLimit(10) // Max 10 concurrent tasks

		var phaseFailed bool
		var phaseFailMu sync.Mutex

		// For sequential output: each task gets a completion channel from the previous task
		var prevTaskDone chan struct{}

		for _, st := range phase.Tasks {
			// Check if should skip due to --fast
			longRunning := st.EstimatedSeconds >= mergedCfg.Defaults.FastThreshold
			if flagFast && longRunning && flagOnly == "" {
				reason := fmt.Sprintf("skipped by --fast (est %ds)", st.EstimatedSeconds)

				// Update tracker if animated
				if tracker != nil {
					tracker.UpdateTask(st.ID, "SKIPPED", 0)
				}

				renderer.RenderTaskSkipped(st.ID, reason, flagVerbose)
				resultsMu.Lock()
				results = append(results, model.TaskResult{
					ID:               st.ID,
					Name:             st.Name,
					Desc:             st.Desc,
					Phase:            st.Phase,
					Type:             st.Type,
					Status:           model.StatusSkipped,
					Skipped:          true,
					SkipReason:       "skipped by --fast",
					Command:          st.Command,
					Workdir:          st.Workdir,
					LogPath:          "",
					EstimatedSeconds: st.EstimatedSeconds,
				})
				resultsMu.Unlock()
				continue
			}

			// Capture task for goroutine
			task := st

			// Create a done channel for this task
			taskDone := make(chan struct{})
			waitForPrev := prevTaskDone
			prevTaskDone = taskDone // Next task will wait for this one

			g.Go(func() error {
				res, taskBuffer, _ := runTask(task, runDir, logDir, flagDryRun, flagVerbose, renderer, tracker, &outputMu, waitForPrev, taskDone)

				// Display buffered output sequentially (always, even in animated mode)
				if taskBuffer != nil && taskBuffer.Len() > 0 {
					outputMu.Lock()
					if tracker != nil {
						// In animated mode, send buffered output to tracker
						lines := strings.Split(strings.TrimRight(taskBuffer.String(), "\n"), "\n")
						for _, line := range lines {
							tracker.AddLogLine(line)
						}
					} else {
						// In non-animated mode, print directly
						fmt.Print(taskBuffer.String())
					}
					outputMu.Unlock()
				}

				resultsMu.Lock()
				results = append(results, res)
				resultsMu.Unlock()

				if res.Status == model.StatusFail {
					phaseFailMu.Lock()
					phaseFailed = true
					anyFailed = true
					overallExitCode = 1
					phaseFailMu.Unlock()

					if flagFailFast {
						if flagVerbose {
							fmt.Printf("[%-15s] FAIL, stopping due to --fail-fast\n", task.ID)
						}
						return fmt.Errorf("task %s failed", task.ID)
					}
				}
				return nil
			})
		}

		// Wait for all tasks in this phase to complete
		if err := g.Wait(); err != nil && flagFailFast {
			// Fail-fast triggered, stop all phases
			break
		}

		// Auto-fix logic: check for failed tasks that have fixType="auto"
		if !flagDryRun {
			resultsMu.Lock()
			var tasksToFix []struct {
				task   model.TaskDefinition
				result model.TaskResult
				index  int
			}

			// Find failed tasks in this phase that need fixing
			for i := len(results) - len(phase.Tasks); i < len(results); i++ {
				res := results[i]
				if res.Status == model.StatusFail {
					// Find the corresponding task definition
					for _, task := range phase.Tasks {
						if task.ID == res.ID && task.FixType == "auto" && task.FixCommand != "" {
							tasksToFix = append(tasksToFix, struct {
								task   model.TaskDefinition
								result model.TaskResult
								index  int
							}{task, res, i})
							break
						}
					}
				}
			}
			resultsMu.Unlock()

			// Run fixes in parallel (same as original tasks)
			if len(tasksToFix) > 0 {
				fixGroup := new(errgroup.Group)
				fixGroup.SetLimit(10)

				for _, item := range tasksToFix {
					task := item.task
					resultIndex := item.index
					originalResult := item.result

					fixGroup.Go(func() error {
						// Open log file for appending
						logFile, err := os.OpenFile(originalResult.LogPath, os.O_APPEND|os.O_WRONLY, 0644)
						if err != nil {
							fmt.Fprintf(os.Stderr, "ERROR: cannot open log file %s: %v\n", originalResult.LogPath, err)
							return nil
						}
						defer func() {
							if err := logFile.Close(); err != nil {
								fmt.Fprintf(os.Stderr, "Warning: failed to close log file: %v\n", err)
							}
						}()

						// Run fix command and time it
						fixCmd := exec.Command("sh", "-c", task.FixCommand)
						fixCmd.Dir = task.Workdir
						fixStart := time.Now()

						// Capture output and write to log
						fixCmd.Stdout = logFile
						fixCmd.Stderr = logFile

						// Write separator to log
						_, _ = fmt.Fprintf(logFile, "\n--- Auto-fix: %s ---\n", task.FixCommand) // Log write

						fixErr := fixCmd.Run()
						fixDuration := time.Since(fixStart)

						// Show fix message with timing
						if tracker != nil {
							tracker.UpdateTask(task.ID, "FIXING", 0)
						} else {
							fmt.Printf("[%-15s] 🔧 %s (%dms)\n", task.ID, renderer.Blue("Auto-fixing: "+task.FixCommand), fixDuration.Milliseconds())
						}

						if fixErr != nil {
							// Fix failed
							if tracker != nil {
								tracker.UpdateTask(task.ID, "FIX FAILED", 0)
							} else {
								fmt.Printf("[%-15s] ❌ %s\n", task.ID, renderer.Red("Failed to fix"))
							}
							return nil // Don't stop other fixes
						}

						// Fix succeeded, re-run original check
						if tracker != nil {
							tracker.UpdateTask(task.ID, "RE-CHECKING", 0)
						} else {
							fmt.Printf("[%-15s] ✅ %s\n", task.ID, renderer.Green("Fix succeeded, re-checking..."))
						}

						// Write separator to log
						_, _ = fmt.Fprintf(logFile, "\n--- Re-check: %s ---\n", task.Command) // Log write

						// Re-run original command
						recheckCmd := exec.Command("sh", "-c", task.Command)
						recheckCmd.Dir = task.Workdir
						recheckCmd.Stdout = logFile
						recheckCmd.Stderr = logFile
						recheckStart := time.Now()
						recheckErr := recheckCmd.Run()
						recheckDuration := time.Since(recheckStart)

						// Calculate total time: original check + fix + recheck
						totalDuration := time.Duration(originalResult.DurationMs)*time.Millisecond + fixDuration + recheckDuration

						// Update result
						resultsMu.Lock()
						if recheckErr == nil {
							// Fixed!
							results[resultIndex].Status = model.StatusPass
							exitCode := 0
							results[resultIndex].ExitCode = &exitCode
							results[resultIndex].DurationMs = totalDuration.Milliseconds()
							results[resultIndex].AutoFixed = true
							results[resultIndex].FixCommand = task.FixCommand
							results[resultIndex].InitialExitCode = originalResult.ExitCode
							results[resultIndex].FixDurationMs = fixDuration.Milliseconds()
							results[resultIndex].RecheckDurationMs = recheckDuration.Milliseconds()

							// Update phase failure status
							phaseFailMu.Lock()
							// Recount failures in this phase
							phaseStillFailed := false
							for i := len(results) - len(phase.Tasks); i < len(results); i++ {
								if results[i].Status == model.StatusFail {
									phaseStillFailed = true
									break
								}
							}
							phaseFailed = phaseStillFailed
							if !phaseStillFailed {
								anyFailed = false
								overallExitCode = 0
							}
							phaseFailMu.Unlock()

							if tracker != nil {
								tracker.UpdateTask(task.ID, "PASS", recheckDuration.Seconds())
							} else {
								fmt.Printf("[%-15s] ✅ %s (%dms)\n", task.ID, renderer.Green("PASS"), recheckDuration.Milliseconds())
							}
						} else {
							// Still failing after fix
							results[resultIndex].DurationMs = totalDuration.Milliseconds()
							results[resultIndex].FixCommand = task.FixCommand
							results[resultIndex].InitialExitCode = originalResult.ExitCode
							results[resultIndex].FixDurationMs = fixDuration.Milliseconds()
							results[resultIndex].RecheckDurationMs = recheckDuration.Milliseconds()
							if tracker != nil {
								tracker.UpdateTask(task.ID, "STILL FAILING", recheckDuration.Seconds())
							} else {
								fmt.Printf("[%-15s] ❌ %s\n", task.ID, renderer.Red("Still failing after fix"))
							}
						}
						resultsMu.Unlock()

						return nil
					})
				}

				_ = fixGroup.Wait() // Wait for all fixes to complete
			}

			// Show helper messages for failed tasks with fixType="helper"
			resultsMu.Lock()
			for i := len(results) - len(phase.Tasks); i < len(results); i++ {
				res := results[i]
				if res.Status == model.StatusFail {
					for _, task := range phase.Tasks {
						if task.ID == res.ID && task.FixType == "helper" && task.FixCommand != "" {
							if tracker == nil {
								fmt.Printf("[%-15s] 💡 %s\n", task.ID, renderer.Yellow("To fix run: "+task.FixCommand))
							}
							break
						}
					}
				}
			}
			resultsMu.Unlock()
		}

		// Log phase completion
		if len(phases) > 1 {
			phaseName := phase.Name
			if phaseName == "" {
				phaseName = fmt.Sprintf("Phase %d", phaseIdx+1)
			}
			phaseFailMu.Lock()
			status := "✓ Complete"
			if phaseFailed {
				status = "✗ Failed"
			}
			phaseFailMu.Unlock()
			if tracker == nil {
				fmt.Printf("◀ %s %s\n", phaseName, status)
			}
		}

		// If phase failed and fail-fast is enabled, stop
		phaseFailMu.Lock()
		shouldStop := phaseFailed && flagFailFast
		phaseFailMu.Unlock()

		if shouldStop {
			if tracker == nil && len(phases) > 1 {
				fmt.Printf("\n⚠ Stopping execution due to phase failure (fail-fast enabled)\n")
			}
			break
		}
	}

	// Calculate total pipeline duration BEFORE the pause
	pipelineDuration := time.Since(pipelineStart)
	totalMs := pipelineDuration.Milliseconds()

	// Stop animation if it was running
	if tracker != nil {
		// Do a final render to show completed state
		time.Sleep(100 * time.Millisecond)

		tracker.Stop()

		// Show completion message and wait for user input
		fmt.Print(renderer.Green("✓ Done") + " - Press Enter to continue...")

		// Wait for Enter key
		_, _ = fmt.Scanln() // Best effort wait for user
	}

	// Render summary
	var summaries []ui.TaskSummary
	for _, r := range results {
		summaries = append(summaries, ui.TaskSummary{
			ID:         r.ID,
			Status:     string(r.Status),
			DurationMs: r.DurationMs,
			AutoFixed:  r.AutoFixed,
		})
	}
	renderer.RenderSummary(summaries, anyFailed, totalMs)

	// Show where to find logs and reports
	fmt.Println()
	fmt.Printf("📁 Run logs:  %s\n", filepath.Join(outputRoot, "runs", runID, "logs"))
	fmt.Printf("📊 Dashboard: %s\n", filepath.Join(outputRoot, "report.html"))

	// Build effective config tracking
	effectiveConfig := buildEffectiveConfig(cfg, &mergedCfg, flagSince, flagUI, uiModeStr, gitMode, gitRef, historicalAvg)

	// Determine the actual config path used
	actualConfigPath := flagConfig
	if actualConfigPath == "" {
		// Check if default config.toml exists
		if _, err := os.Stat("config.toml"); err == nil {
			actualConfigPath = "config.toml"
		}
	}

	// Write run record and generate dashboard
	runRecord := model.RunRecord{
		RunID:           runID,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		ProjectRoot:     projectRoot,
		OutputRoot:      outputRoot,
		ConfigPath:      actualConfigPath,
		Command:         buildCommandString(),
		PipelineVersion: version, // Version used to run the pipeline
		Git:             gitInfo,
		Flags: model.RunFlags{
			Fast:     flagFast,
			FailFast: flagFailFast,
			DryRun:   flagDryRun,
			Verbose:  flagVerbose,
			Only:     flagOnly,
			Skip:     flagSkipVals,
			Config:   flagConfig,
			Since:    flagSince,
		},
		Tasks:           results,
		EffectiveConfig: effectiveConfig,
	}
	if err := writeRunJSON(runDir, runRecord); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to write run record: %v\n", err)
	}

	// Copy config file to run directory
	if err := copyConfigToRun(runDir, flagConfig, &mergedCfg); err != nil {
		if flagVerbose {
			fmt.Fprintf(os.Stderr, "WARNING: failed to copy config: %v\n", err)
		}
	}

	// Generate dashboard (only generate report for current run)
	if err := dashboard.GenerateDashboardWithOptions(outputRoot, version, false, runID); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to generate dashboard: %v\n", err)
	}

	// Final cursor restoration (belt and suspenders)
	fmt.Print("\033[?25h")

	os.Exit(overallExitCode)
}

// loadHistoricalAverages loads task averages from the dashboard summary
func loadHistoricalAverages(outputRoot string) map[string]int {
	averages := make(map[string]int)

	summaryPath := filepath.Join(outputRoot, "summary.json")
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		return averages // No history yet
	}

	var summary struct {
		TaskStats map[string]struct {
			AvgDuration float64 `json:"avgDuration"`
		} `json:"taskStats"`
	}

	if err := json.Unmarshal(data, &summary); err != nil {
		return averages
	}

	// Convert milliseconds to seconds
	for taskID, stats := range summary.TaskStats {
		if stats.AvgDuration > 0 {
			avgSeconds := int(stats.AvgDuration / 1000)
			if avgSeconds < 1 {
				avgSeconds = 1
			}
			averages[taskID] = avgSeconds
		}
	}

	return averages
}

func buildCommandString() string {
	// Build command line from os.Args
	cmdLine := ""
	for i, arg := range os.Args {
		if i > 0 {
			cmdLine += " "
		}
		// Quote arguments that contain spaces
		if len(arg) > 0 && (arg[0] == '-' || !containsSpace(arg)) {
			cmdLine += arg
		} else {
			cmdLine += `"` + arg + `"`
		}
	}

	return cmdLine
}

func containsSpace(s string) bool {
	for _, c := range s {
		if c == ' ' {
			return true
		}
	}
	return false
}

// buildEffectiveConfig creates a detailed breakdown of configuration values and their sources
func buildEffectiveConfig(cfg *config.Config, mergedCfg *config.Config, flagSince, flagUI, uiModeStr, gitMode, gitRef string, _ map[string]int) *model.EffectiveConfig {
	defaults := config.GetDefaults()
	var values []model.ConfigValue

	// Helper to add a config value
	addValue := func(key, value, source, overrode string) {
		values = append(values, model.ConfigValue{
			Key:      key,
			Value:    value,
			Source:   source,
			Overrode: overrode,
		})
	}

	// Output Root
	if cfg != nil && cfg.Defaults.OutputRoot != "" {
		addValue("defaults.outputRoot", mergedCfg.Defaults.OutputRoot, "config-file", "")
	} else {
		addValue("defaults.outputRoot", mergedCfg.Defaults.OutputRoot, "default", "")
	}

	// Fast Threshold
	if cfg != nil && cfg.Defaults.FastThreshold != 0 {
		addValue("defaults.fastThreshold", fmt.Sprintf("%d", mergedCfg.Defaults.FastThreshold), "config-file", "")
	} else {
		addValue("defaults.fastThreshold", fmt.Sprintf("%d", mergedCfg.Defaults.FastThreshold), "default", "")
	}

	// UI Mode
	var uiSource, uiOverrode string
	if flagUI != "basic" {
		uiSource = "cli-flag"
		if cfg != nil && cfg.Defaults.UIMode != "" {
			uiOverrode = cfg.Defaults.UIMode
		} else {
			uiOverrode = defaults.Defaults.UIMode
		}
	} else if cfg != nil && cfg.Defaults.UIMode != "" {
		uiSource = "config-file"
	} else {
		uiSource = "default"
	}
	addValue("defaults.uiMode", uiModeStr, uiSource, uiOverrode)

	// Animation Refresh
	if cfg != nil && cfg.Defaults.AnimationRefreshMs != 0 {
		addValue("defaults.animationRefreshMs", fmt.Sprintf("%d", mergedCfg.Defaults.AnimationRefreshMs), "config-file", "")
	} else {
		addValue("defaults.animationRefreshMs", fmt.Sprintf("%d", mergedCfg.Defaults.AnimationRefreshMs), "default", "")
	}

	// Git Mode
	var gitModeSource, gitModeOverrode string
	if flagSince != "" {
		gitModeSource = "cli-flag"
		if cfg != nil && cfg.Defaults.Git.Mode != "" {
			gitModeOverrode = cfg.Defaults.Git.Mode
		} else {
			gitModeOverrode = defaults.Defaults.Git.Mode
		}
	} else if cfg != nil && cfg.Defaults.Git.Mode != "" {
		gitModeSource = "config-file"
	} else {
		gitModeSource = "default"
	}
	addValue("defaults.git.mode", gitMode, gitModeSource, gitModeOverrode)

	// Git Ref
	var gitRefSource, gitRefOverrode string
	if flagSince != "" {
		gitRefSource = "cli-flag"
		if cfg != nil && cfg.Defaults.Git.Ref != "" {
			gitRefOverrode = cfg.Defaults.Git.Ref
		} else {
			gitRefOverrode = defaults.Defaults.Git.Ref
		}
	} else if cfg != nil && cfg.Defaults.Git.Ref != "" {
		gitRefSource = "config-file"
	} else {
		gitRefSource = "default"
	}
	addValue("defaults.git.ref", gitRef, gitRefSource, gitRefOverrode)

	// Task Defaults
	if cfg != nil && cfg.TaskDefaults.Enabled != nil {
		addValue("task_defaults.enabled", fmt.Sprintf("%t", *mergedCfg.TaskDefaults.Enabled), "config-file", "")
	} else {
		addValue("task_defaults.enabled", fmt.Sprintf("%t", *mergedCfg.TaskDefaults.Enabled), "default", "")
	}

	if cfg != nil && cfg.TaskDefaults.Workdir != "" {
		addValue("task_defaults.workdir", mergedCfg.TaskDefaults.Workdir, "config-file", "")
	} else {
		addValue("task_defaults.workdir", mergedCfg.TaskDefaults.Workdir, "default", "")
	}

	return &model.EffectiveConfig{
		Values: values,
	}
}

// Phase represents a group of tasks that can run in parallel
type Phase struct {
	Tasks []model.TaskDefinition
	Name  string // Display name for the phase
}

// groupTasksIntoPhases splits tasks into phases based on wait markers
func groupTasksIntoPhases(tasks []model.TaskDefinition, phaseNames map[string]config.PhaseInfo) []Phase {
	if len(tasks) == 0 {
		return nil
	}

	var phases []Phase
	currentPhase := Phase{Tasks: []model.TaskDefinition{}}
	phaseNum := 1

	for _, task := range tasks {
		currentPhase.Tasks = append(currentPhase.Tasks, task)

		// If this task has wait=true, end the current phase
		if task.Wait {
			// Set phase name from phaseNames map, or default to "Phase N"
			phaseKey := "wait-" + fmt.Sprintf("%d", phaseNum)
			if info, ok := phaseNames[phaseKey]; ok && info.Name != "" {
				currentPhase.Name = info.Name
			} else {
				currentPhase.Name = fmt.Sprintf("Phase %d", phaseNum)
			}

			phases = append(phases, currentPhase)
			currentPhase = Phase{Tasks: []model.TaskDefinition{}}
			phaseNum++
		}
	}

	// Add remaining tasks as final phase
	if len(currentPhase.Tasks) > 0 {
		// Set phase name for the last phase
		phaseKey := "wait-" + fmt.Sprintf("%d", phaseNum)
		if info, ok := phaseNames[phaseKey]; ok && info.Name != "" {
			currentPhase.Name = info.Name
		} else {
			currentPhase.Name = fmt.Sprintf("Phase %d", phaseNum)
		}
		phases = append(phases, currentPhase)
	}

	return phases
}

func filterTasks(tasks []model.TaskDefinition, only string, skip sliceFlag, _ bool, _ int, verbose bool) []model.TaskDefinition {
	skipSet := map[string]struct{}{}
	for _, id := range skip {
		skipSet[id] = struct{}{}
	}

	var out []model.TaskDefinition
	if only != "" {
		// Parse comma-separated list from --only
		rawIDs := strings.Split(only, ",")
		var requested []string
		seen := make(map[string]struct{})
		for _, raw := range rawIDs {
			id := strings.TrimSpace(raw)
			if id == "" {
				continue
			}
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			requested = append(requested, id)
		}

		// Index tasks by ID for validation
		taskIndex := make(map[string]model.TaskDefinition, len(tasks))
		for _, s := range tasks {
			taskIndex[s.ID] = s
		}

		// Validate all requested IDs exist
		for _, id := range requested {
			if _, ok := taskIndex[id]; !ok {
				fmt.Fprintf(os.Stderr, "ERROR: --only task id %q not found\n", id)
				os.Exit(1)
			}
		}

		// Build a set of requested IDs to filter by, then
		// walk the full task list in pipeline order and
		// keep only requested tasks that are not skipped.
		requestedSet := make(map[string]struct{}, len(requested))
		for _, id := range requested {
			requestedSet[id] = struct{}{}
		}

		for _, s := range tasks {
			if _, want := requestedSet[s.ID]; !want {
				continue
			}
			if _, skip := skipSet[s.ID]; skip {
				if verbose {
					fmt.Printf("[%-15s] SKIP requested by --skip\n", s.ID)
				}
				continue
			}
			out = append(out, s)
		}

		return out
	}

	for _, s := range tasks {
		if _, ok := skipSet[s.ID]; ok {
			if verbose {
				fmt.Printf("[%-15s] SKIP requested by --skip\n", s.ID)
			}
			continue
		}
		out = append(out, s)
	}
	return out
}

func filterTasksByWatchPaths(tasks []model.TaskDefinition, changedFiles []string, projectRoot string, verbose bool) []model.TaskDefinition {
	var out []model.TaskDefinition
	for _, task := range tasks {
		// If task has no watchPaths, always include it
		if len(task.WatchPaths) == 0 {
			out = append(out, task)
			continue
		}

		// If task has watchPaths but no changed files, skip it
		if len(changedFiles) == 0 {
			if verbose {
				fmt.Printf("[%-15s] SKIP (no changed files, has watchPaths)\n", task.ID)
			}
			continue
		}

		// Check if any changed file matches any watchPath pattern
		matched := false
		for _, changedFile := range changedFiles {
			// Make changed file path absolute
			absChangedFile := changedFile
			if !filepath.IsAbs(changedFile) {
				absChangedFile = filepath.Join(projectRoot, changedFile)
			}

			// Check against each watchPath pattern
			for _, pattern := range task.WatchPaths {
				// Make pattern absolute relative to task workdir
				absPattern := pattern
				if !filepath.IsAbs(pattern) {
					absPattern = filepath.Join(task.Workdir, pattern)
				}

				// Use doublestar for glob matching (supports **)
				match, err := doublestar.Match(absPattern, absChangedFile)
				if err != nil {
					// Invalid pattern, log and skip
					if verbose {
						fmt.Printf("[%-15s] WARNING: invalid watchPath pattern %q: %v\n", task.ID, pattern, err)
					}
					continue
				}

				if match {
					matched = true
					break
				}
			}

			if matched {
				break
			}
		}

		if matched {
			out = append(out, task)
		} else if verbose {
			fmt.Printf("[%-15s] SKIP (no matching changes for watchPaths)\n", task.ID)
		}
	}

	return out
}

// determineProjectRoot resolves the project root directory
// Priority: 1) config.projectRoot override, 2) git root from config location, 3) config directory
func determineProjectRoot(configPath string, cfg config.Config, gitRoot string, inGitRepo bool) string {
	// If projectRoot is explicitly set in config, use it
	if cfg.Defaults.ProjectRoot != "" {
		projectRoot := cfg.Defaults.ProjectRoot
		// If relative, resolve from config file location
		if !filepath.IsAbs(projectRoot) {
			if configPath != "" {
				absConfigPath, err := filepath.Abs(configPath)
				if err == nil {
					configDir := filepath.Dir(absConfigPath)
					projectRoot = filepath.Join(configDir, projectRoot)
				}
			}
		}
		return filepath.Clean(projectRoot)
	}

	// Auto-detect: if config path is provided, detect from config location
	if configPath != "" {
		absConfigPath, err := filepath.Abs(configPath)
		if err == nil {
			configDir := filepath.Dir(absConfigPath)
			// Try to find git root from config directory
			detectedRoot, detectedInRepo := git.DetectProjectRootFrom(configDir)
			if detectedInRepo {
				return detectedRoot
			}
			// No git repo, use config directory
			return configDir
		}
	}

	// Fallback: use git root if in repo, otherwise current directory
	if inGitRepo {
		return gitRoot
	}
	return gitRoot // gitRoot is already CWD if not in repo
}

func makeRunID() string {
	now := time.Now().UTC()
	ts := now.Format("2006-01-02T15-04-05Z")
	// Use PID for uniqueness instead of deprecated rand.Seed
	suffix := os.Getpid() % 1_000_000
	return fmt.Sprintf("%s_%06d", ts, suffix)
}

func runTask(st model.TaskDefinition, runDir, logDir string, dryRun bool, verbose bool, renderer *ui.Renderer, tracker *ui.AnimatedTaskTracker, outputMu *sync.Mutex, waitForPrev chan struct{}, taskDone chan struct{}) (model.TaskResult, *bytes.Buffer, error) {
	res := model.TaskResult{
		ID:               st.ID,
		Name:             st.Name,
		Desc:             st.Desc,
		Phase:            st.Phase,
		Type:             st.Type,
		Status:           model.StatusPending,
		Command:          st.Command,
		Workdir:          st.Workdir,
		LogPath:          "",
		EstimatedSeconds: st.EstimatedSeconds,
	}

	// Create a buffer to capture all output for this task
	var taskOutputBuffer bytes.Buffer

	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", st.ID))
	res.LogPath = logPath

	if dryRun {
		res.Status = model.StatusSkipped
		res.Skipped = true
		res.SkipReason = "dry-run"
		// Don't render anything yet - will be shown when displayed
		return res, &taskOutputBuffer, nil
	}

	// Wait for our turn to display output (non-animated mode only)
	if tracker == nil {
		// Wait for previous task to finish (if there is one)
		if waitForPrev != nil {
			<-waitForPrev
		}

		// Now we can stream output
		if verbose {
			fmt.Printf("[%-15s] %s    %s\n", st.ID, renderer.Blue("RUN"), st.Command)
		} else {
			fmt.Printf("[%-15s] %s\n", st.ID, renderer.Blue("RUN"))
		}
	} else {
		// Animated mode: buffer the RUN message with a blank line before it
		renderer.RenderTaskStart(st.ID, st.Command, verbose)
		taskOutputBuffer.WriteString("\n") // Blank line before task
		if verbose {
			taskOutputBuffer.WriteString(fmt.Sprintf("[%-15s] %s    %s\n", st.ID, renderer.Blue("RUN"), st.Command))
		} else {
			taskOutputBuffer.WriteString(fmt.Sprintf("[%-15s] %s\n", st.ID, renderer.Blue("RUN")))
		}
	}

	start := time.Now().UTC()
	res.StartTime = start.Format(time.RFC3339)
	res.Status = model.StatusRunning

	// Update tracker if animated
	if tracker != nil {
		tracker.UpdateTask(st.ID, "RUNNING", 0)
	}

	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot create log file %s: %v\n", logPath, err)
		res.Status = model.StatusFail
		return res, &taskOutputBuffer, err
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close log file: %v\n", err)
		}
	}()

	cmd := exec.Command("sh", "-c", st.Command)
	cmd.Dir = st.Workdir
	cmd.Env = append(os.Environ(), "FORCE_COLOR=1")

	// Setup output handling
	var bufferMu sync.Mutex

	if tracker != nil {
		// Animated mode: buffer output for sequential display
		stdoutWriter := &lineWriter{taskID: st.ID, file: logFile, outputBuffer: &taskOutputBuffer, mu: &bufferMu, renderer: renderer}
		stderrWriter := &lineWriter{taskID: st.ID, file: logFile, outputBuffer: &taskOutputBuffer, mu: &bufferMu, renderer: renderer}
		cmd.Stdout = stdoutWriter
		cmd.Stderr = stderrWriter
	} else {
		// Non-animated mode: stream output directly (we already have the turn)
		stdoutWriter := &lineWriter{taskID: st.ID, file: logFile, console: os.Stdout, renderer: renderer}
		stderrWriter := &lineWriter{taskID: st.ID, file: logFile, console: os.Stderr, renderer: renderer}
		cmd.Stdout = stdoutWriter
		cmd.Stderr = stderrWriter
	}

	// Start ticker to update progress during execution
	var tickerDone chan struct{}
	if tracker != nil {
		tickerDone = make(chan struct{})
		go func() {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			startTime := time.Now()

			for {
				select {
				case <-tickerDone:
					return
				case <-ticker.C:
					elapsed := time.Since(startTime).Seconds()
					tracker.UpdateTask(st.ID, "RUNNING", elapsed)
				}
			}
		}()
	}

	err = cmd.Run()

	// Stop ticker
	if tickerDone != nil {
		close(tickerDone)
	}

	end := time.Now().UTC()
	res.EndTime = end.Format(time.RFC3339)
	res.DurationMs = end.Sub(start).Milliseconds()
	elapsed := end.Sub(start).Seconds()

	exitCode := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			exitCode = 1
		}
		res.Status = model.StatusFail
		res.ExitCode = &exitCode

		// Parse metrics even on failure (especially useful for SARIF/JUnit)
		// This allows us to show what failed in the dashboard
		if st.MetricsFormat != "" && st.MetricsPath != "" {
			renderer.Verbose(verbose, "%s Task failed, but attempting to parse metrics: format=%s, path=%s", st.ID, st.MetricsFormat, st.MetricsPath)
			res.Metrics = parseTaskMetrics(st, verbose)
			if res.Metrics != nil {
				renderer.Verbose(verbose, "%s Metrics parsed successfully despite failure: %+v", st.ID, res.Metrics.Data)
			}
		}

		// Update tracker with final status
		if tracker != nil {
			tracker.UpdateTask(st.ID, "FAIL", elapsed)
			renderer.RenderTaskComplete(st.ID, string(res.Status), &exitCode, res.DurationMs, verbose)

			// Also buffer the failure message for the output section
			taskOutputBuffer.WriteString(fmt.Sprintf("[%-15s] ✗ %s (%dms)\n", st.ID, renderer.Red("FAIL"), res.DurationMs))
		} else {
			// Stream the failure message with color
			fmt.Printf("[%-15s] ✗ %s (%dms)\n\n", st.ID, renderer.Red("FAIL"), res.DurationMs)

			// Signal that this task is done streaming
			close(taskDone)
		}
		return res, &taskOutputBuffer, err
	}

	res.Status = model.StatusPass
	res.ExitCode = &exitCode

	// Parse metrics if configured
	if st.MetricsFormat != "" && st.MetricsPath != "" {
		renderer.Verbose(verbose, "%s Parsing metrics: format=%s, path=%s", st.ID, st.MetricsFormat, st.MetricsPath)
		res.Metrics = parseTaskMetrics(st, verbose)
		if res.Metrics != nil {
			renderer.Verbose(verbose, "%s Metrics parsed successfully: %+v", st.ID, res.Metrics.Data)
		}
		// Metrics parsing failed - this means either:
		// 1. File doesn't exist (will be caught below)
		// 2. Invalid format (error already printed)
		// 3. Parse error (error already printed)
		// We'll fail the task below if file is missing/empty, or here if it's a parse/format error

		// Validate artifact if metrics path specified
		// Handle both absolute and relative paths
		var artifactPath string
		if filepath.IsAbs(st.MetricsPath) {
			artifactPath = st.MetricsPath
		} else {
			artifactPath = filepath.Join(st.Workdir, st.MetricsPath)
		}
		if info, err := os.Stat(artifactPath); err != nil || info.Size() == 0 {
			// Artifact missing or empty - fail the task
			res.Status = model.StatusFail
			if err != nil {
				// Always show this error (not just in verbose)
				fmt.Fprintf(os.Stderr, "[%-15s] ❌ ERROR: Metrics file not found: %s\n", st.ID, st.MetricsPath)
				renderer.Verbose(verbose, "%s Full path: %s", st.ID, artifactPath)
			} else {
				// Always show this error (not just in verbose)
				fmt.Fprintf(os.Stderr, "[%-15s] ❌ ERROR: Metrics file is empty: %s\n", st.ID, st.MetricsPath)
				renderer.Verbose(verbose, "%s Full path: %s", st.ID, artifactPath)
			}
		} else if res.Metrics == nil {
			// File exists but metrics parsing failed (invalid format or parse error)
			res.Status = model.StatusFail
			// Error already printed by parseTaskMetrics
			renderer.Verbose(verbose, "%s Metrics validation FAILED: file exists but parsing failed", st.ID)
		} else {
			// Artifact exists, has size, and metrics parsed successfully
			res.Metrics.Data["path"] = artifactPath
			res.Metrics.Data["size"] = info.Size()

			renderer.Verbose(verbose, "%s Artifact validation PASSED: %s (%d bytes)", st.ID, artifactPath, info.Size())

			// Copy artifact to run directory for historical preservation
			artifactsDir := filepath.Join(runDir, "artifacts")
			if err := os.MkdirAll(artifactsDir, 0755); err != nil {
				renderer.Verbose(verbose, "%s Failed to create artifacts directory: %v", st.ID, err)
			} else {
				// Determine destination path based on whether source is absolute or relative
				var destPath string
				if filepath.IsAbs(st.MetricsPath) {
					// For absolute paths, store under task ID with full path to avoid conflicts
					// e.g., /foo/bar/file.xml -> artifacts/<task-id>/foo/bar/file.xml
					destPath = filepath.Join(artifactsDir, st.ID, st.MetricsPath)
				} else {
					// For relative paths, preserve directory structure
					destPath = filepath.Join(artifactsDir, st.MetricsPath)
				}
				destDir := filepath.Dir(destPath)
				if err := os.MkdirAll(destDir, 0755); err != nil {
					renderer.Verbose(verbose, "%s Failed to create artifact subdirectory: %v", st.ID, err)
				} else {
					// Copy the file
					if content, err := os.ReadFile(artifactPath); err != nil {
						renderer.Verbose(verbose, "%s Failed to read artifact for copying: %v", st.ID, err)
					} else if err := os.WriteFile(destPath, content, 0644); err != nil {
						renderer.Verbose(verbose, "%s Failed to copy artifact: %v", st.ID, err)
					} else {
						renderer.Verbose(verbose, "%s Artifact copied to: %s", st.ID, destPath)
					}
				}
			}
		}
	} else {
		renderer.Verbose(verbose, "%s No metrics configured (format=%s, path=%s)", st.ID, st.MetricsFormat, st.MetricsPath)
	}

	// Update tracker with final status
	if tracker != nil {
		tracker.UpdateTask(st.ID, string(res.Status), elapsed)
		renderer.RenderTaskComplete(st.ID, string(res.Status), &exitCode, res.DurationMs, verbose)

		// Also buffer the completion message for the output section
		symbol := "•"
		var statusText string

		switch res.Status {
		case model.StatusPass:
			symbol = "✓"
			statusText = renderer.Green(string(res.Status))
		case model.StatusFail:
			symbol = "✗"
			statusText = renderer.Red(string(res.Status))
		case model.StatusSkipped:
			symbol = "⊘"
			statusText = renderer.Yellow(string(res.Status))
		default:
			statusText = string(res.Status)
		}

		if verbose && exitCode != 0 {
			taskOutputBuffer.WriteString(fmt.Sprintf("[%-15s] %s %s (exit %d, %dms)\n", st.ID, symbol, statusText, exitCode, res.DurationMs))
		} else {
			taskOutputBuffer.WriteString(fmt.Sprintf("[%-15s] %s %s (%dms)\n", st.ID, symbol, statusText, res.DurationMs))
		}
	} else {
		// Stream the completion message for non-animated mode with colors
		symbol := "•"
		var statusText string

		switch res.Status {
		case model.StatusPass:
			symbol = "✓"
			statusText = renderer.Green(string(res.Status))
		case model.StatusFail:
			symbol = "✗"
			statusText = renderer.Red(string(res.Status))
		case model.StatusSkipped:
			symbol = "⊘"
			statusText = renderer.Yellow(string(res.Status))
		default:
			statusText = string(res.Status)
		}

		if verbose && exitCode != 0 {
			fmt.Printf("[%-15s] %s %s (exit %d, %dms)\n", st.ID, symbol, statusText, exitCode, res.DurationMs)
		} else {
			fmt.Printf("[%-15s] %s %s (%dms)\n", st.ID, symbol, statusText, res.DurationMs)
		}
		fmt.Println() // Blank line after task

		// Signal that this task is done streaming
		close(taskDone)
	}

	return res, &taskOutputBuffer, nil
}

func writeRunJSON(runDir string, record model.RunRecord) error {
	path := filepath.Join(runDir, "run.json")
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// parseTaskMetrics parses metrics for a completed task
func parseTaskMetrics(st model.TaskDefinition, verbose bool) *model.TaskMetrics {
	// Build full path to metrics file (handle both absolute and relative paths)
	var metricsPath string
	if filepath.IsAbs(st.MetricsPath) {
		metricsPath = st.MetricsPath
	} else {
		metricsPath = filepath.Join(st.Workdir, st.MetricsPath)
	}

	// Check if file exists
	if _, err := os.Stat(metricsPath); os.IsNotExist(err) {
		if verbose {
			fmt.Fprintf(os.Stderr, "[%-15s] Metrics file not found: %s\n", st.ID, metricsPath)
		}
		return nil
	}

	// Parse based on format
	switch st.MetricsFormat {
	case "junit":
		m, err := metrics.ParseJUnitXML(metricsPath)
		if err != nil {
			// Always show parse failures (not just in verbose mode)
			fmt.Fprintf(os.Stderr, "[%-15s] ❌ ERROR: Failed to parse JUnit XML: %v\n", st.ID, err)
			fmt.Fprintf(os.Stderr, "[%-15s]          File: %s\n", st.ID, st.MetricsPath)
			return nil
		}
		return m
	case "sarif":
		m, err := metrics.ParseSARIF(metricsPath)
		if err != nil {
			// Always show parse failures (not just in verbose mode)
			fmt.Fprintf(os.Stderr, "[%-15s] ❌ ERROR: Failed to parse SARIF: %v\n", st.ID, err)
			fmt.Fprintf(os.Stderr, "[%-15s]          File: %s\n", st.ID, st.MetricsPath)
			return nil
		}
		return m
	case "artifact":
		// For artifact format, just verify file exists and has content (already done above)
		return &model.TaskMetrics{
			Kind:          "artifact",
			SummaryFormat: "artifact",
			Data: map[string]interface{}{
				"path": metricsPath,
			},
		}
	default:
		// Unknown format - this is an error
		fmt.Fprintf(os.Stderr, "[%-15s] ❌ ERROR: Unknown metrics format: %s\n", st.ID, st.MetricsFormat)
		fmt.Fprintf(os.Stderr, "[%-15s]          Supported formats: junit, sarif, artifact\n", st.ID)
		return nil
	}
}

// copyConfigToRun copies the config file to the run directory
func copyConfigToRun(runDir, configPath string, mergedCfg *config.Config) error {
	destPath := filepath.Join(runDir, "config.toml")

	// If no config path specified, try default location
	if configPath == "" {
		configPath = "config.toml"
	}

	// If config file exists, copy it
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, 0644)
	}

	// Otherwise, write the merged config as JSON (built-in + defaults)
	data, err := json.MarshalIndent(mergedCfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(runDir, "config.json"), data, 0644)
}

// lineWriter captures output line by line and sends to tracker
type lineWriter struct {
	tracker      *ui.AnimatedTaskTracker
	taskID       string
	file         *os.File
	buffer       []byte
	outputBuffer *bytes.Buffer // Buffer all output until task completes
	mu           *sync.Mutex   // Protect outputBuffer
	console      *os.File      // For streaming output directly
	renderer     *ui.Renderer  // For colorizing output
}

func (w *lineWriter) Write(p []byte) (n int, err error) {
	// Write to log file (unprefixed)
	_, _ = w.file.Write(p) // Best effort log write

	// Add to buffer and extract complete lines
	w.buffer = append(w.buffer, p...)

	// Process complete lines
	for {
		idx := bytes.IndexByte(w.buffer, '\n')
		if idx == -1 {
			break
		}

		line := string(w.buffer[:idx])

		// Prefix line with task ID
		prefixedLine := fmt.Sprintf("[%-15s] %s", w.taskID, line)

		if w.tracker != nil {
			w.tracker.AddLogLine(prefixedLine)
		} else if w.outputBuffer != nil {
			// Buffer output for sequential display
			w.mu.Lock()
			w.outputBuffer.WriteString(prefixedLine)
			w.outputBuffer.WriteString("\n")
			w.mu.Unlock()
		} else if w.console != nil {
			// Stream output directly
			_, _ = fmt.Fprintln(w.console, prefixedLine) // Best effort console write
		}

		w.buffer = w.buffer[idx+1:]
	}

	return len(p), nil
}

// printVersion prints version information
func printVersion() {
	fmt.Printf("devpipe version %s\n", version)
	fmt.Printf("  commit: %s\n", commit)
	fmt.Printf("  built: %s\n", buildDate)
	fmt.Printf("  go: %s\n", runtime.Version())
	fmt.Printf("  platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// findSimilarCommand finds the most similar command using simple string matching
func findSimilarCommand(input string, commands []string) string {
	minDist := len(input)
	var bestMatch string

	for _, cmd := range commands {
		dist := levenshteinDistance(input, cmd)
		// Only suggest if distance is small (typo, not completely different)
		if dist < minDist && dist <= 2 {
			minDist = dist
			bestMatch = cmd
		}
	}

	return bestMatch
}

// levenshteinDistance calculates the edit distance between two strings
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// min returns the minimum of three integers
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// printHelp prints usage information
func printHelp() {
	fmt.Println("devpipe - Fast, local pipeline runner")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  devpipe [flags]              Run the pipeline")
	fmt.Println("  devpipe list [--verbose]     List all tasks")
	fmt.Println("  devpipe validate [files...]  Validate config file(s)")
	fmt.Println("  devpipe generate-reports     Regenerate all reports with latest template")
	fmt.Println("  devpipe sarif [options] ...  View SARIF security scan results")
	fmt.Println("  devpipe version              Show version information")
	fmt.Println("  devpipe help                 Show this help")
	fmt.Println()
	fmt.Println("RUN FLAGS:")
	fmt.Println("  --config <path>       Path to config file (default: config.toml)")
	fmt.Println("  --since <ref>         Git ref to compare against (overrides config)")
	fmt.Println("  --only <task-ids>     Run only specific task(s) by id (comma-separated)")
	fmt.Println("  --skip <task-id>      Skip a task by id (can be specified multiple times)")
	fmt.Println("  --ui <mode>           UI mode: basic, full (default: basic)")
	fmt.Println("  --dashboard           Show dashboard with live progress")
	fmt.Println("  --fail-fast           Stop on first task failure")
	fmt.Println("  --fast                Skip long running tasks")
	fmt.Println("  --ignore-watch-paths  Ignore watchPaths and run all tasks")
	fmt.Println("  --dry-run             Do not execute commands, simulate only")
	fmt.Println("  --verbose             Verbose logging")
	fmt.Println("  --no-color            Disable colored output")
	fmt.Println()
	fmt.Println("VALIDATE FLAGS:")
	fmt.Println("  --config <path>       Path to config file to validate (default: config.toml)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  devpipe                                    # Run pipeline with default config")
	fmt.Println("  devpipe --config config/custom.toml        # Run with custom config")
	fmt.Println("  devpipe --fast --fail-fast                 # Skip slow tasks, stop on failure")
	fmt.Println("  devpipe list                               # List all task IDs")
	fmt.Println("  devpipe list --verbose                     # List tasks in table format with details")
	fmt.Println("  devpipe validate                           # Validate default config.toml")
	fmt.Println("  devpipe validate config/*.toml             # Validate all configs in folder")
	fmt.Println("  devpipe generate-reports                   # Regenerate all reports with latest template")
	fmt.Println("  devpipe sarif tmp/codeql/results.sarif     # View CodeQL security scan results")
	fmt.Println("  devpipe sarif -s tmp/codeql/results.sarif  # Show summary of security issues")
	fmt.Println()
}

// validateCmd handles the validate subcommand
func validateCmd() {
	// Parse remaining args
	files := os.Args[2:]
	if len(files) == 0 {
		files = []string{"config.toml"}
	}

	hasErrors := false
	for _, file := range files {
		result, err := config.ValidateConfigFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ ERROR: %v\n", err)
			hasErrors = true
			continue
		}

		config.PrintValidationResult(file, result)
		if !result.Valid {
			hasErrors = true
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}

// generateReportsCmd handles the generate-reports subcommand
func generateReportsCmd() {
	startTime := time.Now()
	fmt.Println("Regenerating all reports with latest template...")

	// Determine project root
	projectRoot, _ := git.DetectProjectRoot()

	// Get output root from default config
	cfg, _, _, _, err := config.LoadConfig("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to load config: %v\n", err)
		os.Exit(1)
	}

	mergedCfg := config.MergeWithDefaults(cfg)
	outputRoot := filepath.Join(projectRoot, mergedCfg.Defaults.OutputRoot)

	// Count runs before regenerating
	runsDir := filepath.Join(outputRoot, "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to read runs directory: %v\n", err)
		os.Exit(1)
	}
	numRuns := len(entries)

	// Regenerate all reports
	if err := dashboard.GenerateDashboardWithOptions(outputRoot, version, true, ""); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to regenerate reports: %v\n", err)
		os.Exit(1)
	}

	duration := time.Since(startTime)
	fmt.Printf("✓ Regenerated %d reports in %s\n", numRuns, duration.Round(time.Millisecond))
	fmt.Printf("📊 Dashboard: %s\n", filepath.Join(outputRoot, "report.html"))
}

// getTerminalWidth returns the current terminal width, defaulting to 160 if unable to detect
func getTerminalWidth() int {
	// Try to get terminal width using stty
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err == nil {
		var rows, cols int
		if _, err := fmt.Sscanf(string(out), "%d %d", &rows, &cols); err == nil && cols > 0 {
			return cols
		}
	}

	// Default to 160 columns if we can't detect
	return 160
}

// wrapText wraps text to specified width, breaking on word boundaries
func wrapText(text string, width int) []string {
	if text == "" {
		return []string{}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// truncate truncates a string to the specified length with "..." if needed
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// phaseEmoji returns an emoji for a phase based on its name
func phaseEmoji(phaseName string) string {
	// Normalize to lowercase for matching
	name := strings.ToLower(phaseName)

	// Map phase names/keywords to emojis
	emojiMap := map[string]string{
		"validation":  "🧪",  // Test tube for validation/testing
		"test":        "🧪",  // Test tube
		"testing":     "🧪",  // Test tube
		"build":       "📦",  // Package for build
		"package":     "📦",  // Package
		"compile":     "🔨",  // Hammer for compilation
		"deploy":      "🚀",  // Rocket for deployment
		"release":     "🚀",  // Rocket for release
		"lint":        "🔍",  // Magnifying glass for linting
		"security":    "🔒",  // Lock for security
		"e2e":         "🎯",  // Target for end-to-end tests
		"end-to-end":  "🎯",  // Target
		"integration": "🔗",  // Link for integration
		"setup":       "⚙️", // Gear for setup
		"cleanup":     "🧹",  // Broom for cleanup
		"docs":        "📚",  // Books for documentation
		"publish":     "📤",  // Outbox for publishing
	}

	// Check for exact match first
	if emoji, ok := emojiMap[name]; ok {
		return emoji
	}

	// Check if any keyword is contained in the phase name
	for keyword, emoji := range emojiMap {
		if strings.Contains(name, keyword) {
			return emoji
		}
	}

	// Default emoji
	return "📋" // Clipboard as default
}

// loadTaskAveragesLast25 loads task average durations from last 25 runs
func loadTaskAveragesLast25(outputRoot string) map[string]float64 {
	averages := make(map[string]float64)

	summaryPath := filepath.Join(outputRoot, "summary.json")
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		return averages // No history yet
	}

	var summary struct {
		TaskStatsLast25 map[string]struct {
			AvgDuration float64 `json:"avgDuration"`
		} `json:"taskStatsLast25"`
	}

	if err := json.Unmarshal(data, &summary); err != nil {
		return averages
	}

	for taskID, stats := range summary.TaskStatsLast25 {
		if stats.AvgDuration > 0 {
			averages[taskID] = stats.AvgDuration
		}
	}

	return averages
}

// listCmd handles the list subcommand
func listCmd() {
	// Parse flags
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	verbose := fs.Bool("verbose", false, "Show detailed table view with phases")
	configPath := fs.String("config", "", "Path to config file (default: config.toml)")
	_ = fs.Parse(os.Args[2:]) // Flag parsing

	// Load configuration
	cfg, configTaskOrder, phaseNames, taskToPhase, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

	if cfg == nil {
		fmt.Fprintf(os.Stderr, "ERROR: No config.toml found. Create one or specify with --config flag\n")
		os.Exit(1)
	}

	// Merge with defaults
	mergedCfg := config.MergeWithDefaults(cfg)

	// Determine project root
	projectRoot, _ := git.DetectProjectRoot()

	// Load historical averages
	outputRoot := filepath.Join(projectRoot, mergedCfg.Defaults.OutputRoot)
	taskAverages := loadTaskAveragesLast25(outputRoot)

	// Build task list (filter out phase markers)
	var tasks []struct {
		id    string
		task  config.TaskConfig
		phase string
	}

	for _, id := range configTaskOrder {
		// Skip wait markers and phase headers
		if id == "wait" || strings.HasPrefix(id, "wait-") || strings.HasPrefix(id, "phase-") {
			continue
		}

		taskCfg, ok := mergedCfg.Tasks[id]
		if !ok {
			continue
		}

		// Get phase name
		phaseName := ""
		if phaseID, ok := taskToPhase[id]; ok {
			for _, phaseInfo := range phaseNames {
				if phaseInfo.ID == phaseID {
					phaseName = phaseInfo.Name
					break
				}
			}
		}

		tasks = append(tasks, struct {
			id    string
			task  config.TaskConfig
			phase string
		}{id, taskCfg, phaseName})
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found in config")
		return
	}

	// Simple mode: just list task IDs
	if !*verbose {
		for _, t := range tasks {
			fmt.Println(t.id)
		}
		return
	}

	// Verbose mode: table view grouped by phase
	fmt.Println("Tasks:")
	fmt.Println()

	// Group tasks by phase
	type phaseGroup struct {
		name  string
		tasks []struct {
			id   string
			task config.TaskConfig
		}
	}

	var phases []phaseGroup
	currentPhase := phaseGroup{name: "(no phase)"}

	for _, t := range tasks {
		if t.phase != "" && (len(currentPhase.tasks) == 0 || currentPhase.name != t.phase) {
			// Start new phase
			if len(currentPhase.tasks) > 0 {
				phases = append(phases, currentPhase)
			}
			currentPhase = phaseGroup{name: t.phase}
		}
		currentPhase.tasks = append(currentPhase.tasks, struct {
			id   string
			task config.TaskConfig
		}{t.id, t.task})
	}
	if len(currentPhase.tasks) > 0 {
		phases = append(phases, currentPhase)
	}

	// Get terminal width
	termWidth := getTerminalWidth()

	// Column widths (fixed for name, desc, type, duration, command)
	nameWidth := 40
	durationWidth := 13 // "⚠️  ~2ms" format (emoji takes more space)
	typeWidth := 18
	cmdWidth := 45
	spacing := 10 // 2 spaces between each column (5 gaps)
	descWidth := termWidth - nameWidth - durationWidth - typeWidth - cmdWidth - spacing
	if descWidth < 20 {
		descWidth = 20 // Minimum description width
	}

	// Print each phase
	for _, phase := range phases {
		// Calculate phase average duration
		var phaseAvgMs float64
		var phaseTaskCount int
		for _, t := range phase.tasks {
			if avg, ok := taskAverages[t.id]; ok {
				phaseAvgMs += avg
				phaseTaskCount++
			}
		}

		// Phase header with emoji and duration in a box
		emoji := phaseEmoji(phase.name)
		var phaseText string
		var durationTextPlain string
		if phaseTaskCount > 0 {
			phaseAvgSec := phaseAvgMs / 1000
			durationTextPlain = fmt.Sprintf(" (~%.1fs)", phaseAvgSec)

			// Gray color for phase duration (grouping, not individual timing)
			grayDuration := fmt.Sprintf("\033[90m(~%.1fs)\033[0m", phaseAvgSec)
			phaseText = fmt.Sprintf("%s %s %s", emoji, phase.name, grayDuration)
		} else {
			phaseText = fmt.Sprintf("%s %s", emoji, phase.name)
		}
		// Calculate visual width: emoji (2) + space (1) + name + duration text + padding (2)
		visualWidth := 2 + 1 + len(phase.name) + len(durationTextPlain) + 2

		// Top border
		fmt.Println("┌" + strings.Repeat("─", visualWidth) + "┐")
		// Header with bold
		fmt.Printf("│\033[1m %s \033[0m│\n", phaseText)
		// Bottom border
		fmt.Println("└" + strings.Repeat("─", visualWidth) + "┘")
		fmt.Println()

		// Table header (AVG is right-aligned)
		fmt.Printf("%-*s  %-*s  %-*s  %-*s  %*s\n", nameWidth, "NAME", descWidth, "DESCRIPTION", typeWidth, "TYPE", cmdWidth, "COMMAND", durationWidth, "AVG")
		fmt.Printf("%s  %s  %s  %s  %s\n", strings.Repeat("─", nameWidth), strings.Repeat("─", descWidth), strings.Repeat("─", typeWidth), strings.Repeat("─", cmdWidth), strings.Repeat("─", durationWidth))

		// Tasks
		for _, t := range phase.tasks {
			resolvedTask := mergedCfg.ResolveTaskConfig(t.id, t.task, projectRoot)

			// Add metrics emoji if present
			metricsEmoji := ""
			emojiDisplayWidth := 0
			if resolvedTask.MetricsFormat != "" {
				switch resolvedTask.MetricsFormat {
				case "junit":
					metricsEmoji = " 🧪"
					emojiDisplayWidth = 3 // space + emoji (2 display chars)
				case "sarif":
					metricsEmoji = " 🔒"
					emojiDisplayWidth = 3
				case "artifact":
					metricsEmoji = " 📦"
					emojiDisplayWidth = 3
				}
			}

			// Calculate display widths
			baseName := resolvedTask.Name
			idText := t.id
			// Total display width: name + emoji(if any) + space + id
			totalDisplayWidth := len(baseName) + emojiDisplayWidth + 1 + len(idText)

			// Truncate if needed
			if totalDisplayWidth > nameWidth {
				// Truncate the base name to fit
				availableForName := nameWidth - emojiDisplayWidth - 1 - len(idText) - 3 // -3 for "..."
				if availableForName > 0 {
					baseName = baseName[:availableForName] + "..."
					totalDisplayWidth = nameWidth
				} else {
					// Even the ID is too long, truncate everything
					baseName = truncate(baseName+metricsEmoji+" "+idText, nameWidth)
					metricsEmoji = ""
					idText = ""
					totalDisplayWidth = nameWidth
				}
			}

			// Format with colors: name + emoji, then gray ID
			var nameFormatted string
			if idText != "" {
				nameFormatted = fmt.Sprintf("%s%s \033[90m%s\033[0m", baseName, metricsEmoji, idText)
			} else {
				nameFormatted = baseName
			}

			// Calculate padding needed
			padding := nameWidth - totalDisplayWidth
			if padding < 0 {
				padding = 0
			}

			// Format type
			taskType := resolvedTask.Type
			if taskType == "" {
				taskType = "-"
			}

			// Truncate command if too long
			cmd := resolvedTask.Command
			if len(cmd) > cmdWidth {
				cmd = cmd[:cmdWidth-3] + "..."
			}

			// Truncate description to fit in one line
			desc := resolvedTask.Desc
			if len(desc) > descWidth {
				desc = desc[:descWidth-3] + "..."
			}

			// Format duration with color coding
			var durationStr string
			var visualLen int
			if avgMs, ok := taskAverages[t.id]; ok && avgMs > 0 {
				avgSec := avgMs / 1000
				var timeStr string
				if avgSec < 1 {
					// Force 1 decimal place for milliseconds too
					timeStr = fmt.Sprintf("~%.1fms", avgMs)
				} else if avgSec < 60 {
					timeStr = fmt.Sprintf("~%.1fs", avgSec)
				} else {
					minutes := avgSec / 60
					timeStr = fmt.Sprintf("~%.1fm", minutes)
				}

				// Special case: ≤3ms likely means echo/mock, not real timing
				if avgMs <= 3 {
					durationStr = fmt.Sprintf("\033[31m[!] %s\033[0m", timeStr) // Red with warning (using [!] instead of emoji)
					visualLen = 4 + len(timeStr)                                // [!] (3) + space (1) + time
					// Green shades (fast - instant feedback)
				} else if avgMs < 100 {
					durationStr = fmt.Sprintf("\033[38;5;46m%s\033[0m", timeStr) // Bright green
					visualLen = len(timeStr)
				} else if avgMs < 500 {
					durationStr = fmt.Sprintf("\033[38;5;40m%s\033[0m", timeStr) // Medium green
					visualLen = len(timeStr)
				} else if avgSec < 1 {
					durationStr = fmt.Sprintf("\033[38;5;34m%s\033[0m", timeStr) // Darker green
					visualLen = len(timeStr)
					// Yellow shades (moderate - acceptable)
				} else if avgSec < 5 {
					durationStr = fmt.Sprintf("\033[38;5;226m%s\033[0m", timeStr) // Bright yellow
					visualLen = len(timeStr)
				} else if avgSec < 10 {
					durationStr = fmt.Sprintf("\033[38;5;220m%s\033[0m", timeStr) // Medium yellow
					visualLen = len(timeStr)
				} else if avgSec < 20 {
					durationStr = fmt.Sprintf("\033[38;5;214m%s\033[0m", timeStr) // Darker yellow/amber
					visualLen = len(timeStr)
					// Orange shades (slow - noticeable)
				} else if avgSec < 30 {
					durationStr = fmt.Sprintf("\033[38;5;208m%s\033[0m", timeStr) // Light orange
					visualLen = len(timeStr)
				} else if avgSec < 45 {
					durationStr = fmt.Sprintf("\033[38;5;202m%s\033[0m", timeStr) // Medium orange
					visualLen = len(timeStr)
				} else if avgSec < 60 {
					durationStr = fmt.Sprintf("\033[38;5;196m%s\033[0m", timeStr) // Dark orange
					visualLen = len(timeStr)
					// Red shades (very slow - needs attention)
				} else if avgSec < 180 { // < 3 min
					durationStr = fmt.Sprintf("\033[38;5;160m%s\033[0m", timeStr) // Light red
					visualLen = len(timeStr)
				} else if avgSec < 600 { // < 10 min
					durationStr = fmt.Sprintf("\033[38;5;124m%s\033[0m", timeStr) // Medium red
					visualLen = len(timeStr)
				} else {
					durationStr = fmt.Sprintf("\033[38;5;88m%s\033[0m", timeStr) // Dark red
					visualLen = len(timeStr)
				}
			} else {
				durationStr = "-"
				visualLen = 1
			}

			// Right-align the duration by padding on the left
			leftPadding := durationWidth - visualLen
			if leftPadding < 0 {
				leftPadding = 0
			}

			// Print row with proper padding (name, desc, type, command, right-aligned avg)
			fmt.Printf("%s%s  %-*s  %-*s  %-*s  %s%s\n", nameFormatted, strings.Repeat(" ", padding), descWidth, desc, typeWidth, taskType, cmdWidth, cmd, strings.Repeat(" ", leftPadding), durationStr)
		}
		fmt.Println()
	}

	// Calculate total average
	var totalAvgMs float64
	var totalTasksWithAvg int
	for _, t := range tasks {
		if avg, ok := taskAverages[t.id]; ok {
			totalAvgMs += avg
			totalTasksWithAvg++
		}
	}

	if totalTasksWithAvg > 0 {
		totalAvgSec := totalAvgMs / 1000
		// Gray color for total (summary, not individual timing)
		fmt.Printf("Total: %d tasks \033[90m(~%.1fs avg)\033[0m\n", len(tasks), totalAvgSec)
	} else {
		fmt.Printf("Total: %d tasks\n", len(tasks))
	}
}

// sarifCmd handles the sarif subcommand
func sarifCmd() {
	// Define flags for sarif subcommand
	fs := flag.NewFlagSet("sarif", flag.ExitOnError)
	verbose := fs.Bool("v", false, "Verbose output (show severity, tags, precision, descriptions)")
	summary := fs.Bool("s", false, "Show summary grouped by rule")
	dir := fs.String("d", "", "Directory to search for SARIF files")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s sarif [options] <sarif-file> [<sarif-file>...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "   or: %s sarif -d <directory>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "View SARIF (Static Analysis Results Interchange Format) files in human-readable format.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nIf no file is specified, looks for tmp/codeql/results.sarif\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s sarif tmp/codeql/results.sarif           # Default format\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s sarif -v tmp/codeql/results.sarif        # Verbose with metadata\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s sarif -s tmp/codeql/results.sarif        # Summary by rule\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s sarif -d tmp/                            # Scan directory\n", os.Args[0])
	}

	// Parse flags (skip "sarif" subcommand)
	_ = fs.Parse(os.Args[2:]) // Flag parsing

	// Get SARIF file(s)
	var files []string
	if *dir != "" {
		// Search directory for SARIF files
		found, err := sarif.FindSARIFFiles(*dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching directory: %v\n", err)
			os.Exit(1)
		}
		files = found
	} else if fs.NArg() > 0 {
		// Use files from arguments
		files = fs.Args()
	} else {
		// Default: look for SARIF files in tmp/codeql
		defaultPath := "tmp/codeql/results.sarif"
		if _, err := os.Stat(defaultPath); err == nil {
			files = []string{defaultPath}
		} else {
			fs.Usage()
			os.Exit(1)
		}
	}

	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "No SARIF files found\n")
		os.Exit(1)
	}

	// Process all files
	var allFindings []sarif.Finding
	var parseErrors bool
	for _, file := range files {
		// Parse SARIF file
		doc, err := sarif.Parse(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", file, err)
			parseErrors = true
			continue
		}

		findings := doc.GetFindings()

		// If multiple files, show which file we're processing
		if len(files) > 1 && len(findings) > 0 {
			fmt.Printf("\n📄 %s:\n", filepath.Base(file))
		}

		allFindings = append(allFindings, findings...)
	}

	// Display results
	if *summary {
		sarif.PrintSummary(allFindings)
	} else {
		sarif.PrintFindings(allFindings, *verbose)
	}

	// Exit with error code only if parse errors occurred
	if parseErrors {
		os.Exit(1)
	}
}
