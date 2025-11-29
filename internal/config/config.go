package config

import (
	"fmt"
	"os"
	"path/filepath"

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
	OutputRoot        string    `toml:"outputRoot"`
	FastThreshold     int       `toml:"fastThreshold"`
	UIMode            string    `toml:"uiMode"`
	AnimationRefreshMs int      `toml:"animationRefreshMs"`
	Git               GitConfig `toml:"git"`
}

// GitConfig holds git-related configuration
type GitConfig struct {
	Mode string `toml:"mode"` // "staged", "staged_unstaged", "ref"
	Ref  string `toml:"ref"`  // used when mode = "ref"
}

// TaskDefaultsConfig holds default values for all tasks
type TaskDefaultsConfig struct {
	Enabled          *bool  `toml:"enabled"`
	Workdir          string `toml:"workdir"`
	EstimatedSeconds int    `toml:"estimatedSeconds"`
}

// TaskConfig represents a single task configuration
type TaskConfig struct {
	Name             string `toml:"name"`
	Type             string `toml:"type"`
	Command          string `toml:"command"`
	Workdir          string `toml:"workdir"`
	EstimatedSeconds int    `toml:"estimatedSeconds"`
	Enabled          *bool  `toml:"enabled"`
	Wait             bool   `toml:"wait"`           // If true, wait for all previous tasks to complete before continuing
	MetricsFormat    string `toml:"metricsFormat"` // "junit", "eslint", "sarif"
	MetricsPath      string `toml:"metricsPath"`   // Path to metrics file (relative to workdir)
}

// LoadConfig loads configuration from a TOML file
// Returns nil if file doesn't exist (use built-in defaults)
func LoadConfig(path string) (*Config, error) {
	// If no path specified, look for config.toml in current directory
	if path == "" {
		path = "config.toml"
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil // No config file, use defaults
	}

	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	return &cfg, nil
}

// GetDefaults returns the default configuration
func GetDefaults() Config {
	return Config{
		Defaults: DefaultsConfig{
			OutputRoot:         ".devpipe",
			FastThreshold:      300,
			UIMode:             "basic",
			AnimationRefreshMs: 500, // 500ms = 2 FPS (efficient default)
			Git: GitConfig{
				Mode: "staged_unstaged",
				Ref:  "HEAD",
			},
		},
		TaskDefaults: TaskDefaultsConfig{
			Enabled:          boolPtr(true),
			Workdir:          ".",
			EstimatedSeconds: 10,
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
	if cfg.TaskDefaults.EstimatedSeconds == 0 {
		cfg.TaskDefaults.EstimatedSeconds = defaults.TaskDefaults.EstimatedSeconds
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

	if taskCfg.EstimatedSeconds == 0 {
		taskCfg.EstimatedSeconds = c.TaskDefaults.EstimatedSeconds
	}

	if taskCfg.Enabled == nil {
		taskCfg.Enabled = c.TaskDefaults.Enabled
	}

	return taskCfg
}

func boolPtr(b bool) *bool {
	return &b
}

// GenerateDefaultConfig creates a config.toml file with built-in task definitions
func GenerateDefaultConfig(path string, repoRoot string) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}
	
	// Build config from built-in tasks
	builtInTasks := BuiltInTasks(repoRoot)
	taskOrder := GetTaskOrder()
	
	// Create config structure
	cfg := Config{
		Defaults: DefaultsConfig{
			OutputRoot:    ".devpipe",
			FastThreshold: 300,
			Git: GitConfig{
				Mode: "staged_unstaged",
				Ref:  "main",
			},
		},
		TaskDefaults: TaskDefaultsConfig{
			Enabled:          boolPtr(true),
			Workdir:          ".",
			EstimatedSeconds: 10,
		},
		Tasks: make(map[string]TaskConfig),
	}
	
	// Add tasks in order
	for _, id := range taskOrder {
		if task, ok := builtInTasks[id]; ok {
			// Make paths relative
			task.Command = "./" + filepath.Base(task.Command)
			task.Workdir = "."
			cfg.Tasks[id] = task
		}
	}
	
	// Write to file
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()
	
	// Write header comment
	fmt.Fprintln(f, "# devpipe configuration file")
	fmt.Fprintln(f, "# Auto-generated on first run - customize as needed")
	fmt.Fprintln(f, "")
	
	// Encode config as TOML
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	
	return nil
}
