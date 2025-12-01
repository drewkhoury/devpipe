package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// ValidationResult holds the results of config validation
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
}

// ValidateConfig validates an already-loaded config
func ValidateConfig(cfg *Config) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	if cfg == nil {
		return result, nil
	}

	// Validate defaults section
	validateDefaults(&cfg.Defaults, result)

	// Validate task_defaults section
	validateTaskDefaults(&cfg.TaskDefaults, result)

	// Validate tasks
	for taskID, task := range cfg.Tasks {
		validateTask(taskID, task, result)
	}

	return result, nil
}

// ValidateConfigFile validates a TOML config file
func ValidateConfigFile(path string) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	// Try to parse as TOML first
	var cfg Config
	metadata, err := toml.DecodeFile(path, &cfg)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Message: fmt.Sprintf("Invalid TOML syntax: %v", err),
		})
		return result, nil
	}

	// Check for unknown fields
	undecoded := metadata.Undecoded()
	if len(undecoded) > 0 {
		result.Valid = false
		for _, key := range undecoded {
			result.Errors = append(result.Errors, ValidationError{
				Field:   key.String(),
				Message: "Unknown configuration field",
			})
		}
	}

	// Validate defaults section
	validateDefaults(&cfg.Defaults, result)

	// Validate task_defaults section
	validateTaskDefaults(&cfg.TaskDefaults, result)

	// Validate tasks
	for taskID, task := range cfg.Tasks {
		validateTask(taskID, task, result)
	}

	// Additional validation: check for phase headers
	if err := validatePhaseHeaders(path, result); err != nil {
		result.Warnings = append(result.Warnings, ValidationError{
			Message: fmt.Sprintf("Could not validate phase headers: %v", err),
		})
	}

	return result, nil
}

// validateDefaults validates the defaults section
func validateDefaults(defaults *DefaultsConfig, result *ValidationResult) {
	// Validate UIMode
	if defaults.UIMode != "" {
		validModes := []string{"basic", "full"}
		if !contains(validModes, defaults.UIMode) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "defaults.uiMode",
				Message: fmt.Sprintf("Invalid UI mode '%s'. Valid options: %s", defaults.UIMode, strings.Join(validModes, ", ")),
			})
		}
	}

	// Validate AnimatedGroupBy
	if defaults.AnimatedGroupBy != "" {
		validGroupBy := []string{"type", "phase"}
		if !contains(validGroupBy, defaults.AnimatedGroupBy) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "defaults.animatedGroupBy",
				Message: fmt.Sprintf("Invalid groupBy '%s'. Valid options: %s", defaults.AnimatedGroupBy, strings.Join(validGroupBy, ", ")),
			})
		}
	}

	// Validate FastThreshold
	if defaults.FastThreshold < 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "defaults.fastThreshold",
			Message: "Fast threshold must be non-negative",
		})
	}

	// Validate AnimationRefreshMs
	if defaults.AnimationRefreshMs < 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "defaults.animationRefreshMs",
			Message: "Animation refresh rate must be non-negative",
		})
	}

	// Validate Git config
	validateGitConfig(&defaults.Git, result)
}

// validateGitConfig validates git configuration
func validateGitConfig(git *GitConfig, result *ValidationResult) {
	if git.Mode != "" {
		validModes := []string{"staged", "staged_unstaged", "ref"}
		if !contains(validModes, git.Mode) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "defaults.git.mode",
				Message: fmt.Sprintf("Invalid git mode '%s'. Valid options: %s", git.Mode, strings.Join(validModes, ", ")),
			})
		}
	}

	// Warn if mode is "ref" but no ref is specified
	if git.Mode == "ref" && git.Ref == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "defaults.git.ref",
			Message: "Git mode is 'ref' but no ref is specified",
		})
	}
}

// validateTaskDefaults validates the task_defaults section
func validateTaskDefaults(taskDefaults *TaskDefaultsConfig, result *ValidationResult) {
	// Validate fixType if specified
	if taskDefaults.FixType != "" {
		validFixTypes := []string{"auto", "helper", "none"}
		if !contains(validFixTypes, taskDefaults.FixType) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "task_defaults.fixType",
				Message: fmt.Sprintf("Invalid fix type '%s'. Valid options: %s", taskDefaults.FixType, strings.Join(validFixTypes, ", ")),
			})
		}
	}
}

// validateTask validates a single task configuration
func validateTask(taskID string, task TaskConfig, result *ValidationResult) {
	prefix := fmt.Sprintf("tasks.%s", taskID)

	// Check if it's a phase header
	if strings.HasPrefix(taskID, "phase-") {
		// Phase headers should have name but no command
		if task.Command != "" {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   prefix + ".command",
				Message: "Phase headers should not have a command",
			})
		}
		if task.Name == "" {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   prefix + ".name",
				Message: "Phase header should have a name",
			})
		}
		return
	}

	// Regular tasks should have a command
	if task.Command == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   prefix + ".command",
			Message: "Task must have a command",
		})
	}

	// Note: task.Type is user-defined and can be any string, so we don't validate it

	// Validate metricsFormat if specified
	if task.MetricsFormat != "" {
		validFormats := []string{"junit", "artifact"}
		if !contains(validFormats, task.MetricsFormat) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   prefix + ".metricsFormat",
				Message: fmt.Sprintf("Invalid metrics format '%s'. Valid options: %s", task.MetricsFormat, strings.Join(validFormats, ", ")),
			})
		}

		// Warn if metricsFormat is set but metricsPath is not
		if task.MetricsPath == "" {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   prefix + ".metricsPath",
				Message: "metricsFormat is set but metricsPath is not specified",
			})
		}
	}

	// Validate fixType if specified
	if task.FixType != "" {
		validFixTypes := []string{"auto", "helper", "none"}
		if !contains(validFixTypes, task.FixType) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   prefix + ".fixType",
				Message: fmt.Sprintf("Invalid fix type '%s'. Valid options: %s", task.FixType, strings.Join(validFixTypes, ", ")),
			})
		}

		// ERROR if fixType is set at task level (and not "none") but no fixCommand
		if task.FixType != "none" && task.FixCommand == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   prefix + ".fixCommand",
				Message: "fixType is set but fixCommand is not specified",
			})
		}
	}

	// Warn if metricsPath is set but metricsFormat is not
	if task.MetricsPath != "" && task.MetricsFormat == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   prefix + ".metricsFormat",
			Message: "metricsPath is set but metricsFormat is not specified",
		})
	}
}

// validatePhaseHeaders checks that phase headers are properly formatted
func validatePhaseHeaders(path string, result *ValidationResult) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var currentSection string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for section headers
		if strings.HasPrefix(trimmed, "[tasks.phase-") {
			// Extract section name
			end := strings.Index(trimmed, "]")
			if end > 0 {
				currentSection = trimmed[1:end]
			}
		}

		// If we're in a phase section, check for required fields
		if currentSection != "" && strings.HasPrefix(currentSection, "tasks.phase-") {
			if strings.HasPrefix(trimmed, "[") && trimmed != "["+currentSection+"]" {
				// We've moved to a new section
				currentSection = ""
			}
		}
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// PrintValidationResult prints the validation result in a human-readable format
func PrintValidationResult(path string, result *ValidationResult) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📋 Validating: %s\n", path)

	if result.Valid && len(result.Warnings) == 0 {
		fmt.Println("✅ Configuration is valid!")
		fmt.Println()
		return
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n❌ Found %d error(s):\n", len(result.Errors))
		for _, err := range result.Errors {
			if err.Field != "" {
				fmt.Printf("  • [%s] %s\n", err.Field, err.Message)
			} else {
				fmt.Printf("  • %s\n", err.Message)
			}
		}
		fmt.Println()
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("⚠️  Found %d warning(s):\n", len(result.Warnings))
		for _, warn := range result.Warnings {
			if warn.Field != "" {
				fmt.Printf("  • [%s] %s\n", warn.Field, warn.Message)
			} else {
				fmt.Printf("  • %s\n", warn.Message)
			}
		}
		fmt.Println()
	}

	if !result.Valid {
		fmt.Println("❌ Configuration is INVALID")
	} else {
		fmt.Println("✅ Configuration is valid (with warnings)")
	}
	fmt.Println()
}
