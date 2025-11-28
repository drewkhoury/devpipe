package model

// StageStatus represents the status of a stage
type StageStatus string

const (
	StatusPending StageStatus = "PENDING"
	StatusRunning StageStatus = "RUNNING"
	StatusPass    StageStatus = "PASS"
	StatusFail    StageStatus = "FAIL"
	StatusSkipped StageStatus = "SKIPPED"
)

// StageDefinition is the resolved definition of a stage ready to execute
type StageDefinition struct {
	ID               string
	Name             string
	Group            string
	Command          string
	Workdir          string
	EstimatedSeconds int
}

// StageResult is the per-stage record written into run.json
type StageResult struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Group            string      `json:"group"`
	Status           StageStatus `json:"status"`
	ExitCode         *int        `json:"exitCode,omitempty"`
	Skipped          bool        `json:"skipped"`
	SkipReason       string      `json:"skipReason,omitempty"`
	Command          string      `json:"command"`
	Workdir          string      `json:"workdir"`
	LogPath          string      `json:"logPath"`
	StartTime        string      `json:"startTime,omitempty"`
	EndTime          string      `json:"endTime,omitempty"`
	DurationMs       int64       `json:"durationMs"`
	EstimatedSeconds int         `json:"estimatedSeconds"`
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

// RunRecord is the top-level JSON written per run
// Note: GitInfo is imported from the git package
type RunRecord struct {
	RunID      string        `json:"runId"`
	Timestamp  string        `json:"timestamp"`
	RepoRoot   string        `json:"repoRoot"`
	OutputRoot string        `json:"outputRoot"`
	ConfigPath string        `json:"configPath,omitempty"`
	Git        interface{}   `json:"git"` // git.GitInfo
	Flags      RunFlags      `json:"flags"`
	Stages     []StageResult `json:"stages"`
}
