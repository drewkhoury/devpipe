// Package config handles loading, validation, and merging of devpipe configuration files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config represents the complete devpipe configuration
type Config struct {
	Defaults     DefaultsConfig        `toml:"defaults"`
	TaskDefaults TaskDefaultsConfig    `toml:"task_defaults"`
	Tasks        map[string]TaskConfig `toml:"tasks"`
}

// DefaultsConfig holds global defaults
type DefaultsConfig struct {
	// Repo/project root directory (optional, auto-detected if not set)
	ProjectRoot string `toml:"projectRoot" doc:"Repo/project root directory (optional override, auto-detected from git or config location if not set)"`
	// Directory for run outputs and logs
	OutputRoot string `toml:"outputRoot" doc:"Directory for run outputs and logs"`
	// Tasks longer than this (seconds) are skipped with --fast
	FastThreshold int `toml:"fastThreshold" doc:"Tasks longer than this (seconds) are skipped with --fast"`
	// UI mode: basic or full
	UIMode string `toml:"uiMode" doc:"UI mode: basic or full" enum:"basic,full"`
	// Dashboard refresh rate in milliseconds
	AnimationRefreshMs int `toml:"animationRefreshMs" doc:"Dashboard refresh rate in milliseconds"`
	// Group tasks by phase or type in dashboard
	AnimatedGroupBy string `toml:"animatedGroupBy" doc:"Group tasks by phase or type in dashboard" enum:"phase,type"`
	// Git integration settings
	Git GitConfig `toml:"git"`
}

// GitConfig holds git-related configuration
type GitConfig struct {
	// Git mode: staged, staged_unstaged, or ref
	Mode string `toml:"mode" doc:"Git mode: staged, staged_unstaged, or ref" enum:"staged,staged_unstaged,ref"`
	// Git ref to compare against when mode is ref
	Ref string `toml:"ref" doc:"Git ref to compare against when mode is ref"`
}

// TaskDefaultsConfig holds default values for all tasks
type TaskDefaultsConfig struct {
	// Whether tasks are enabled by default
	Enabled *bool `toml:"enabled" doc:"Whether tasks are enabled by default"`
	// Default working directory for tasks
	Workdir string `toml:"workdir" doc:"Default working directory for tasks"`
	// Default fix behavior: auto, helper, or none
	FixType string `toml:"fixType" doc:"Default fix behavior: auto, helper, or none" enum:"auto,helper,none"`
}

// TaskConfig represents a single task configuration
type TaskConfig struct {
	// Shell command to execute
	Command string `toml:"command" doc:"Shell command to execute" required:"true"`
	// Display name for the task
	Name string `toml:"name" doc:"Display name for the task"`
	// Description
	Desc string `toml:"desc" doc:"Description"`
	// Task type for grouping (e.g., check, build, test)
	Type string `toml:"type" doc:"Task type for grouping (e.g., check, build, test)"`
	// Working directory for this task
	Workdir string `toml:"workdir" doc:"Working directory for this task"`
	// Whether this task is enabled
	Enabled *bool `toml:"enabled" doc:"Whether this task is enabled"`
	// Internal use only: set automatically by phase headers
	Wait bool `toml:"wait"`
	// Output type: junit, sarif, artifact
	OutputType string `toml:"outputType" doc:"Output type: junit, sarif, artifact" enum:"junit,sarif,artifact"`
	// Path to output file (relative to workdir)
	OutputPath string `toml:"outputPath" doc:"Path to output file (relative to workdir)"`
	// Fix behavior: auto, helper, none (overrides task_defaults)
	FixType string `toml:"fixType" doc:"Fix behavior: auto, helper, none (overrides task_defaults)" enum:"auto,helper,none"`
	// Command to run to fix issues (required if fixType is set)
	FixCommand string `toml:"fixCommand" doc:"Command to run to fix issues (required if fixType is set)"`
	// File patterns to watch (glob patterns relative to workdir). Task runs only if matching files changed.
	WatchPaths []string `toml:"watchPaths" doc:"File patterns to watch (glob patterns relative to workdir). Task runs only if matching files changed."`
}

// LoadConfig loads configuration from a TOML file
// Returns config, task order, phase info, task-to-phase mapping, and error
func LoadConfig(path string) (*Config, []string, map[string]PhaseInfo, map[string]string, error) {
	// If no path specified, look for config.toml in current directory
	explicitPath := path != ""
	if path == "" {
		path = "config.toml"
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// If user explicitly specified a config file, fail
		if explicitPath {
			return nil, nil, nil, nil, fmt.Errorf("config file not found: %s", path)
		}
		// Otherwise, return nil to allow auto-generation
		return nil, nil, nil, nil, nil
	}

	var cfg Config
	metadata, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	// Check for unknown fields
	undecoded := metadata.Undecoded()
	if len(undecoded) > 0 {
		var unknownFields []string
		for _, key := range undecoded {
			unknownFields = append(unknownFields, key.String())
		}
		return nil, nil, nil, nil, fmt.Errorf("unknown fields in config: %s", strings.Join(unknownFields, ", "))
	}

	// Validate tasks - only command is required (except for phase headers and wait markers)
	for taskID, task := range cfg.Tasks {
		// Skip validation for phase headers (phase-*) and wait markers (wait, wait-*)
		if strings.HasPrefix(taskID, "phase-") || taskID == "wait" || strings.HasPrefix(taskID, "wait-") {
			continue
		}
		if task.Command == "" {
			return nil, nil, nil, nil, fmt.Errorf("task %q is missing required field: command", taskID)
		}
	}

	// Extract task order, phase names, and task-to-phase mapping from the TOML file
	taskOrder, phaseNames, taskToPhase, err := extractTaskOrder(path)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to extract task order: %w", err)
	}

	return &cfg, taskOrder, phaseNames, taskToPhase, nil
}

// GetDefaults returns the default configuration
func GetDefaults() Config {
	return Config{
		Defaults: DefaultsConfig{
			OutputRoot:         ".devpipe",
			FastThreshold:      300,
			UIMode:             "basic",
			AnimationRefreshMs: 500,     // 500ms = 2 FPS (efficient default)
			AnimatedGroupBy:    "phase", // "type" or "phase"
			Git: GitConfig{
				Mode: "staged_unstaged",
				Ref:  "HEAD",
			},
		},
		TaskDefaults: TaskDefaultsConfig{
			Enabled: boolPtr(true),
			Workdir: ".",
		},
		Tasks: make(map[string]TaskConfig),
	}
}

// MergeWithDefaults merges loaded config with defaults
func MergeWithDefaults(cfg *Config) Config {
	defaults := GetDefaults()

	if cfg == nil {
		return defaults
	}

	// Merge defaults
	if cfg.Defaults.OutputRoot == "" {
		cfg.Defaults.OutputRoot = defaults.Defaults.OutputRoot
	}
	if cfg.Defaults.FastThreshold == 0 {
		cfg.Defaults.FastThreshold = defaults.Defaults.FastThreshold
	}
	if cfg.Defaults.UIMode == "" {
		cfg.Defaults.UIMode = defaults.Defaults.UIMode
	}
	if cfg.Defaults.AnimationRefreshMs == 0 {
		cfg.Defaults.AnimationRefreshMs = defaults.Defaults.AnimationRefreshMs
	}
	if cfg.Defaults.AnimatedGroupBy == "" {
		cfg.Defaults.AnimatedGroupBy = defaults.Defaults.AnimatedGroupBy
	}
	if cfg.Defaults.Git.Mode == "" {
		cfg.Defaults.Git.Mode = defaults.Defaults.Git.Mode
	}
	if cfg.Defaults.Git.Ref == "" {
		cfg.Defaults.Git.Ref = defaults.Defaults.Git.Ref
	}

	// Merge task defaults
	if cfg.TaskDefaults.Enabled == nil {
		cfg.TaskDefaults.Enabled = defaults.TaskDefaults.Enabled
	}
	if cfg.TaskDefaults.Workdir == "" {
		cfg.TaskDefaults.Workdir = defaults.TaskDefaults.Workdir
	}

	return *cfg
}

// ResolveTaskConfig resolves a task config by applying defaults
func (c *Config) ResolveTaskConfig(_ string, taskCfg TaskConfig, projectRoot string) TaskConfig {
	// Apply task defaults
	if taskCfg.Workdir == "" {
		if c.TaskDefaults.Workdir != "" {
			taskCfg.Workdir = c.TaskDefaults.Workdir
		} else {
			taskCfg.Workdir = "."
		}
	}

	// Make workdir absolute relative to project root
	if !filepath.IsAbs(taskCfg.Workdir) {
		taskCfg.Workdir = filepath.Join(projectRoot, taskCfg.Workdir)
	}

	if taskCfg.Enabled == nil {
		taskCfg.Enabled = c.TaskDefaults.Enabled
	}

	// Inherit fixType from task_defaults if not set at task level
	if taskCfg.FixType == "" && c.TaskDefaults.FixType != "" {
		taskCfg.FixType = c.TaskDefaults.FixType
	}

	return taskCfg
}

func boolPtr(b bool) *bool {
	return &b
}

func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// PhaseInfo holds information about a phase
type PhaseInfo struct {
	ID   string
	Name string
	Desc string
}

// extractTaskOrder parses the TOML file to extract the order of [tasks.X] sections
// Tasks starting with "phase-" are treated as phase headers - all tasks after a phase
// header belong to that phase until the next phase header
// Returns task order, phase info map, and task-to-phase mapping
func extractTaskOrder(path string) ([]string, map[string]PhaseInfo, map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, err
	}

	var order []string
	phaseNames := make(map[string]PhaseInfo)
	taskToPhase := make(map[string]string)
	lines := splitLines(string(data))
	phaseCounter := 0
	var currentPhaseID string
	var currentPhaseMarker string

	// Parse [tasks.X] sections to build order
	// When we see [tasks.phase-*], insert a wait marker before it (except for the first phase)
	for i, line := range lines {
		trimmed := trimSpace(line)
		if len(trimmed) > 7 && trimmed[0] == '[' && trimmed[1:7] == "tasks." {
			// Extract task ID from [tasks.ID]
			end := -1
			for j := 7; j < len(trimmed); j++ {
				if trimmed[j] == ']' {
					end = j
					break
				}
			}
			if end > 7 {
				taskID := trimmed[7:end]

				// Check if this is a phase header
				if len(taskID) > 6 && taskID[0:6] == "phase-" {
					// Insert a wait marker before this phase (except for the first phase)
					if phaseCounter > 0 {
						order = append(order, "wait-"+intToString(phaseCounter))
					}
					phaseCounter++
					currentPhaseID = "wait-" + intToString(phaseCounter)
					currentPhaseMarker = taskID // Save the actual phase ID (e.g., "phase-validation")

					// Extract phase name and desc from following lines
					phaseName := ""
					phaseDesc := ""
					for j := i + 1; j < len(lines) && j < i+10; j++ {
						nextLine := trimSpace(lines[j])
						if len(nextLine) > 0 && nextLine[0] == '[' {
							break // Hit next section
						}
						if strings.HasPrefix(nextLine, "name = ") {
							phaseName = extractQuotedValue(nextLine[7:])
						}
						if strings.HasPrefix(nextLine, "desc = ") {
							phaseDesc = extractQuotedValue(nextLine[7:])
						}
					}

					if phaseName != "" {
						phaseNames[currentPhaseID] = PhaseInfo{
							ID:   currentPhaseMarker,
							Name: phaseName,
							Desc: phaseDesc,
						}
					}

					// Don't add the phase header itself to the order
					continue
				}

				// Regular task - map it to current phase
				if currentPhaseMarker != "" {
					taskToPhase[taskID] = currentPhaseMarker
				}
				order = append(order, taskID)
			}
		}
	}

	return order, phaseNames, taskToPhase, nil
}

func extractQuotedValue(s string) string {
	s = trimSpace(s)
	if len(s) >= 2 && s[0] == '"' {
		// Find closing quote
		for i := 1; i < len(s); i++ {
			if s[i] == '"' {
				return s[1:i]
			}
		}
	}
	return s
}

// GenerateDefaultConfig creates a minimal config.toml file
func GenerateDefaultConfig(path string, _ string) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}

	// Write to file
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close config file: %w", cerr)
		}
	}()

	// Write minimal, runnable config
	content := `# devpipe configuration file
# Full reference: https://github.com/drewkhoury/devpipe/blob/main/config.example.toml

[tasks.lint]
name = "Lint"
command = "echo 'Running linter...'"
type = "check"

[tasks.format]
name = "Format Check"
command = "echo 'Checking code formatting...'"
type = "check"

[tasks.build]
name = "Build"
command = "echo 'Building application...'"
type = "build"
`

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
