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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/drew/devpipe/internal/config"
	"github.com/drew/devpipe/internal/git"
	"github.com/drew/devpipe/internal/model"
	"github.com/drew/devpipe/internal/ui"
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
	flag.StringVar(&flagOnly, "only", "", "Run only a single stage by id")
	flag.StringVar(&flagUI, "ui", "basic", "UI mode: basic, full")
	flag.BoolVar(&flagNoColor, "no-color", false, "Disable colored output")
	flag.BoolVar(&flagAnimated, "animated", false, "Show live progress animation (experimental)")
	flag.Var(&flagSkipVals, "skip", "Skip a stage by id (can be specified multiple times)")
	flag.BoolVar(&flagFailFast, "fail-fast", false, "Stop on first stage failure")
	flag.BoolVar(&flagDryRun, "dry-run", false, "Do not execute commands, simulate only")
	flag.BoolVar(&flagVerbose, "verbose", false, "Verbose logging")
	flag.BoolVar(&flagFast, "fast", false, "Skip long running stages")
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
	
	// Determine which stages to use
	var stages map[string]config.StageConfig
	var stageOrder []string
	
	if cfg == nil || len(cfg.Stages) == 0 {
		// No config file or no stages defined, use built-in
		if flagVerbose {
			fmt.Println("No config file found, using built-in stages")
		}
		stages = config.BuiltInStages(repoRoot)
		stageOrder = config.GetStageOrder()
	} else {
		// Use stages from config
		stages = mergedCfg.Stages
		// Use the built-in order if stages match, otherwise alphabetical
		builtInOrder := config.GetStageOrder()
		for _, id := range builtInOrder {
			if _, exists := stages[id]; exists {
				stageOrder = append(stageOrder, id)
			}
		}
		// Add any additional stages not in built-in order
		for id := range stages {
			found := false
			for _, existing := range stageOrder {
				if existing == id {
					found = true
					break
				}
			}
			if !found {
				stageOrder = append(stageOrder, id)
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

	// Build stage definitions
	var stageDefs []model.StageDefinition
	for _, id := range stageOrder {
		stageCfg, ok := stages[id]
		if !ok {
			continue
		}
		
		// Resolve with defaults
		resolved := mergedCfg.ResolveStageConfig(id, stageCfg, repoRoot)
		
		// Skip if disabled
		if resolved.Enabled != nil && !*resolved.Enabled {
			if flagVerbose {
				fmt.Printf("[%-15s] DISABLED in config\n", id)
			}
			continue
		}
		
		stageDefs = append(stageDefs, model.StageDefinition{
			ID:               id,
			Name:             resolved.Name,
			Group:            resolved.Group,
			Command:          resolved.Command,
			Workdir:          resolved.Workdir,
			EstimatedSeconds: resolved.EstimatedSeconds,
		})
	}

	// Apply CLI filters
	filteredStages := filterStages(stageDefs, flagOnly, flagSkipVals, flagFast, mergedCfg.Defaults.FastThreshold, flagVerbose)

	// Run stages
	var (
		results         []model.StageResult
		overallExitCode int
		anyFailed       bool
	)

	// Render header
	renderer.RenderHeader(runID, repoRoot, gitMode, len(gitInfo.ChangedFiles))
	
	// Setup animation if enabled
	var tracker *ui.AnimatedStageTracker
	
	// Ensure cursor is restored on exit (in case of panic or early exit)
	defer func() {
		if tracker != nil {
			fmt.Print("\033[?25h") // Show cursor
		}
	}()
	
	if renderer.IsAnimated() {
		// Build stage progress list
		var stageProgress []ui.StageProgress
		for _, st := range filteredStages {
			stageProgress = append(stageProgress, ui.StageProgress{
				ID:               st.ID,
				Name:             st.Name,
				Group:            st.Group,
				Status:           "PENDING",
				EstimatedSeconds: st.EstimatedSeconds,
				ElapsedSeconds:   0,
				StartTime:        time.Time{},
			})
		}
		
		// Calculate header lines
		headerLines := 4 // basic mode: run ID, repo, git mode, changed files
		if renderer.IsAnimated() {
			headerLines = 4
		}
		
		tracker = renderer.CreateAnimatedTracker(stageProgress, headerLines, mergedCfg.Defaults.AnimationRefreshMs)
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

	for _, st := range filteredStages {
		// Check if should skip due to --fast
		longRunning := st.EstimatedSeconds >= mergedCfg.Defaults.FastThreshold
		if flagFast && longRunning && flagOnly == "" {
			reason := fmt.Sprintf("skipped by --fast (est %ds)", st.EstimatedSeconds)
			
			// Update tracker if animated
			if tracker != nil {
				tracker.UpdateStage(st.ID, "SKIPPED", 0)
			}
			
			renderer.RenderStageSkipped(st.ID, reason, flagVerbose)
			results = append(results, model.StageResult{
				ID:               st.ID,
				Name:             st.Name,
				Group:            st.Group,
				Status:           model.StatusSkipped,
				Skipped:          true,
				SkipReason:       "skipped by --fast",
				Command:          st.Command,
				Workdir:          st.Workdir,
				LogPath:          "",
				EstimatedSeconds: st.EstimatedSeconds,
			})
			continue
		}

		res, _ := runStage(st, runDir, logDir, flagDryRun, flagVerbose, renderer, tracker)
		results = append(results, res)

		if res.Status == model.StatusFail {
			anyFailed = true
			overallExitCode = 1
			if flagFailFast {
				if flagVerbose {
					fmt.Printf("[%-15s] FAIL, stopping due to --fail-fast\n", st.ID)
				}
				break
			}
		}
	}
	
	// Stop animation if it was running
	if tracker != nil {
		tracker.Stop()
	}

	// Build run record
	now := time.Now().UTC().Format(time.RFC3339)
	runRecord := model.RunRecord{
		RunID:      runID,
		Timestamp:  now,
		RepoRoot:   repoRoot,
		OutputRoot: outputRoot,
		ConfigPath: flagConfig,
		Git:        gitInfo,
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
		Stages: results,
	}

	if err := writeRunJSON(runDir, runRecord); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to write run.json: %v\n", err)
		if overallExitCode == 0 {
			overallExitCode = 1
		}
	}

	// Final summary
	var summaries []ui.StageSummary
	for _, r := range results {
		summaries = append(summaries, ui.StageSummary{
			ID:         r.ID,
			Status:     string(r.Status),
			DurationMs: r.DurationMs,
		})
	}
	renderer.RenderSummary(summaries, anyFailed)
	os.Exit(overallExitCode)
}

func filterStages(stages []model.StageDefinition, only string, skip sliceFlag, fast bool, fastThreshold int, verbose bool) []model.StageDefinition {
	skipSet := map[string]struct{}{}
	for _, id := range skip {
		skipSet[id] = struct{}{}
	}

	var out []model.StageDefinition
	if only != "" {
		for _, s := range stages {
			if s.ID == only {
				out = append(out, s)
				return out
			}
		}
		fmt.Fprintf(os.Stderr, "ERROR: --only stage id %q not found\n", only)
		os.Exit(1)
	}

	for _, s := range stages {
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

func runStage(st model.StageDefinition, runDir, logDir string, dryRun bool, verbose bool, renderer *ui.Renderer, tracker *ui.AnimatedStageTracker) (model.StageResult, error) {
	res := model.StageResult{
		ID:               st.ID,
		Name:             st.Name,
		Group:            st.Group,
		Status:           model.StatusPending,
		Command:          st.Command,
		Workdir:          st.Workdir,
		LogPath:          "",
		EstimatedSeconds: st.EstimatedSeconds,
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", st.ID))
	res.LogPath = logPath

	if dryRun {
		renderer.RenderStageSkipped(st.ID, "dry-run", verbose)
		res.Status = model.StatusSkipped
		res.Skipped = true
		res.SkipReason = "dry-run"
		return res, nil
	}

	renderer.RenderStageStart(st.ID, st.Command, verbose)

	start := time.Now().UTC()
	res.StartTime = start.Format(time.RFC3339)
	res.Status = model.StatusRunning
	
	// Update tracker if animated
	if tracker != nil {
		tracker.UpdateStage(st.ID, "RUNNING", 0)
	}

	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot create log file %s: %v\n", logPath, err)
		res.Status = model.StatusFail
		return res, err
	}
	defer logFile.Close()

	cmd := exec.Command("sh", "-c", st.Command)
	cmd.Dir = st.Workdir

	// Setup output handling
	if tracker != nil {
		// Animated mode: capture output and send to tracker
		stdoutWriter := &lineWriter{tracker: tracker, file: logFile}
		stderrWriter := &lineWriter{tracker: tracker, file: logFile}
		cmd.Stdout = stdoutWriter
		cmd.Stderr = stderrWriter
	} else {
		// Non-animated: show on console and log
		cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
		cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
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
					tracker.UpdateStage(st.ID, "RUNNING", elapsed)
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
			tracker.UpdateStage(st.ID, "FAIL", elapsed)
		}
		
		renderer.RenderStageComplete(st.ID, string(res.Status), &exitCode, res.DurationMs, verbose)
		return res, err
	}

	res.Status = model.StatusPass
	res.ExitCode = &exitCode
	
	// Update tracker with final status
	if tracker != nil {
		tracker.UpdateStage(st.ID, "PASS", elapsed)
	}
	
	renderer.RenderStageComplete(st.ID, string(res.Status), &exitCode, res.DurationMs, verbose)
	return res, nil
}

func writeRunJSON(runDir string, record model.RunRecord) error {
	path := filepath.Join(runDir, "run.json")
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// lineWriter captures output line by line and sends to tracker
type lineWriter struct {
	tracker *ui.AnimatedStageTracker
	file    *os.File
	buffer  []byte
}

func (w *lineWriter) Write(p []byte) (n int, err error) {
	// Write to log file
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
		w.tracker.AddLogLine(line)
		w.buffer = w.buffer[idx+1:]
	}
	
	return len(p), nil
}
