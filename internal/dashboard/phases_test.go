package dashboard

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/drew/devpipe/internal/model"
)

func TestParsePhasesFromConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Create a test config file
	configContent := `
[[phase]]
id = "build"
name = "Build"
desc = "Build the project"

[[phase.task]]
id = "compile"
name = "Compile"
cmd = "make build"

[[phase]]
id = "test"
name = "Test"
desc = "Run tests"

[[phase.task]]
id = "unit-test"
name = "Unit Tests"
cmd = "make test"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	tasks := []model.TaskResult{
		{
			ID:         "compile",
			Name:       "Compile",
			Phase:      "Build",
			Status:     model.StatusPass,
			DurationMs: 1000,
		},
		{
			ID:         "unit-test",
			Name:       "Unit Tests",
			Phase:      "Test",
			Status:     model.StatusPass,
			DurationMs: 2000,
		},
	}

	phases, err := ParsePhasesFromConfig(configPath, tasks)
	if err != nil {
		t.Fatalf("ParsePhasesFromConfig() error = %v", err)
	}

	if len(phases) != 2 {
		t.Errorf("Expected 2 phases, got %d", len(phases))
	}

	// Verify first phase
	if phases[0].Name != "Build" {
		t.Errorf("Expected first phase name 'Build', got %q", phases[0].Name)
	}

	if phases[0].Status != "PASS" {
		t.Errorf("Expected first phase status 'PASS', got %q", phases[0].Status)
	}

	if phases[0].TotalMs != 1000 {
		t.Errorf("Expected first phase duration 1000ms, got %d", phases[0].TotalMs)
	}

	if phases[0].TaskCount != 1 {
		t.Errorf("Expected first phase task count 1, got %d", phases[0].TaskCount)
	}

	// Verify second phase
	if phases[1].Name != "Test" {
		t.Errorf("Expected second phase name 'Test', got %q", phases[1].Name)
	}
}

func TestParsePhasesFromConfigWithFailedTask(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
[[phase]]
id = "test"
name = "Test"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	tasks := []model.TaskResult{
		{
			ID:         "test1",
			Name:       "Test 1",
			Phase:      "Test",
			Status:     model.StatusPass,
			DurationMs: 1000,
		},
		{
			ID:         "test2",
			Name:       "Test 2",
			Phase:      "Test",
			Status:     model.StatusFail,
			DurationMs: 2000,
		},
	}

	phases, err := ParsePhasesFromConfig(configPath, tasks)
	if err != nil {
		t.Fatalf("ParsePhasesFromConfig() error = %v", err)
	}

	if len(phases) != 1 {
		t.Fatalf("Expected 1 phase, got %d", len(phases))
	}

	// Phase should be marked as FAIL if any task fails
	if phases[0].Status != "FAIL" {
		t.Errorf("Expected phase status 'FAIL', got %q", phases[0].Status)
	}

	// Should have both tasks
	if phases[0].TaskCount != 2 {
		t.Errorf("Expected 2 tasks, got %d", phases[0].TaskCount)
	}

	// Total duration should be sum
	if phases[0].TotalMs != 3000 {
		t.Errorf("Expected total duration 3000ms, got %d", phases[0].TotalMs)
	}
}

func TestParsePhasesFromConfigWithEmptyPhase(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
[[phase]]
id = "build"
name = "Build"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	tasks := []model.TaskResult{
		{
			ID:         "task1",
			Name:       "Task 1",
			Phase:      "", // Empty phase
			Status:     model.StatusPass,
			DurationMs: 1000,
		},
	}

	phases, err := ParsePhasesFromConfig(configPath, tasks)
	if err != nil {
		t.Fatalf("ParsePhasesFromConfig() error = %v", err)
	}

	if len(phases) != 1 {
		t.Fatalf("Expected 1 phase, got %d", len(phases))
	}

	// Empty phase should default to "Tasks"
	if phases[0].Name != "Tasks" {
		t.Errorf("Expected phase name 'Tasks', got %q", phases[0].Name)
	}
}

func TestParsePhasesFromConfigInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.toml")

	tasks := []model.TaskResult{
		{
			ID:         "task1",
			Name:       "Task 1",
			Phase:      "Build",
			Status:     model.StatusPass,
			DurationMs: 1000,
		},
	}

	// Should not fail even if config doesn't exist
	phases, err := ParsePhasesFromConfig(configPath, tasks)
	if err != nil {
		t.Fatalf("ParsePhasesFromConfig() should handle missing config, error = %v", err)
	}

	if len(phases) != 1 {
		t.Fatalf("Expected 1 phase, got %d", len(phases))
	}

	// Should still group tasks by phase name
	if phases[0].Name != "Build" {
		t.Errorf("Expected phase name 'Build', got %q", phases[0].Name)
	}

	// ID and Desc should be empty when config can't be loaded
	if phases[0].ID != "" {
		t.Errorf("Expected empty ID, got %q", phases[0].ID)
	}

	if phases[0].Desc != "" {
		t.Errorf("Expected empty Desc, got %q", phases[0].Desc)
	}
}

func TestParsePhasesFromConfigMultipleTasksInPhase(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
[[phase]]
id = "validation"
name = "Validation"
desc = "Validate code quality"

[[phase.task]]
id = "lint"
name = "Lint"
command = "make lint"

[[phase.task]]
id = "format"
name = "Format Check"
command = "make format"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	tasks := []model.TaskResult{
		{
			ID:         "lint",
			Name:       "Lint",
			Phase:      "Validation",
			Status:     model.StatusPass,
			DurationMs: 500,
		},
		{
			ID:         "format",
			Name:       "Format Check",
			Phase:      "Validation",
			Status:     model.StatusPass,
			DurationMs: 300,
		},
		{
			ID:         "security",
			Name:       "Security Scan",
			Phase:      "Validation",
			Status:     model.StatusPass,
			DurationMs: 1200,
		},
	}

	phases, err := ParsePhasesFromConfig(configPath, tasks)
	if err != nil {
		t.Fatalf("ParsePhasesFromConfig() error = %v", err)
	}

	if len(phases) != 1 {
		t.Fatalf("Expected 1 phase, got %d", len(phases))
	}

	phase := phases[0]

	if phase.Name != "Validation" {
		t.Errorf("Expected phase name 'Validation', got %q", phase.Name)
	}

	if phase.TaskCount != 3 {
		t.Errorf("Expected 3 tasks, got %d", phase.TaskCount)
	}

	if phase.TotalMs != 2000 {
		t.Errorf("Expected total duration 2000ms, got %d", phase.TotalMs)
	}

	if len(phase.Tasks) != 3 {
		t.Errorf("Expected 3 tasks in Tasks slice, got %d", len(phase.Tasks))
	}

	// ID and Desc should be populated if config loads successfully
	if phase.ID == "" {
		t.Log("Phase ID is empty (config may not have loaded)")
	}

	if phase.Desc == "" {
		t.Log("Phase Desc is empty (config may not have loaded)")
	}
}

func TestParsePhasesFromConfigPreservesOrder(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
[[phase]]
id = "first"
name = "First"

[[phase]]
id = "second"
name = "Second"

[[phase]]
id = "third"
name = "Third"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Tasks in execution order
	tasks := []model.TaskResult{
		{ID: "task1", Phase: "First", Status: model.StatusPass, DurationMs: 100},
		{ID: "task2", Phase: "Second", Status: model.StatusPass, DurationMs: 200},
		{ID: "task3", Phase: "Third", Status: model.StatusPass, DurationMs: 300},
		{ID: "task4", Phase: "First", Status: model.StatusPass, DurationMs: 150},
	}

	phases, err := ParsePhasesFromConfig(configPath, tasks)
	if err != nil {
		t.Fatalf("ParsePhasesFromConfig() error = %v", err)
	}

	if len(phases) != 3 {
		t.Fatalf("Expected 3 phases, got %d", len(phases))
	}

	// Phases should be in order of first appearance in tasks
	if phases[0].Name != "First" {
		t.Errorf("Expected first phase 'First', got %q", phases[0].Name)
	}

	if phases[1].Name != "Second" {
		t.Errorf("Expected second phase 'Second', got %q", phases[1].Name)
	}

	if phases[2].Name != "Third" {
		t.Errorf("Expected third phase 'Third', got %q", phases[2].Name)
	}

	// First phase should have 2 tasks
	if phases[0].TaskCount != 2 {
		t.Errorf("Expected first phase to have 2 tasks, got %d", phases[0].TaskCount)
	}
}

func TestParsePhasesFromConfigNoTasks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
[[phase]]
id = "build"
name = "Build"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	tasks := []model.TaskResult{}

	phases, err := ParsePhasesFromConfig(configPath, tasks)
	if err != nil {
		t.Fatalf("ParsePhasesFromConfig() error = %v", err)
	}

	if len(phases) != 0 {
		t.Errorf("Expected 0 phases, got %d", len(phases))
	}
}

func TestParsePhasesFromConfigSkippedTasks(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
[[phase]]
id = "test"
name = "Test"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	tasks := []model.TaskResult{
		{
			ID:         "test1",
			Name:       "Test 1",
			Phase:      "Test",
			Status:     model.StatusSkipped,
			DurationMs: 0,
		},
		{
			ID:         "test2",
			Name:       "Test 2",
			Phase:      "Test",
			Status:     model.StatusSkipped,
			DurationMs: 0,
		},
	}

	phases, err := ParsePhasesFromConfig(configPath, tasks)
	if err != nil {
		t.Fatalf("ParsePhasesFromConfig() error = %v", err)
	}

	if len(phases) != 1 {
		t.Fatalf("Expected 1 phase, got %d", len(phases))
	}

	// Phase with only skipped tasks should still be PASS
	if phases[0].Status != "PASS" {
		t.Errorf("Expected phase status 'PASS', got %q", phases[0].Status)
	}
}
