package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDefaults(t *testing.T) {
	tests := []struct {
		name      string
		defaults  DefaultsConfig
		wantValid bool
	}{
		{
			name: "valid defaults",
			defaults: DefaultsConfig{
				OutputRoot:    ".devpipe",
				UIMode:        "full",
				FastThreshold: 5000,
			},
			wantValid: true,
		},
		{
			name: "invalid UI mode",
			defaults: DefaultsConfig{
				OutputRoot:    ".devpipe",
				UIMode:        "invalid",
				FastThreshold: 5000,
			},
			wantValid: false,
		},
		{
			name: "negative fast threshold",
			defaults: DefaultsConfig{
				OutputRoot:    ".devpipe",
				UIMode:        "full",
				FastThreshold: -100,
			},
			wantValid: false,
		},
		{
			name: "invalid animated group by",
			defaults: DefaultsConfig{
				OutputRoot:      ".devpipe",
				UIMode:          "full",
				AnimatedGroupBy: "invalid",
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Valid:  true,
				Errors: []ValidationError{},
			}

			validateDefaults(&tt.defaults, result)

			if result.Valid != tt.wantValid {
				t.Errorf("validateDefaults() valid = %v, want %v, errors: %v",
					result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestValidatePhaseHeadersWithFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Test valid phase header
	validConfig := filepath.Join(tmpDir, "valid.toml")
	validContent := `[tasks."phase:build"]
wait = true

[tasks.compile]
command = "go build"`

	if err := os.WriteFile(validConfig, []byte(validContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	err := validatePhaseHeaders(validConfig, result)
	if err != nil {
		t.Fatalf("validatePhaseHeaders() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid config, got errors: %v", result.Errors)
	}
}

func TestValidateTaskMetrics(t *testing.T) {
	tests := []struct {
		name      string
		task      TaskConfig
		taskID    string
		wantValid bool
	}{
		{
			name: "valid junit metrics",
			task: TaskConfig{
				Command:    "go test",
				OutputType: "junit",
				OutputPath: "artifacts/junit.xml",
			},
			taskID:    "test",
			wantValid: true,
		},
		{
			name: "valid sarif metrics",
			task: TaskConfig{
				Command:    "golangci-lint run",
				OutputType: "sarif",
				OutputPath: "results.sarif",
			},
			taskID:    "lint",
			wantValid: true,
		},
		{
			name: "invalid output type",
			task: TaskConfig{
				Command:    "go test",
				OutputType: "invalid",
				OutputPath: "results.xml",
			},
			taskID:    "test",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Valid:  true,
				Errors: []ValidationError{},
			}

			validateTask(tt.taskID, tt.task, result)

			if result.Valid != tt.wantValid {
				t.Errorf("validateTask() valid = %v, want %v, errors: %v",
					result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestValidateGitConfigEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		git       GitConfig
		wantValid bool
	}{
		{
			name: "ref mode with ref",
			git: GitConfig{
				Mode: "ref",
				Ref:  "main",
			},
			wantValid: true,
		},
		{
			name: "staged mode",
			git: GitConfig{
				Mode: "staged",
			},
			wantValid: true,
		},
		{
			name: "invalid mode",
			git: GitConfig{
				Mode: "invalid",
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Valid:  true,
				Errors: []ValidationError{},
			}

			validateGitConfig(&tt.git, result)

			if result.Valid != tt.wantValid {
				t.Errorf("validateGitConfig() valid = %v, want %v, errors: %v",
					result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestValidateTaskDefaultsEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		defaults  TaskDefaultsConfig
		wantValid bool
	}{
		{
			name: "valid auto fix",
			defaults: TaskDefaultsConfig{
				FixType: "auto",
			},
			wantValid: true,
		},
		{
			name: "valid helper fix",
			defaults: TaskDefaultsConfig{
				FixType: "helper",
			},
			wantValid: true,
		},
		{
			name: "valid none fix",
			defaults: TaskDefaultsConfig{
				FixType: "none",
			},
			wantValid: true,
		},
		{
			name: "invalid fix type",
			defaults: TaskDefaultsConfig{
				FixType: "invalid",
			},
			wantValid: false,
		},
		{
			name: "empty fix type",
			defaults: TaskDefaultsConfig{
				FixType: "",
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Valid:  true,
				Errors: []ValidationError{},
			}

			validateTaskDefaults(&tt.defaults, result)

			if result.Valid != tt.wantValid {
				t.Errorf("validateTaskDefaults() valid = %v, want %v, errors: %v",
					result.Valid, tt.wantValid, result.Errors)
			}
		})
	}
}

func TestValidationErrorFormatting(t *testing.T) {
	err := ValidationError{
		Field:   "tasks.test.command",
		Message: "Command is required",
	}

	expected := "tasks.test.command: Command is required"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}

	// Test without field
	err2 := ValidationError{
		Message: "General error",
	}

	if err2.Error() != "General error" {
		t.Errorf("Error() = %q, want %q", err2.Error(), "General error")
	}
}

func TestValidateConfigFileNonexistent(t *testing.T) {
	_, err := ValidateConfigFile("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestValidateConfigFileInvalidTOML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.toml")

	// Write invalid TOML
	if err := os.WriteFile(configPath, []byte("invalid toml [["), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := ValidateConfigFile(configPath)
	if err != nil {
		t.Fatalf("ValidateConfigFile() unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for malformed TOML")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for invalid TOML")
	}
}

func TestValidateConfigFileUnknownFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "unknown.toml")

	content := `[tasks.test]
command = "echo test"
unknownField = "value"`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := ValidateConfigFile(configPath)
	if err != nil {
		t.Fatalf("ValidateConfigFile() unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result for unknown fields")
	}

	foundUnknownField := false
	for _, e := range result.Errors {
		if e.Message == "Unknown configuration field" {
			foundUnknownField = true
			break
		}
	}
	if !foundUnknownField {
		t.Error("Expected 'Unknown configuration field' error")
	}
}

func TestValidateTaskPhaseHeader(t *testing.T) {
	tests := []struct {
		name         string
		taskID       string
		task         TaskConfig
		wantWarnings int
	}{
		{
			name:   "phase header with command",
			taskID: "phase-build",
			task: TaskConfig{
				Name:    "Build Phase",
				Command: "should not have command",
			},
			wantWarnings: 1,
		},
		{
			name:   "phase header without name",
			taskID: "phase-test",
			task:   TaskConfig{
				// No name
			},
			wantWarnings: 1,
		},
		{
			name:   "phase header valid",
			taskID: "phase-validation",
			task: TaskConfig{
				Name: "Validation Phase",
			},
			wantWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Valid:    true,
				Errors:   []ValidationError{},
				Warnings: []ValidationError{},
			}

			validateTask(tt.taskID, tt.task, result)

			if len(result.Warnings) != tt.wantWarnings {
				t.Errorf("validateTask() warnings = %d, want %d, warnings: %v",
					len(result.Warnings), tt.wantWarnings, result.Warnings)
			}
		})
	}
}

func TestValidateTaskMissingCommand(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	task := TaskConfig{
		Name: "Test Task",
		// Missing command
	}

	validateTask("test", task, result)

	if result.Valid {
		t.Error("Expected invalid result for task without command")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected error for missing command")
	}
}

func TestValidateTaskMetricsWarnings(t *testing.T) {
	tests := []struct {
		name         string
		task         TaskConfig
		wantWarnings int
	}{
		{
			name: "outputType without outputPath",
			task: TaskConfig{
				Command:    "go test",
				OutputType: "junit",
				// Missing outputPath
			},
			wantWarnings: 1,
		},
		{
			name: "outputPath without outputType",
			task: TaskConfig{
				Command:    "go test",
				OutputPath: "results.xml",
				// Missing outputType
			},
			wantWarnings: 1,
		},
		{
			name: "both set correctly",
			task: TaskConfig{
				Command:    "go test",
				OutputType: "junit",
				OutputPath: "results.xml",
			},
			wantWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ValidationResult{
				Valid:    true,
				Errors:   []ValidationError{},
				Warnings: []ValidationError{},
			}

			validateTask("test", tt.task, result)

			if len(result.Warnings) != tt.wantWarnings {
				t.Errorf("validateTask() warnings = %d, want %d, warnings: %v",
					len(result.Warnings), tt.wantWarnings, result.Warnings)
			}
		})
	}
}

func TestValidateTaskFixTypeNone(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	task := TaskConfig{
		Command: "go test",
		FixType: "none",
		// No fixCommand needed when fixType is "none"
	}

	validateTask("test", task, result)

	if !result.Valid {
		t.Errorf("Expected valid result for fixType=none without fixCommand, errors: %v", result.Errors)
	}
}

func TestValidateDefaultsNegativeAnimationRefresh(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	defaults := DefaultsConfig{
		AnimationRefreshMs: -100,
	}

	validateDefaults(&defaults, result)

	if result.Valid {
		t.Error("Expected invalid result for negative animation refresh")
	}
}

func TestValidateConfigNil(t *testing.T) {
	result, err := ValidateConfig(nil)
	if err != nil {
		t.Fatalf("ValidateConfig() unexpected error: %v", err)
	}

	if !result.Valid {
		t.Error("Expected valid result for nil config")
	}
}

func TestPrintValidationResultWithFieldErrors(_ *testing.T) {
	result := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{Field: "tasks.test.command", Message: "Missing command"},
			{Message: "General error without field"},
		},
		Warnings: []ValidationError{
			{Field: "tasks.lint.metricsPath", Message: "Missing path"},
			{Message: "General warning"},
		},
	}

	// Should not panic
	PrintValidationResult("test.toml", result)
}

func TestPrintValidationResultValidWithWarnings(_ *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
		Warnings: []ValidationError{
			{Field: "tasks.test", Message: "Deprecated option"},
		},
	}

	// Should not panic
	PrintValidationResult("test.toml", result)
}

func TestValidatePhaseHeadersNonexistentFile(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	err := validatePhaseHeaders("/nonexistent/file.toml", result)
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestContainsHelper(t *testing.T) {
	slice := []string{"apple", "banana", "cherry"}

	if !contains(slice, "banana") {
		t.Error("Expected contains to return true for 'banana'")
	}

	if contains(slice, "orange") {
		t.Error("Expected contains to return false for 'orange'")
	}

	if contains([]string{}, "anything") {
		t.Error("Expected contains to return false for empty slice")
	}
}

func TestValidatePhaseHeadersWithPhases(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "phases.toml")

	content := `[tasks.phase-build]
name = "Build Phase"

[tasks.compile]
command = "go build"

[tasks.phase-test]
name = "Test Phase"

[tasks.unittest]
command = "go test"`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	err := validatePhaseHeaders(configPath, result)
	if err != nil {
		t.Fatalf("validatePhaseHeaders() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid result, got errors: %v", result.Errors)
	}
}
