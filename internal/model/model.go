// Package model defines the core data structures for devpipe tasks and results.
package model

// TaskStatus represents the status of a task
type TaskStatus string

// Task status constants
const (
	StatusPending TaskStatus = "PENDING"
	StatusRunning TaskStatus = "RUNNING"
	StatusPass    TaskStatus = "PASS"
	StatusFail    TaskStatus = "FAIL"
	StatusSkipped TaskStatus = "SKIPPED"
)

// TaskDefinition is the resolved definition of a task ready to execute
type TaskDefinition struct {
	ID               string
	Name             string
	Desc             string
	Phase            string
	Type             string
	Command          string
	Workdir          string
	EstimatedSeconds int
	IsEstimateGuess  bool     // True if estimate is a default guess (show as "10s?")
	Wait             bool     // If true, marks end of phase (wait for all previous tasks)
	MetricsFormat    string   // "junit", "eslint", etc.
	MetricsPath      string   // Path to metrics file
	FixType          string   // "auto", "helper", "none", or ""
	FixCommand       string   // Command to run to fix issues
	WatchPaths       []string // Glob patterns to watch (relative to workdir)
}

// TaskResult is the per-task record written into run.json
type TaskResult struct {
	ID                string       `json:"id"`
	Name              string       `json:"name"`
	Desc              string       `json:"desc,omitempty"`
	Phase             string       `json:"phase,omitempty"`
	Type              string       `json:"type"`
	Status            TaskStatus   `json:"status"`
	ExitCode          *int         `json:"exitCode,omitempty"`
	Skipped           bool         `json:"skipped"`
	SkipReason        string       `json:"skipReason,omitempty"`
	Command           string       `json:"command"`
	Workdir           string       `json:"workdir"`
	LogPath           string       `json:"logPath"`
	StartTime         string       `json:"startTime,omitempty"`
	EndTime           string       `json:"endTime,omitempty"`
	DurationMs        int64        `json:"durationMs"`
	EstimatedSeconds  int          `json:"estimatedSeconds"`
	AutoFixed         bool         `json:"autoFixed,omitempty"`
	FixCommand        string       `json:"fixCommand,omitempty"`
	InitialExitCode   *int         `json:"initialExitCode,omitempty"`
	FixDurationMs     int64        `json:"fixDurationMs,omitempty"`
	RecheckDurationMs int64        `json:"recheckDurationMs,omitempty"`
	Metrics           *TaskMetrics `json:"metrics,omitempty"`
}

// TaskMetrics holds parsed metrics from task artifacts
type TaskMetrics struct {
	Kind          string                 `json:"kind"`                    // "test", "lint", "coverage", "build"
	SummaryFormat string                 `json:"summaryFormat,omitempty"` // "junit", "eslint", "sarif"
	Data          map[string]interface{} `json:"data,omitempty"`
}

// TestMetrics holds test-specific metrics
type TestMetrics struct {
	Tests    int     `json:"tests"`
	Failures int     `json:"failures"`
	Errors   int     `json:"errors"`
	Skipped  int     `json:"skipped"`
	Time     float64 `json:"time"`
}

// RunFlags captures CLI flags for run.json
type RunFlags struct {
	Fast     bool     `json:"fast"`
	FailFast bool     `json:"failFast"`
	DryRun   bool     `json:"dryRun"`
	Verbose  bool     `json:"verbose"`
	Only     string   `json:"only,omitempty"`
	Skip     []string `json:"skip,omitempty"`
	Config   string   `json:"config,omitempty"`
	Since    string   `json:"since,omitempty"`
}

// ConfigValue represents a single configuration value with its source
type ConfigValue struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Source   string `json:"source"`             // "config-file", "cli-flag", "default", "historical"
	Overrode string `json:"overrode,omitempty"` // What value it replaced, if any
}

// EffectiveConfig holds the resolved configuration with source tracking
type EffectiveConfig struct {
	Values []ConfigValue `json:"values"`
}

// RunRecord is the top-level JSON written per run
// Note: GitInfo is imported from the git package
type RunRecord struct {
	RunID           string           `json:"runId"`
	Timestamp       string           `json:"timestamp"`
	RepoRoot        string           `json:"repoRoot"`
	OutputRoot      string           `json:"outputRoot"`
	ConfigPath      string           `json:"configPath,omitempty"`
	Command         string           `json:"command,omitempty"` // Full command line that was executed
	PipelineVersion string           `json:"pipelineVersion"`   // devpipe version used to run the pipeline
	ReportVersion   string           `json:"reportVersion"`     // devpipe version used to generate the report
	Git             interface{}      `json:"git"`               // git.GitInfo
	Flags           RunFlags         `json:"flags"`
	Tasks           []TaskResult     `json:"tasks"`
	EffectiveConfig *EffectiveConfig `json:"effectiveConfig,omitempty"`
}
