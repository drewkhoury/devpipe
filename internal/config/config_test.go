package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid minimal config",
			content: `
[tasks.test]
command = "echo test"
`,
			wantErr: false,
		},
		{
			name: "valid config with defaults",
			content: `
[defaults]
uiMode = "basic"

[task_defaults]
enabled = true
workdir = "."

[tasks.test]
command = "echo test"
`,
			wantErr: false,
		},
		{
			name: "invalid toml",
			content: `
[tasks.test
command = "echo test"
`,
			wantErr: true,
		},
		{
			name:    "empty file",
			content: ``,
			wantErr: false, // Empty config is valid, will use defaults
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")

			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			// Load config
			_, _, _, _, err := LoadConfig(configPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveTaskConfig(t *testing.T) {
	cfg := &Config{
		TaskDefaults: TaskDefaultsConfig{
			Enabled: boolPtr(true),
			Workdir: ".",
			FixType: "auto",
		},
	}

	tests := []struct {
		name     string
		taskCfg  TaskConfig
		wantType string
	}{
		{
			name: "inherits fixType from defaults",
			taskCfg: TaskConfig{
				Command: "test",
			},
			wantType: "auto",
		},
		{
			name: "overrides fixType",
			taskCfg: TaskConfig{
				Command: "test",
				FixType: "helper",
			},
			wantType: "helper",
		},
		{
			name: "empty fixType uses default",
			taskCfg: TaskConfig{
				Command: "test",
				FixType: "",
			},
			wantType: "auto",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := cfg.ResolveTaskConfig("test", tt.taskCfg, "/tmp")

			if resolved.FixType != tt.wantType {
				t.Errorf("ResolveTaskConfig() fixType = %v, want %v", resolved.FixType, tt.wantType)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantValid bool
	}{
		{
			name: "valid config",
			cfg: Config{
				Defaults: DefaultsConfig{
					UIMode: "basic",
				},
				TaskDefaults: TaskDefaultsConfig{
					Enabled: boolPtr(true),
					Workdir: ".",
				},
				Tasks: map[string]TaskConfig{
					"test": {
						Command: "echo test",
					},
				},
			},
			wantValid: true,
		},
		{
			name: "invalid ui mode",
			cfg: Config{
				Defaults: DefaultsConfig{
					UIMode: "invalid",
				},
				Tasks: map[string]TaskConfig{
					"test": {
						Command: "echo test",
					},
				},
			},
			wantValid: false,
		},
		{
			name: "task with fixType but no fixCommand",
			cfg: Config{
				Tasks: map[string]TaskConfig{
					"test": {
						Command: "echo test",
						FixType: "auto",
						// Missing fixCommand
					},
				},
			},
			wantValid: false,
		},
		{
			name: "task with invalid fixType",
			cfg: Config{
				Tasks: map[string]TaskConfig{
					"test": {
						Command:    "echo test",
						FixType:    "invalid",
						FixCommand: "echo fix",
					},
				},
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := ValidateConfig(&tt.cfg)

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateConfig() valid = %v, want %v", result.Valid, tt.wantValid)
				if !result.Valid {
					t.Logf("Errors: %v", result.Errors)
				}
			}
		})
	}
}

func TestMergeWithDefaults(t *testing.T) {
	cfg := Config{
		Defaults: DefaultsConfig{
			UIMode: "full",
		},
		TaskDefaults: TaskDefaultsConfig{
			Workdir: "/custom",
		},
	}

	merged := MergeWithDefaults(&cfg)

	// Should preserve custom values
	if merged.Defaults.UIMode != "full" {
		t.Errorf("Expected UIMode 'full', got '%s'", merged.Defaults.UIMode)
	}

	if merged.TaskDefaults.Workdir != "/custom" {
		t.Errorf("Expected workdir '/custom', got '%s'", merged.TaskDefaults.Workdir)
	}

	// Should fill in missing defaults
	if merged.TaskDefaults.Enabled == nil {
		t.Error("Expected Enabled to be set by defaults")
	}
}

func TestValidateGitConfig(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantValid bool
	}{
		{
			name: "valid git mode staged",
			cfg: Config{
				Defaults: DefaultsConfig{
					Git: GitConfig{
						Mode: "staged",
					},
				},
			},
			wantValid: true,
		},
		{
			name: "valid git mode ref with ref specified",
			cfg: Config{
				Defaults: DefaultsConfig{
					Git: GitConfig{
						Mode: "ref",
						Ref:  "main",
					},
				},
			},
			wantValid: true,
		},
		{
			name: "invalid git mode",
			cfg: Config{
				Defaults: DefaultsConfig{
					Git: GitConfig{
						Mode: "invalid",
					},
				},
			},
			wantValid: false,
		},
		{
			name: "git mode ref without ref is valid (ref is optional)",
			cfg: Config{
				Defaults: DefaultsConfig{
					Git: GitConfig{
						Mode: "ref",
					},
				},
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := ValidateConfig(&tt.cfg)

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateConfig() valid = %v, want %v", result.Valid, tt.wantValid)
				if !result.Valid {
					t.Logf("Errors: %v", result.Errors)
				}
			}
		})
	}
}

func TestValidateTaskDefaults(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantValid bool
	}{
		{
			name: "valid task defaults",
			cfg: Config{
				TaskDefaults: TaskDefaultsConfig{
					Enabled: boolPtr(true),
					Workdir: ".",
					FixType: "auto",
				},
			},
			wantValid: true,
		},
		{
			name: "invalid fixType in defaults",
			cfg: Config{
				TaskDefaults: TaskDefaultsConfig{
					FixType: "invalid",
				},
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := ValidateConfig(&tt.cfg)

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateConfig() valid = %v, want %v", result.Valid, tt.wantValid)
			}
		})
	}
}

func TestExtractTaskOrder(t *testing.T) {
	// Test that task order extraction works
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `[tasks.test1]
command = "echo test1"

[tasks.test2]
command = "echo test2"`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, taskOrder, _, _, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(taskOrder) != 2 {
		t.Errorf("Expected 2 tasks in order, got %d", len(taskOrder))
	}
}

func TestLoadConfigWithPhases(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `[tasks."phase:build"]
wait = true

[tasks.compile]
command = "go build"

[tasks."phase:test"]
wait = true

[tasks.unittest]
command = "go test"`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, _, phaseInfo, _, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify config was loaded
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	// Phase info extraction is optional - just verify no error
	t.Logf("Loaded config with %d phase markers", len(phaseInfo))
}

func TestValidateTaskWithMetrics(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantValid bool
	}{
		{
			name: "valid junit metrics",
			cfg: Config{
				Tasks: map[string]TaskConfig{
					"test": {
						Command:    "go test",
						OutputType: "junit",
						OutputPath: "results.xml",
					},
				},
			},
			wantValid: true,
		},
		{
			name: "valid sarif metrics",
			cfg: Config{
				Tasks: map[string]TaskConfig{
					"security": {
						Command:    "gosec",
						OutputType: "sarif",
						OutputPath: "results.sarif",
					},
				},
			},
			wantValid: true,
		},
		{
			name: "invalid metrics format",
			cfg: Config{
				Tasks: map[string]TaskConfig{
					"test": {
						Command:    "go test",
						OutputType: "invalid",
						OutputPath: "results.xml",
					},
				},
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := ValidateConfig(&tt.cfg)

			if result.Valid != tt.wantValid {
				t.Errorf("ValidateConfig() valid = %v, want %v", result.Valid, tt.wantValid)
			}
		})
	}
}

func TestGetDefaults(t *testing.T) {
	defaults := GetDefaults()

	// Should have default values
	if defaults.Defaults.OutputRoot == "" {
		t.Error("Expected non-empty OutputRoot default")
	}

	if defaults.TaskDefaults.Enabled == nil {
		t.Error("Expected Enabled default to be set")
	}

	// Verify default values are reasonable
	if defaults.Defaults.FastThreshold <= 0 {
		t.Error("Expected positive FastThreshold default")
	}
}

func TestLoadConfigErrors(t *testing.T) {
	// Test nonexistent file
	_, _, _, _, err := LoadConfig("/nonexistent/config.toml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	// Test invalid TOML in actual file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "bad.toml")

	if err := os.WriteFile(configPath, []byte("invalid toml [["), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, _, _, _, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid TOML")
	}
}

func TestMergeWithDefaultsComprehensive(t *testing.T) {
	cfg := Config{
		Defaults: DefaultsConfig{
			UIMode:        "full",
			FastThreshold: 5000,
		},
		TaskDefaults: TaskDefaultsConfig{
			Enabled: boolPtr(false),
			Workdir: "/custom",
			FixType: "helper",
		},
		Tasks: map[string]TaskConfig{
			"test": {
				Command: "go test",
			},
		},
	}

	merged := MergeWithDefaults(&cfg)

	// Should preserve custom values
	if merged.Defaults.UIMode != "full" {
		t.Errorf("Expected UIMode 'full', got '%s'", merged.Defaults.UIMode)
	}

	if merged.Defaults.FastThreshold != 5000 {
		t.Errorf("Expected FastThreshold 5000, got %d", merged.Defaults.FastThreshold)
	}

	if merged.TaskDefaults.FixType != "helper" {
		t.Errorf("Expected FixType 'helper', got '%s'", merged.TaskDefaults.FixType)
	}

	// Should have filled in OutputRoot from defaults
	if merged.Defaults.OutputRoot == "" {
		t.Error("Expected OutputRoot to be filled from defaults")
	}
}

func TestValidateMultipleTasks(t *testing.T) {
	cfg := Config{
		Tasks: map[string]TaskConfig{
			"build": {
				Command: "go build",
				Type:    "build",
			},
			"test": {
				Command: "go test",
				Type:    "test",
			},
			"lint": {
				Command: "golangci-lint run",
				Type:    "check",
			},
		},
	}

	result, _ := ValidateConfig(&cfg)

	// All tasks should be valid
	if !result.Valid {
		t.Errorf("Expected valid config, got errors: %v", result.Errors)
	}
}

func TestBuiltInTasks(t *testing.T) {
	// Test that BuiltInTasks returns tasks
	tasks := BuiltInTasks(".")

	if len(tasks) == 0 {
		t.Error("Expected at least one built-in task")
	}

	// Verify tasks have required fields
	for id, task := range tasks {
		if task.Command == "" {
			t.Errorf("Task %s has empty command", id)
		}
	}
}

func TestGetTaskOrder(t *testing.T) {
	// Test GetTaskOrder with built-in tasks
	order := GetTaskOrder()

	if len(order) == 0 {
		t.Error("Expected at least one task in order")
	}
}

func TestGenerateDefaultConfig(t *testing.T) {
	// Test that GenerateDefaultConfig creates a file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	err := GenerateDefaultConfig(configPath, ".")
	if err != nil {
		t.Fatalf("GenerateDefaultConfig() error = %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	contentStr := string(content)

	// Should contain task definitions
	if !strings.Contains(contentStr, "[tasks") {
		t.Error("Expected [tasks] section in generated config")
	}

	// Should be valid TOML
	if len(contentStr) == 0 {
		t.Error("Expected non-empty generated config")
	}
}

func TestValidationResultErrors(t *testing.T) {
	result := ValidationResult{
		Valid:  false,
		Errors: []ValidationError{{Message: "error 1"}, {Message: "error 2"}},
	}

	if result.Valid {
		t.Error("Expected result to be invalid")
	}

	if len(result.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(result.Errors))
	}
}

func TestValidateConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.toml")

	// Create valid config
	content := `[tasks.test]
command = "echo test"`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := ValidateConfigFile(configPath)
	if err != nil {
		t.Fatalf("ValidateConfigFile() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid config, got errors: %v", result.Errors)
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{-5, "-5"},
		{1000, "1000"},
	}

	for _, tt := range tests {
		got := intToString(tt.input)
		if got != tt.want {
			t.Errorf("intToString(%d) = %s, want %s", tt.input, got, tt.want)
		}
	}
}

func TestExtractQuotedValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "quoted string",
			input: `"hello world"`,
			want:  "hello world",
		},
		{
			name:  "quoted with spaces",
			input: `  "test"  `,
			want:  "test",
		},
		{
			name:  "unquoted",
			input: "plain",
			want:  "plain",
		},
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "single quote",
			input: `"`,
			want:  `"`,
		},
		{
			name:  "no closing quote",
			input: `"unclosed`,
			want:  `"unclosed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractQuotedValue(tt.input)
			if got != tt.want {
				t.Errorf("extractQuotedValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  ValidationError
		want string
	}{
		{
			name: "with field",
			err: ValidationError{
				Field:   "tasks.test.command",
				Message: "Task must have a command",
			},
			want: "tasks.test.command: Task must have a command",
		},
		{
			name: "without field",
			err: ValidationError{
				Message: "Invalid configuration",
			},
			want: "Invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrintValidationResult(t *testing.T) {
	// Test valid config
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Should not panic
	PrintValidationResult("test.toml", result)

	// Test with errors
	resultWithErrors := &ValidationResult{
		Valid: false,
		Errors: []ValidationError{
			{Field: "tasks.test", Message: "Missing command"},
		},
		Warnings: []ValidationError{
			{Message: "Deprecated option"},
		},
	}

	// Should not panic
	PrintValidationResult("test.toml", resultWithErrors)

	t.Log("PrintValidationResult completed without panic")
}

func TestLoadConfigUnknownFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Config with unknown field
	content := `[tasks.test]
command = "echo test"
unknownField = "value"`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, _, _, _, err := LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for unknown field")
	}
	if !strings.Contains(err.Error(), "unknown fields") {
		t.Errorf("Expected 'unknown fields' error, got: %v", err)
	}
}

func TestLoadConfigNoPath(t *testing.T) {
	// Test with empty path and no config.toml in current directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }() // Best effort restore

	_ = os.Chdir(tmpDir) // Test setup

	cfg, _, _, _, err := LoadConfig("")
	if err != nil {
		t.Errorf("Expected no error when config.toml doesn't exist with empty path, got: %v", err)
	}
	if cfg != nil {
		t.Error("Expected nil config when file doesn't exist")
	}
}

func TestExtractTaskOrderWithPhases(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Config with phase headers
	content := `[tasks.phase-validation]
name = "Validation Phase"
desc = "Run validation checks"

[tasks.lint]
command = "golangci-lint run"

[tasks.phase-build]
name = "Build Phase"
desc = "Build the application"

[tasks.compile]
command = "go build"`

	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	taskOrder, phaseInfo, taskToPhase, err := extractTaskOrder(configPath)
	if err != nil {
		t.Fatalf("extractTaskOrder() error = %v", err)
	}

	// Should have wait markers and tasks
	if len(taskOrder) < 2 {
		t.Errorf("Expected at least 2 items in order, got %d", len(taskOrder))
	}

	// Should have phase info
	if len(phaseInfo) == 0 {
		t.Error("Expected phase info to be extracted")
	}

	// Verify phase names
	foundPhase := false
	for _, info := range phaseInfo {
		if info.Name == "Validation Phase" {
			foundPhase = true
			if info.Desc != "Run validation checks" {
				t.Errorf("Expected phase desc 'Run validation checks', got '%s'", info.Desc)
			}
		}
	}
	if !foundPhase {
		t.Error("Expected to find 'Validation Phase' in phase info")
	}

	// Verify task-to-phase mapping
	if taskToPhase["lint"] != "phase-validation" {
		t.Errorf("Expected lint to map to phase-validation, got %s", taskToPhase["lint"])
	}
}

func TestExtractTaskOrderEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		content  string
		wantErr  bool
		minTasks int
	}{
		{
			name: "malformed section header",
			content: `[tasks.test
command = "echo"`,
			wantErr:  false, // extractTaskOrder doesn't fail on malformed TOML
			minTasks: 0,
		},
		{
			name: "phase without name",
			content: `[tasks.phase-test]
desc = "Test phase"

[tasks.test]
command = "echo test"`,
			wantErr:  false,
			minTasks: 1,
		},
		{
			name: "multiple phases",
			content: `[tasks.phase-one]
name = "Phase One"

[tasks.task1]
command = "echo 1"

[tasks.phase-two]
name = "Phase Two"

[tasks.task2]
command = "echo 2"`,
			wantErr:  false,
			minTasks: 3, // 2 tasks + 1 wait marker
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, tt.name+".toml")
			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			taskOrder, _, _, err := extractTaskOrder(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTaskOrder() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(taskOrder) < tt.minTasks {
				t.Errorf("Expected at least %d tasks, got %d", tt.minTasks, len(taskOrder))
			}
		})
	}
}

func TestGenerateDefaultConfigAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create file first
	if err := os.WriteFile(configPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Try to generate - should fail
	err := GenerateDefaultConfig(configPath, ".")
	if err == nil {
		t.Error("Expected error when file already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %v", err)
	}
}

func TestMergeWithDefaultsNilConfig(t *testing.T) {
	merged := MergeWithDefaults(nil)

	// Should return defaults
	if merged.Defaults.OutputRoot == "" {
		t.Error("Expected OutputRoot to be set from defaults")
	}
	if merged.TaskDefaults.Enabled == nil {
		t.Error("Expected Enabled to be set from defaults")
	}
}

func TestMergeWithDefaultsEmptyFields(t *testing.T) {
	cfg := Config{
		Defaults: DefaultsConfig{
			// All fields empty/zero
		},
		TaskDefaults: TaskDefaultsConfig{
			// All fields empty/zero
		},
	}

	merged := MergeWithDefaults(&cfg)

	// All fields should be filled from defaults
	if merged.Defaults.OutputRoot == "" {
		t.Error("Expected OutputRoot to be filled from defaults")
	}
	if merged.Defaults.FastThreshold == 0 {
		t.Error("Expected FastThreshold to be filled from defaults")
	}
	if merged.Defaults.UIMode == "" {
		t.Error("Expected UIMode to be filled from defaults")
	}
	if merged.Defaults.AnimationRefreshMs == 0 {
		t.Error("Expected AnimationRefreshMs to be filled from defaults")
	}
	if merged.Defaults.AnimatedGroupBy == "" {
		t.Error("Expected AnimatedGroupBy to be filled from defaults")
	}
	if merged.Defaults.Git.Mode == "" {
		t.Error("Expected Git.Mode to be filled from defaults")
	}
	if merged.Defaults.Git.Ref == "" {
		t.Error("Expected Git.Ref to be filled from defaults")
	}
	if merged.TaskDefaults.Enabled == nil {
		t.Error("Expected TaskDefaults.Enabled to be filled from defaults")
	}
	if merged.TaskDefaults.Workdir == "" {
		t.Error("Expected TaskDefaults.Workdir to be filled from defaults")
	}
}

func TestResolveTaskConfigWorkdir(t *testing.T) {
	cfg := &Config{
		TaskDefaults: TaskDefaultsConfig{
			Workdir: "./custom",
		},
	}

	tests := []struct {
		name        string
		taskCfg     TaskConfig
		projectRoot string
		wantWorkdir string
	}{
		{
			name: "uses task workdir if set",
			taskCfg: TaskConfig{
				Command: "test",
				Workdir: "/absolute/path",
			},
			projectRoot: "/repo",
			wantWorkdir: "/absolute/path",
		},
		{
			name: "uses default workdir if task workdir empty",
			taskCfg: TaskConfig{
				Command: "test",
			},
			projectRoot: "/repo",
			wantWorkdir: "/repo/custom", // filepath.Join normalizes ./custom to custom
		},
		{
			name: "makes relative workdir absolute",
			taskCfg: TaskConfig{
				Command: "test",
				Workdir: "subdir",
			},
			projectRoot: "/repo",
			wantWorkdir: "/repo/subdir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := cfg.ResolveTaskConfig("test", tt.taskCfg, tt.projectRoot)
			if resolved.Workdir != tt.wantWorkdir {
				t.Errorf("ResolveTaskConfig() workdir = %v, want %v", resolved.Workdir, tt.wantWorkdir)
			}
		})
	}
}

func TestResolveTaskConfigEnabled(t *testing.T) {
	cfg := &Config{
		TaskDefaults: TaskDefaultsConfig{
			Enabled: boolPtr(false),
		},
	}

	// Task without enabled set should inherit from defaults
	taskCfg := TaskConfig{
		Command: "test",
	}

	resolved := cfg.ResolveTaskConfig("test", taskCfg, "/repo")
	if resolved.Enabled == nil {
		t.Error("Expected Enabled to be set")
	}
	if *resolved.Enabled != false {
		t.Error("Expected Enabled to be false from defaults")
	}

	// Task with enabled set should keep its value
	taskCfg2 := TaskConfig{
		Command: "test",
		Enabled: boolPtr(true),
	}

	resolved2 := cfg.ResolveTaskConfig("test", taskCfg2, "/repo")
	if *resolved2.Enabled != true {
		t.Error("Expected Enabled to be true from task config")
	}
}

func TestExtractTaskOrderFileReadError(t *testing.T) {
	// Test with nonexistent file
	_, _, _, err := extractTaskOrder("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestResolveTaskConfigWithNoDefaults(t *testing.T) {
	cfg := &Config{
		TaskDefaults: TaskDefaultsConfig{
			// Empty defaults
		},
	}

	taskCfg := TaskConfig{
		Command: "test",
	}

	resolved := cfg.ResolveTaskConfig("test", taskCfg, "/repo")

	// Should use fallback workdir "."
	if resolved.Workdir != "/repo" {
		t.Errorf("Expected workdir '/repo', got '%s'", resolved.Workdir)
	}
}
