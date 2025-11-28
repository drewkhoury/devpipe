package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the complete devpipe configuration
type Config struct {
	Defaults      DefaultsConfig              `toml:"defaults"`
	StageDefaults StageDefaultsConfig         `toml:"stage_defaults"`
	Stages        map[string]StageConfig      `toml:"stages"`
}

// DefaultsConfig holds global defaults
type DefaultsConfig struct {
	OutputRoot     string    `toml:"outputRoot"`
	FastThreshold  int       `toml:"fastThreshold"`
	Git            GitConfig `toml:"git"`
}

// GitConfig holds git-related configuration
type GitConfig struct {
	Mode string `toml:"mode"` // "staged", "staged_unstaged", "ref"
	Ref  string `toml:"ref"`  // used when mode = "ref"
}

// StageDefaultsConfig holds default values for all stages
type StageDefaultsConfig struct {
	Enabled          *bool  `toml:"enabled"`
	Workdir          string `toml:"workdir"`
	EstimatedSeconds int    `toml:"estimatedSeconds"`
}

// StageConfig represents a single stage configuration
type StageConfig struct {
	Name             string `toml:"name"`
	Group            string `toml:"group"`
	Command          string `toml:"command"`
	Workdir          string `toml:"workdir"`
	EstimatedSeconds int    `toml:"estimatedSeconds"`
	Enabled          *bool  `toml:"enabled"`
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
			OutputRoot:    ".devpipe",
			FastThreshold: 300,
			Git: GitConfig{
				Mode: "staged_unstaged",
				Ref:  "HEAD",
			},
		},
		StageDefaults: StageDefaultsConfig{
			Enabled:          boolPtr(true),
			Workdir:          ".",
			EstimatedSeconds: 10,
		},
		Stages: make(map[string]StageConfig),
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
	if cfg.Defaults.Git.Mode == "" {
		cfg.Defaults.Git.Mode = defaults.Defaults.Git.Mode
	}
	if cfg.Defaults.Git.Ref == "" {
		cfg.Defaults.Git.Ref = defaults.Defaults.Git.Ref
	}

	// Merge stage defaults
	if cfg.StageDefaults.Enabled == nil {
		cfg.StageDefaults.Enabled = defaults.StageDefaults.Enabled
	}
	if cfg.StageDefaults.Workdir == "" {
		cfg.StageDefaults.Workdir = defaults.StageDefaults.Workdir
	}
	if cfg.StageDefaults.EstimatedSeconds == 0 {
		cfg.StageDefaults.EstimatedSeconds = defaults.StageDefaults.EstimatedSeconds
	}

	return *cfg
}

// ResolveStageConfig resolves a stage config by applying defaults
func (c *Config) ResolveStageConfig(id string, stageCfg StageConfig, repoRoot string) StageConfig {
	// Apply stage defaults
	if stageCfg.Workdir == "" {
		if c.StageDefaults.Workdir != "" {
			stageCfg.Workdir = c.StageDefaults.Workdir
		} else {
			stageCfg.Workdir = "."
		}
	}

	// Make workdir absolute relative to repo root
	if !filepath.IsAbs(stageCfg.Workdir) {
		stageCfg.Workdir = filepath.Join(repoRoot, stageCfg.Workdir)
	}

	if stageCfg.EstimatedSeconds == 0 {
		stageCfg.EstimatedSeconds = c.StageDefaults.EstimatedSeconds
	}

	if stageCfg.Enabled == nil {
		stageCfg.Enabled = c.StageDefaults.Enabled
	}

	return stageCfg
}

func boolPtr(b bool) *bool {
	return &b
}
