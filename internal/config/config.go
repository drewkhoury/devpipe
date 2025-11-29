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
	Defaults      DefaultsConfig            `toml:"defaults"`
	TaskDefaults  TaskDefaultsConfig        `toml:"task_defaults"`
	Tasks         map[string]TaskConfig     `toml:"tasks"`
}

// DefaultsConfig holds global defaults
type DefaultsConfig struct {
	OutputRoot         string    `toml:"outputRoot"`
	FastThreshold      int       `toml:"fastThreshold"`
	UIMode             string    `toml:"uiMode"`
	AnimationRefreshMs int       `toml:"animationRefreshMs"`
	AnimatedGroupBy    string    `toml:"animatedGroupBy"` // "type" or "phase"
	Git                GitConfig `toml:"git"`
}

// GitConfig holds git-related configuration
type GitConfig struct {
	Mode string `toml:"mode"` // "staged", "staged_unstaged", "ref"
	Ref  string `toml:"ref"`  // used when mode = "ref"
}

// TaskDefaultsConfig holds default values for all tasks
type TaskDefaultsConfig struct {
	Enabled *bool  `toml:"enabled"`
	Workdir string `toml:"workdir"`
}

// TaskConfig represents a single task configuration
type TaskConfig struct {
	Name          string `toml:"name"`
	Desc          string `toml:"desc"` // Description (used for phase headers)
	Type          string `toml:"type"`
	Command       string `toml:"command"`
	Workdir       string `toml:"workdir"`
	Enabled       *bool  `toml:"enabled"`
	Wait          bool   `toml:"wait"`          // Internal use only: set automatically by phase headers
	MetricsFormat string `toml:"metricsFormat"` // "junit", "eslint", "sarif"
	MetricsPath   string `toml:"metricsPath"`   // Path to metrics file (relative to workdir)
}

// LoadConfig loads configuration from a TOML file
// Returns nil if file doesn't exist (use built-in defaults)
func LoadConfig(path string) (*Config, []string, map[string]PhaseInfo, error) {
	// If no path specified, look for config.toml in current directory
	explicitPath := path != ""
	if path == "" {
		path = "config.toml"
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// If user explicitly specified a config file, fail
		if explicitPath {
			return nil, nil, nil, fmt.Errorf("config file not found: %s", path)
		}
		// Otherwise, return nil to allow auto-generation
		return nil, nil, nil, nil
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	// Extract task order and phase names from the TOML file by parsing it manually
	taskOrder, phaseNames, err := extractTaskOrder(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to extract task order: %w", err)
	}

	return &cfg, taskOrder, phaseNames, nil
}

// GetDefaults returns the default configuration
func GetDefaults() Config {
	return Config{
		Defaults: DefaultsConfig{
			OutputRoot:         ".devpipe",
			FastThreshold:      300,
			UIMode:             "basic",
			AnimationRefreshMs: 500, // 500ms = 2 FPS (efficient default)
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
func (c *Config) ResolveTaskConfig(id string, taskCfg TaskConfig, repoRoot string) TaskConfig {
	// Apply task defaults
	if taskCfg.Workdir == "" {
		if c.TaskDefaults.Workdir != "" {
			taskCfg.Workdir = c.TaskDefaults.Workdir
		} else {
			taskCfg.Workdir = "."
		}
	}

	// Make workdir absolute relative to repo root
	if !filepath.IsAbs(taskCfg.Workdir) {
		taskCfg.Workdir = filepath.Join(repoRoot, taskCfg.Workdir)
	}

	if taskCfg.Enabled == nil {
		taskCfg.Enabled = c.TaskDefaults.Enabled
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
	Name string
	Desc string
}

// extractTaskOrder parses the TOML file to extract the order of [tasks.X] sections
// Tasks starting with "phase-" are treated as phase headers - all tasks after a phase
// header belong to that phase until the next phase header
// Returns task order and a map of phase names
func extractTaskOrder(path string) ([]string, map[string]PhaseInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	var order []string
	phaseNames := make(map[string]PhaseInfo)
	lines := splitLines(string(data))
	phaseCounter := 0
	var currentPhaseID string
	
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
						phaseNames[currentPhaseID] = PhaseInfo{Name: phaseName, Desc: phaseDesc}
					}
					
					// Don't add the phase header itself to the order
					continue
				}
				
				order = append(order, taskID)
			}
		}
	}
	
	return order, phaseNames, nil
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
func GenerateDefaultConfig(path string, repoRoot string) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}
	
	// Write to file
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()
	
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
