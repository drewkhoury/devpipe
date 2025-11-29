// devpipe - Iteration 2
//
// Local pipeline runner with TOML config support and git modes.

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
	"strings"
	"sync"
	"time"

	"github.com/drew/devpipe/internal/config"
	"github.com/drew/devpipe/internal/dashboard"
	"github.com/drew/devpipe/internal/git"
	"github.com/drew/devpipe/internal/metrics"
	"github.com/drew/devpipe/internal/model"
	"github.com/drew/devpipe/internal/ui"
	"golang.org/x/sync/errgroup"
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
	// CLI flags
	var (
		flagConfig   string
		flagSince    string
		flagOnly     string
		flagUI       string
		flagNoColor  bool
		flagAnimated bool
		flagFailFast bool
		flagDryRun   bool
		flagVerbose  bool
		flagFast     bool
		flagSkipVals sliceFlag
	)
	
	flag.StringVar(&flagConfig, "config", "", "Path to config file (default: config.toml)")
	flag.StringVar(&flagSince, "since", "", "Git ref to compare against (overrides config)")
	flag.StringVar(&flagOnly, "only", "", "Run only a single task by id")
	flag.StringVar(&flagUI, "ui", "basic", "UI mode: basic, full")
	flag.BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	flag.BoolVar(&flagAnimated, "animated", false, "Show live progress animation (experimental)")
	flag.Var(&flagSkipVals, "skip", "Skip a task by id (can be specified multiple times)")
	flag.BoolVar(&flagFailFast, "fail-fast", false, "Stop on first task failure")
	flag.BoolVar(&flagDryRun, "dry-run", false, "Do not execute commands, simulate only")
	flag.BoolVar(&flagVerbose, "verbose", false, "Verbose logging")
	flag.BoolVar(&flagFast, "fast", false, "Skip long running tasks")
	flag.Parse()
	
	// Load configuration first to get UI mode
	cfg, err := config.LoadConfig(flagConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	
	// Merge with defaults
	mergedCfg := config.MergeWithDefaults(cfg)
	
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
	// Animated mode requires TTY
	animated := flagAnimated && ui.IsTTY(uintptr(1))
	renderer := ui.NewRenderer(uiMode, enableColors, animated)

	// Determine repo root
	repoRoot, inGitRepo := git.DetectRepoRoot()
	if !inGitRepo && flagVerbose {
		fmt.Println("WARNING: not in a git repo, using current directory as repo root")
	}
	
	// Auto-generate config.toml if it doesn't exist and no custom config specified
	if cfg == nil && flagConfig == "" {
		defaultConfigPath := "config.toml"
		if err := config.GenerateDefaultConfig(defaultConfigPath, repoRoot); err != nil {
			if flagVerbose {
				fmt.Fprintf(os.Stderr, "WARNING: Could not generate config.toml: %v\n", err)
			}
		} else {
			if flagVerbose {
				fmt.Printf("Generated config.toml with built-in tasks\n")
			}
			// Reload config after generating
			cfg, _ = config.LoadConfig(defaultConfigPath)
			mergedCfg = config.MergeWithDefaults(cfg)
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
		tasks = config.BuiltInTasks(repoRoot)
		taskOrder = config.GetTaskOrder()
	} else {
		// Use tasks from config
		tasks = mergedCfg.Tasks
		// Use the built-in order if tasks match, otherwise alphabetical
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

	// Determine git mode and ref
	gitMode := mergedCfg.Defaults.Git.Mode
	gitRef := mergedCfg.Defaults.Git.Ref
	
	// CLI --since overrides config
	if flagSince != "" {
		gitMode = "ref"
		gitRef = flagSince
	}

	// Get changed files
	gitInfo := git.DetectChangedFiles(repoRoot, inGitRepo, gitMode, gitRef, flagVerbose)

	// Prepare output dir
	outputRoot := filepath.Join(repoRoot, mergedCfg.Defaults.OutputRoot)
	runID := makeRunID()
	runDir := filepath.Join(outputRoot, "runs", runID)
	logDir := filepath.Join(runDir, "logs")

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to create run directories: %v\n", err)
		os.Exit(1)
	}

	// Load historical averages
	historicalAvg := loadHistoricalAverages(outputRoot)
	
	// Build task definitions
	var taskDefs []model.TaskDefinition
	for _, id := range taskOrder {
		taskCfg, ok := tasks[id]
		if !ok {
			continue
		}
		
		// Resolve with defaults
		resolved := mergedCfg.ResolveTaskConfig(id, taskCfg, repoRoot)
		
		// Skip if disabled
		if resolved.Enabled != nil && !*resolved.Enabled {
			if flagVerbose {
				fmt.Printf("[%-15s] DISABLED in config\n", id)
			}
			continue
		}
		
		// Use historical average if available, otherwise use configured value
		estimatedSeconds := resolved.EstimatedSeconds
		isGuess := false
		
		if avgSeconds, hasHistory := historicalAvg[id]; hasHistory {
			// Always prefer historical average
			estimatedSeconds = avgSeconds
			isGuess = false // Historical data, not a guess
		} else if resolved.EstimatedSeconds == mergedCfg.TaskDefaults.EstimatedSeconds {
			// Using default value - mark as guess
			isGuess = true
		}
		
		taskDef := model.TaskDefinition{
			ID:               id,
			Name:             resolved.Name,
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
			if flagVerbose {
				fmt.Printf("[%-15s] Metrics configured: format=%s, path=%s\n", id, resolved.MetricsFormat, resolved.MetricsPath)
			}
		}
		
		taskDefs = append(taskDefs, taskDef)
	}

	// Apply CLI filters
	filteredTasks := filterTasks(taskDefs, flagOnly, flagSkipVals, flagFast, mergedCfg.Defaults.FastThreshold, flagVerbose)

	// Run tasks
	var (
		results         []model.TaskResult
		overallExitCode int
		anyFailed       bool
	)

	// Render header
	renderer.RenderHeader(runID, repoRoot, gitMode, len(gitInfo.ChangedFiles))
	
	// Setup animation if enabled
	var tracker *ui.AnimatedTaskTracker
	
	// Ensure cursor is restored on exit (in case of panic or early exit)
	defer func() {
		if tracker != nil {
			fmt.Print("\033[?25h") // Show cursor
			fmt.Println() // Add newline to ensure clean exit
		}
	}()
	
	if renderer.IsAnimated() {
		// Build task progress list
		var taskProgress []ui.TaskProgress
		for _, st := range filteredTasks {
			taskProgress = append(taskProgress, ui.TaskProgress{
				ID:               st.ID,
				Name:             st.Name,
				Type:             st.Type,
				Status:           "PENDING",
				EstimatedSeconds: st.EstimatedSeconds,
				IsEstimateGuess:  st.IsEstimateGuess,
				ElapsedSeconds:   0,
				StartTime:        time.Time{},
			})
		}
		
		// Calculate header lines
		headerLines := 4 // basic mode: run ID, repo, git mode, changed files
		if renderer.IsAnimated() {
			headerLines = 4
		}
		
		tracker = renderer.CreateAnimatedTracker(taskProgress, headerLines, mergedCfg.Defaults.AnimationRefreshMs)
		if tracker != nil {
			if err := tracker.Start(); err != nil {
				// Animation failed, fall back to non-animated
				if flagVerbose {
					fmt.Fprintf(os.Stderr, "WARNING: Animation not supported, using basic mode\n")
				}
				tracker = nil
			}
		}
	}

	// Group tasks into phases based on wait markers
	phases := groupTasksIntoPhases(filteredTasks)
	
	if flagVerbose && len(phases) > 1 {
		fmt.Printf("Executing %d phases with parallel tasks\n", len(phases))
	}
	
	// Execute phases sequentially, tasks within each phase in parallel
	var resultsMu sync.Mutex
	var outputMu sync.Mutex // For sequential output display
	
	for phaseIdx, phase := range phases {
		if flagVerbose && len(phases) > 1 {
			fmt.Printf("\n=== Phase %d/%d (%d tasks) ===\n", phaseIdx+1, len(phases), len(phase.Tasks))
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
		
		// If phase failed and fail-fast is enabled, stop
		phaseFailMu.Lock()
		shouldStop := phaseFailed && flagFailFast
		phaseFailMu.Unlock()
		
		if shouldStop {
			break
		}
	}

	// Stop animation if it was running
	if tracker != nil {
		tracker.Stop()
	}

	// Render summary
	var summaries []ui.TaskSummary
	for _, r := range results {
		summaries = append(summaries, ui.TaskSummary{
			ID:         r.ID,
			Status:     string(r.Status),
			DurationMs: r.DurationMs,
		})
	}
	renderer.RenderSummary(summaries, anyFailed)
	
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
		RepoRoot:        repoRoot,
		OutputRoot:      outputRoot,
		ConfigPath:      actualConfigPath,
		Command:         buildCommandString(),
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
	
	// Generate dashboard
	if err := dashboard.GenerateDashboard(outputRoot); err != nil {
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
	// Get username and hostname
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME") // Windows fallback
	}
	hostname, _ := os.Hostname()
	
	// Get current working directory
	cwd, _ := os.Getwd()
	
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
	
	// Format like: drew@drews-MBP devpipe % ./devpipe --config ...
	return fmt.Sprintf("%s@%s %s %% %s", username, hostname, filepath.Base(cwd), cmdLine)
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
func buildEffectiveConfig(cfg *config.Config, mergedCfg *config.Config, flagSince, flagUI, uiModeStr, gitMode, gitRef string, historicalAvg map[string]int) *model.EffectiveConfig {
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
	
	if cfg != nil && cfg.TaskDefaults.EstimatedSeconds != 0 {
		addValue("task_defaults.estimatedSeconds", fmt.Sprintf("%d", mergedCfg.TaskDefaults.EstimatedSeconds), "config-file", "")
	} else {
		addValue("task_defaults.estimatedSeconds", fmt.Sprintf("%d", mergedCfg.TaskDefaults.EstimatedSeconds), "default", "")
	}
	
	// Task-specific overrides (show if historical data was used)
	if len(historicalAvg) > 0 {
		for taskID, avgSeconds := range historicalAvg {
			var configSeconds int
			if cfg != nil {
				if taskCfg, ok := cfg.Tasks[taskID]; ok && taskCfg.EstimatedSeconds != 0 {
					configSeconds = taskCfg.EstimatedSeconds
				} else {
					configSeconds = mergedCfg.TaskDefaults.EstimatedSeconds
				}
			} else {
				configSeconds = defaults.TaskDefaults.EstimatedSeconds
			}
			
			if avgSeconds != configSeconds {
				addValue(
					fmt.Sprintf("tasks.%s.estimatedSeconds", taskID),
					fmt.Sprintf("%d", avgSeconds),
					"historical",
					fmt.Sprintf("%d", configSeconds),
				)
			}
		}
	}
	
	return &model.EffectiveConfig{
		Values: values,
	}
}

// Phase represents a group of tasks that can run in parallel
type Phase struct {
	Tasks []model.TaskDefinition
}

// groupTasksIntoPhases splits tasks into phases based on wait markers
func groupTasksIntoPhases(tasks []model.TaskDefinition) []Phase {
	if len(tasks) == 0 {
		return nil
	}
	
	var phases []Phase
	currentPhase := Phase{Tasks: []model.TaskDefinition{}}
	
	for _, task := range tasks {
		currentPhase.Tasks = append(currentPhase.Tasks, task)
		
		// If this task has wait=true, end the current phase
		if task.Wait {
			phases = append(phases, currentPhase)
			currentPhase = Phase{Tasks: []model.TaskDefinition{}}
		}
	}
	
	// Add remaining tasks as final phase
	if len(currentPhase.Tasks) > 0 {
		phases = append(phases, currentPhase)
	}
	
	return phases
}

func filterTasks(tasks []model.TaskDefinition, only string, skip sliceFlag, fast bool, fastThreshold int, verbose bool) []model.TaskDefinition {
	skipSet := map[string]struct{}{}
	for _, id := range skip {
		skipSet[id] = struct{}{}
	}

	var out []model.TaskDefinition
	if only != "" {
		for _, s := range tasks {
			if s.ID == only {
				out = append(out, s)
				return out
			}
		}
		fmt.Fprintf(os.Stderr, "ERROR: --only task id %q not found\n", only)
		os.Exit(1)
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
	defer logFile.Close()

	cmd := exec.Command("sh", "-c", st.Command)
	cmd.Dir = st.Workdir

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
		if verbose {
			fmt.Printf("[%-15s] Parsing metrics: format=%s, path=%s\n", st.ID, st.MetricsFormat, st.MetricsPath)
		}
		res.Metrics = parseTaskMetrics(st, verbose)
		if verbose && res.Metrics != nil {
			fmt.Printf("[%-15s] Metrics parsed successfully: %+v\n", st.ID, res.Metrics.Data)
		}
		
		// Validate artifact if metrics path specified
		artifactPath := filepath.Join(st.Workdir, st.MetricsPath)
		if info, err := os.Stat(artifactPath); err != nil || info.Size() == 0 {
			// Artifact missing or empty - fail the task
			res.Status = model.StatusFail
			if verbose {
				if err != nil {
					fmt.Printf("[%-15s] Artifact validation FAILED: file not found: %s\n", st.ID, artifactPath)
				} else {
					fmt.Printf("[%-15s] Artifact validation FAILED: file is empty: %s\n", st.ID, artifactPath)
				}
			}
		} else {
			// Artifact exists and has size - store info in metrics
			if res.Metrics == nil {
				res.Metrics = &model.TaskMetrics{
					Kind:          "artifact",
					SummaryFormat: "artifact",
					Data:          make(map[string]interface{}),
				}
			}
			res.Metrics.Data["path"] = artifactPath
			res.Metrics.Data["size"] = info.Size()
			
			if verbose {
				fmt.Printf("[%-15s] Artifact validation PASSED: %s (%d bytes)\n", st.ID, artifactPath, info.Size())
			}
		}
	} else if verbose {
		fmt.Printf("[%-15s] No metrics configured (format=%s, path=%s)\n", st.ID, st.MetricsFormat, st.MetricsPath)
	}
	
	// Update tracker with final status
	if tracker != nil {
		tracker.UpdateTask(st.ID, string(res.Status), elapsed)
		renderer.RenderTaskComplete(st.ID, string(res.Status), &exitCode, res.DurationMs, verbose)
		
		// Also buffer the completion message for the output section
		symbol := "•"
		var statusText string
		
		if res.Status == model.StatusPass {
			symbol = "✓"
			statusText = renderer.Green(string(res.Status))
		} else if res.Status == model.StatusFail {
			symbol = "✗"
			statusText = renderer.Red(string(res.Status))
		} else if res.Status == model.StatusSkipped {
			symbol = "⊘"
			statusText = renderer.Yellow(string(res.Status))
		} else {
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
		
		if res.Status == model.StatusPass {
			symbol = "✓"
			statusText = renderer.Green(string(res.Status))
		} else if res.Status == model.StatusFail {
			symbol = "✗"
			statusText = renderer.Red(string(res.Status))
		} else if res.Status == model.StatusSkipped {
			symbol = "⊘"
			statusText = renderer.Yellow(string(res.Status))
		} else {
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
	// Build full path to metrics file
	metricsPath := filepath.Join(st.Workdir, st.MetricsPath)
	
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
			if verbose {
				fmt.Fprintf(os.Stderr, "[%-15s] Failed to parse JUnit XML: %v\n", st.ID, err)
			}
			return nil
		}
		return m
	default:
		if verbose {
			fmt.Fprintf(os.Stderr, "[%-15s] Unknown metrics format: %s\n", st.ID, st.MetricsFormat)
		}
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
	tracker       *ui.AnimatedTaskTracker
	taskID        string
	file          *os.File
	buffer        []byte
	outputBuffer  *bytes.Buffer // Buffer all output until task completes
	mu            *sync.Mutex   // Protect outputBuffer
	console       *os.File      // For streaming output directly
	renderer      *ui.Renderer  // For colorizing output
}

func (w *lineWriter) Write(p []byte) (n int, err error) {
	// Write to log file (unprefixed)
	w.file.Write(p)
	
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
			fmt.Fprintln(w.console, prefixedLine)
		}
		
		w.buffer = w.buffer[idx+1:]
	}
	
	return len(p), nil
}
