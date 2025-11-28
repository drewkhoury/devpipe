// devpipe - Iteration 1
//
// Minimal local pipeline runner with hardcoded stages.
//
// Build:
//   go build -o devpipe .
//
// Example usage:
//   ./devpipe
//   ./devpipe --verbose
//   ./devpipe --only unit-tests
//   ./devpipe --skip lint --skip format
//   ./devpipe --fail-fast
//   ./devpipe --dry-run
//
// To test quickly, use the companion ./hello-world.sh script as commands.
//
// This iteration provides:
//   - Hardcoded stages and commands
//   - Basic CLI flags (--only, --skip, --fail-fast, --dry-run, --verbose, --fast)
//   - Git repo root detection (or fallback to CWD)
//   - Simple changed file detection (git diff --name-only HEAD)
//   - Per-stage logs under .devpipe/runs/<run-id>/logs/
//   - run.json summary per run
//   - Plain text console output (no TUI yet)

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type StageStatus string

const (
	StatusPending StageStatus = "PENDING"
	StatusRunning StageStatus = "RUNNING"
	StatusPass    StageStatus = "PASS"
	StatusFail    StageStatus = "FAIL"
	StatusSkipped StageStatus = "SKIPPED"
)

// StageDefinition is the built in definition of a stage (iteration 1, no config).
type StageDefinition struct {
	ID              string
	Name            string
	Group           string
	Command         string
	Workdir         string
	EstimatedSeconds int
}

// StageResult is the per stage record written into run.json.
type StageResult struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Group           string      `json:"group"`
	Status          StageStatus `json:"status"`
	ExitCode        *int        `json:"exitCode,omitempty"`
	Skipped         bool        `json:"skipped"`
	SkipReason      string      `json:"skipReason,omitempty"`
	Command         string      `json:"command"`
	Workdir         string      `json:"workdir"`
	LogPath         string      `json:"logPath"`
	StartTime       string      `json:"startTime,omitempty"`
	EndTime         string      `json:"endTime,omitempty"`
	DurationMs      int64       `json:"durationMs"`
	EstimatedSeconds int        `json:"estimatedSeconds"`
}

// RunFlags captures CLI flags for run.json.
type RunFlags struct {
	Fast      bool     `json:"fast"`
	FailFast  bool     `json:"failFast"`
	DryRun    bool     `json:"dryRun"`
	Verbose   bool     `json:"verbose"`
	Only      string   `json:"only,omitempty"`
	Skip      []string `json:"skip,omitempty"`
}

// RunRecord is the top level JSON written per run.
type RunRecord struct {
	RunID      string        `json:"runId"`
	Timestamp  string        `json:"timestamp"`
	RepoRoot   string        `json:"repoRoot"`
	OutputRoot string        `json:"outputRoot"`
	Git        GitInfo       `json:"git"`
	Flags      RunFlags      `json:"flags"`
	Stages     []StageResult `json:"stages"`
}

// GitInfo holds minimal git metadata for iteration 1.
type GitInfo struct {
	InGitRepo    bool     `json:"inGitRepo"`
	RepoRoot     string   `json:"repoRoot"`
	DiffBase     string   `json:"diffBase"`
	ChangedFiles []string `json:"changedFiles"`
}

// sliceFlag allows repeating --skip.
type sliceFlag []string

func (s *sliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *sliceFlag) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func main() {
	// CLI flags
	var (
		flagOnly     string
		flagFailFast bool
		flagDryRun   bool
		flagVerbose  bool
		flagFast     bool
		flagSkipVals sliceFlag
	)
	flag.StringVar(&flagOnly, "only", "", "Run only a single stage by id")
	flag.Var(&flagSkipVals, "skip", "Skip a stage by id (can be specified multiple times)")
	flag.BoolVar(&flagFailFast, "fail-fast", false, "Stop on first stage failure")
	flag.BoolVar(&flagDryRun, "dry-run", false, "Do not execute commands, simulate only")
	flag.BoolVar(&flagVerbose, "verbose", false, "Verbose logging")
	flag.BoolVar(&flagFast, "fast", false, "Skip long running stages (est >= 300s)")
	flag.Parse()

	// Determine repo root (git or cwd).
	repoRoot, inGitRepo := detectRepoRoot()
	if !inGitRepo && flagVerbose {
		fmt.Println("WARNING: not in a git repo, using current directory as repo root")
	}

	// Minimal git info for iteration 1.
	gitInfo := detectGitChanges(repoRoot, inGitRepo, flagVerbose)

	// Prepare output dir for this run: .devpipe/runs/<run-id>
	outputRoot := filepath.Join(repoRoot, ".devpipe")
	runID := makeRunID()
	runDir := filepath.Join(outputRoot, "runs", runID)
	logDir := filepath.Join(runDir, "logs")

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to create run directories: %v\n", err)
		os.Exit(1)
	}

	// Build list of stages.
	stages := builtInStages(repoRoot)

	// Apply CLI filters (only/skip and fast).
	stageOrder := filterStages(stages, flagOnly, flagSkipVals, flagFast, flagVerbose)

	// Run stages sequentially.
	var (
		results         []StageResult
		overallExitCode int
		anyFailed       bool
	)

	fmt.Printf("devpipe run %s\n", runID)
	fmt.Printf("Repo root: %s\n", repoRoot)
	if inGitRepo {
		fmt.Printf("Changed files (HEAD): %d\n", len(gitInfo.ChangedFiles))
	}

	for _, st := range stageOrder {
		// Determine if stage should be skipped due to fast mode.
		longRunning := st.EstimatedSeconds >= 300
		if flagFast && longRunning && flagOnly == "" {
			if flagVerbose {
				fmt.Printf("[%-15s] SKIPPED by --fast (est %ds)\n", st.ID, st.EstimatedSeconds)
			}
			results = append(results, StageResult{
				ID:               st.ID,
				Name:             st.Name,
				Group:            st.Group,
				Status:           StatusSkipped,
				Skipped:          true,
				SkipReason:       "skipped by --fast",
				Command:          st.Command,
				Workdir:          st.Workdir,
				LogPath:          "",
				EstimatedSeconds: st.EstimatedSeconds,
			})
			continue
		}

		res, _ := runStage(st, runDir, logDir, flagDryRun, flagVerbose)
		results = append(results, res)

		if res.Status == StatusFail {
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

	// Build run record.
	now := time.Now().UTC().Format(time.RFC3339)
	runRecord := RunRecord{
		RunID:      runID,
		Timestamp:  now,
		RepoRoot:   repoRoot,
		OutputRoot: outputRoot,
		Git:        gitInfo,
		Flags: RunFlags{
			Fast:     flagFast,
			FailFast: flagFailFast,
			DryRun:   flagDryRun,
			Verbose:  flagVerbose,
			Only:     flagOnly,
			Skip:     flagSkipVals,
		},
		Stages: results,
	}

	if err := writeRunJSON(runDir, runRecord); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to write run.json: %v\n", err)
		if overallExitCode == 0 {
			overallExitCode = 1
		}
	}

	// Final summary to console.
	fmt.Println()
	fmt.Println("Summary:")
	for _, r := range results {
		statusText := string(r.Status)
		fmt.Printf("  %-15s %-10s %6dms\n", r.ID, statusText, r.DurationMs)
	}
	if anyFailed {
		fmt.Println("\ndevpipe: one or more stages failed")
	} else {
		fmt.Println("\ndevpipe: all stages passed or were skipped")
	}
	os.Exit(overallExitCode)
}

// builtInStages returns the hardcoded stage list for iteration 1.
// You can wire these commands to ./hello-world.sh initially.
func builtInStages(repoRoot string) []StageDefinition {
	script := filepath.Join(repoRoot, "hello-world.sh")
	return []StageDefinition{
		{
			ID:              "lint",
			Name:            "Lint",
			Group:           "quality",
			Command:         fmt.Sprintf("%s lint", script),
			Workdir:         repoRoot,
			EstimatedSeconds: 5,
		},
		{
			ID:              "format",
			Name:            "Format",
			Group:           "quality",
			Command:         fmt.Sprintf("%s format", script),
			Workdir:         repoRoot,
			EstimatedSeconds: 5,
		},
		{
			ID:              "type-check",
			Name:            "Type Check",
			Group:           "correctness",
			Command:         fmt.Sprintf("%s type-check", script),
			Workdir:         repoRoot,
			EstimatedSeconds: 10,
		},
		{
			ID:              "build",
			Name:            "Build",
			Group:           "release",
			Command:         fmt.Sprintf("%s build", script),
			Workdir:         repoRoot,
			EstimatedSeconds: 15,
		},
		{
			ID:              "unit-tests",
			Name:            "Unit Tests",
			Group:           "correctness",
			Command:         fmt.Sprintf("%s unit-tests", script),
			Workdir:         repoRoot,
			EstimatedSeconds: 20,
		},
		{
			ID:              "e2e-tests",
			Name:            "E2E Tests",
			Group:           "correctness",
			Command:         fmt.Sprintf("%s e2e-tests", script),
			Workdir:         repoRoot,
			EstimatedSeconds: 600, // 10 minutes, considered long running
		},
	}
}

// filterStages applies --only, --skip, and --fast (long-running) at the selection level.
// The actual --fast skipping happens later so we can record explicit SKIPPED results.
func filterStages(stages []StageDefinition, only string, skip sliceFlag, fast bool, verbose bool) []StageDefinition {
	skipSet := map[string]struct{}{}
	for _, id := range skip {
		skipSet[id] = struct{}{}
	}

	var out []StageDefinition
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

func detectRepoRoot() (string, bool) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		// Not a git repo, use cwd.
		cwd, err2 := os.Getwd()
		if err2 != nil {
			return ".", false
		}
		return cwd, false
	}
	root := strings.TrimSpace(buf.String())
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ".", false
		}
		return cwd, false
	}
	return root, true
}

func detectGitChanges(repoRoot string, inGitRepo bool, verbose bool) GitInfo {
	info := GitInfo{
		InGitRepo:    inGitRepo,
		RepoRoot:     repoRoot,
		DiffBase:     "HEAD",
		ChangedFiles: []string{},
	}
	if !inGitRepo {
		return info
	}

	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	cmd.Dir = repoRoot
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Run(); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "WARNING: git diff failed: %v\n", err)
		}
		return info
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var files []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			files = append(files, l)
		}
	}
	info.ChangedFiles = files
	return info
}

func makeRunID() string {
	now := time.Now().UTC()
	ts := now.Format("2006-01-02T15-04-05Z")
	rand.Seed(time.Now().UnixNano())
	suffix := rand.Intn(1_000_000)
	return fmt.Sprintf("%s_%06d", ts, suffix)
}

func runStage(st StageDefinition, runDir, logDir string, dryRun bool, verbose bool) (StageResult, error) {
	res := StageResult{
		ID:               st.ID,
		Name:             st.Name,
		Group:            st.Group,
		Status:           StatusPending,
		Command:          st.Command,
		Workdir:          st.Workdir,
		LogPath:          "",
		EstimatedSeconds: st.EstimatedSeconds,
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("%s.log", st.ID))
	res.LogPath = logPath

	if dryRun {
		if verbose {
			fmt.Printf("[%-15s] DRY RUN, command: %s\n", st.ID, st.Command)
		} else {
			fmt.Printf("[%-15s] DRY RUN\n", st.ID)
		}
		res.Status = StatusSkipped
		res.Skipped = true
		res.SkipReason = "dry-run"
		return res, nil
	}

	if verbose {
		fmt.Printf("[%-15s] RUN    %s\n", st.ID, st.Command)
	} else {
		fmt.Printf("[%-15s] RUN\n", st.ID)
	}

	start := time.Now().UTC()
	res.StartTime = start.Format(time.RFC3339)
	res.Status = StatusRunning

	logFile, err := os.Create(logPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot create log file %s: %v\n", logPath, err)
		res.Status = StatusFail
		return res, err
	}
	defer logFile.Close()

	cmd := exec.Command("sh", "-c", st.Command)
	cmd.Dir = st.Workdir

	// Send stdout and stderr to both console and log file.
	cmd.Stdout = ioMultiWriter(os.Stdout, logFile)
	cmd.Stderr = ioMultiWriter(os.Stderr, logFile)

	err = cmd.Run()
	end := time.Now().UTC()
	res.EndTime = end.Format(time.RFC3339)
	res.DurationMs = end.Sub(start).Milliseconds()

	exitCode := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			exitCode = ee.ExitCode()
		} else {
			exitCode = 1
		}
		res.Status = StatusFail
		res.ExitCode = &exitCode
		if verbose {
			fmt.Printf("[%-15s] FAIL (exit %d, %dms)\n", st.ID, exitCode, res.DurationMs)
		} else {
			fmt.Printf("[%-15s] FAIL (%dms)\n", st.ID, res.DurationMs)
		}
		return res, err
	}

	res.Status = StatusPass
	res.ExitCode = &exitCode
	if verbose {
		fmt.Printf("[%-15s] PASS (exit 0, %dms)\n", st.ID, res.DurationMs)
	} else {
		fmt.Printf("[%-15s] PASS (%dms)\n", st.ID, res.DurationMs)
	}
	return res, nil
}

func writeRunJSON(runDir string, record RunRecord) error {
	path := filepath.Join(runDir, "run.json")
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ioMultiWriter is a tiny helper to avoid importing io.MultiWriter for clarity if desired.
// Here we just re expose io.MultiWriter from stdlib.
func ioMultiWriter(writers ...*os.File) *multiWriter {
	return &multiWriter{writers: writers}
}

// multiWriter implements io.Writer over multiple *os.File.
type multiWriter struct {
	writers []*os.File
}

func (m *multiWriter) Write(p []byte) (int, error) {
	var firstN int
	for i, w := range m.writers {
		n, err := w.Write(p)
		if i == 0 {
			firstN = n
		}
		if err != nil {
			return n, err
		}
	}
	return firstN, nil
}
