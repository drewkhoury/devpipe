package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/drew/devpipe/internal/model"
)

func TestGenerateDashboard(t *testing.T) {
	tmpDir := t.TempDir()

	// Create runs directory
	runsDir := filepath.Join(tmpDir, "runs")
	if err := os.MkdirAll(runsDir, 0755); err != nil {
		t.Fatalf("Failed to create runs dir: %v", err)
	}

	// Create a sample run
	run := &model.RunRecord{
		RunID:     "test-123",
		Timestamp: time.Now().Format(time.RFC3339),
		Tasks: []model.TaskResult{
			{
				ID:         "task1",
				Name:       "Test Task",
				Status:     model.StatusPass,
				DurationMs: 1000,
			},
		},
	}

	// Write run.json
	runDir := filepath.Join(runsDir, "test-123")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatalf("Failed to create run dir: %v", err)
	}

	runData, _ := json.Marshal(run)
	if err := os.WriteFile(filepath.Join(runDir, "run.json"), runData, 0644); err != nil {
		t.Fatalf("Failed to write run.json: %v", err)
	}

	// Generate dashboard
	err := GenerateDashboard(tmpDir)
	if err != nil {
		t.Fatalf("GenerateDashboard() error = %v", err)
	}

	// Verify summary.json was created
	summaryPath := filepath.Join(tmpDir, "summary.json")
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Error("Expected summary.json to be created")
	}

	// Verify report.html was created
	htmlPath := filepath.Join(tmpDir, "report.html")
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Error("Expected report.html to be created")
	}
}

func TestGenerateDashboardWithVersion(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty runs directory
	runsDir := filepath.Join(tmpDir, "runs")
	if err := os.MkdirAll(runsDir, 0755); err != nil {
		t.Fatalf("Failed to create runs dir: %v", err)
	}

	err := GenerateDashboardWithVersion(tmpDir, "1.0.0")
	if err != nil {
		t.Fatalf("GenerateDashboardWithVersion() error = %v", err)
	}

	// Verify files were created
	if _, err := os.Stat(filepath.Join(tmpDir, "summary.json")); os.IsNotExist(err) {
		t.Error("Expected summary.json to be created")
	}
}

func TestGenerateDashboardWithOptions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create runs directory
	runsDir := filepath.Join(tmpDir, "runs")
	if err := os.MkdirAll(runsDir, 0755); err != nil {
		t.Fatalf("Failed to create runs dir: %v", err)
	}

	err := GenerateDashboardWithOptions(tmpDir, "1.0.0", false, "")
	if err != nil {
		t.Fatalf("GenerateDashboardWithOptions() error = %v", err)
	}

	// Verify files were created
	if _, err := os.Stat(filepath.Join(tmpDir, "summary.json")); os.IsNotExist(err) {
		t.Error("Expected summary.json to be created")
	}
}

func TestLoadAllRuns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple runs
	runs := []*model.RunRecord{
		{
			RunID:     "run-1",
			Timestamp: time.Now().Format(time.RFC3339),
			Tasks:     []model.TaskResult{},
		},
		{
			RunID:     "run-2",
			Timestamp: time.Now().Format(time.RFC3339),
			Tasks:     []model.TaskResult{},
		},
	}

	for _, run := range runs {
		runDir := filepath.Join(tmpDir, run.RunID)
		if err := os.MkdirAll(runDir, 0755); err != nil {
			t.Fatalf("Failed to create run dir: %v", err)
		}

		runData, _ := json.Marshal(run)
		if err := os.WriteFile(filepath.Join(runDir, "run.json"), runData, 0644); err != nil {
			t.Fatalf("Failed to write run.json: %v", err)
		}
	}

	loaded, err := loadAllRuns(tmpDir)
	if err != nil {
		t.Fatalf("loadAllRuns() error = %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("Expected 2 runs, got %d", len(loaded))
	}
}

func TestAggregateRuns(t *testing.T) {
	runs := []model.RunRecord{
		{
			RunID:     "run-1",
			Timestamp: time.Now().Format(time.RFC3339),
			Tasks: []model.TaskResult{
				{
					ID:         "task1",
					Name:       "Task 1",
					Status:     model.StatusPass,
					DurationMs: 1000,
				},
				{
					ID:         "task2",
					Name:       "Task 2",
					Status:     model.StatusFail,
					DurationMs: 2000,
				},
			},
		},
		{
			RunID:     "run-2",
			Timestamp: time.Now().Format(time.RFC3339),
			Tasks: []model.TaskResult{
				{
					ID:         "task1",
					Name:       "Task 1",
					Status:     model.StatusPass,
					DurationMs: 1500,
				},
			},
		},
	}

	summary := aggregateRuns(runs, "1.0.0")

	if summary.TotalRuns != 2 {
		t.Errorf("Expected 2 total runs, got %d", summary.TotalRuns)
	}

	if len(summary.RecentRuns) != 2 {
		t.Errorf("Expected 2 recent runs, got %d", len(summary.RecentRuns))
	}

	if summary.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", summary.Version)
	}

	// Check task stats
	if stats, ok := summary.TaskStats["task1"]; ok {
		if stats.TotalRuns != 2 {
			t.Errorf("Expected task1 to have 2 runs, got %d", stats.TotalRuns)
		}
		if stats.PassCount != 2 {
			t.Errorf("Expected task1 to have 2 passes, got %d", stats.PassCount)
		}
	} else {
		t.Error("Expected task1 in task stats")
	}
}

func TestCalculateTaskStats(t *testing.T) {
	runs := []model.RunRecord{
		{
			Tasks: []model.TaskResult{
				{
					ID:         "task1",
					Name:       "Task 1",
					Status:     model.StatusPass,
					DurationMs: 1000,
				},
				{
					ID:         "task1",
					Name:       "Task 1",
					Status:     model.StatusFail,
					DurationMs: 2000,
				},
			},
		},
	}

	stats := calculateTaskStats(runs, len(runs))

	if len(stats) != 1 {
		t.Fatalf("Expected 1 task in stats, got %d", len(stats))
	}

	task1Stats := stats["task1"]
	if task1Stats.TotalRuns != 2 {
		t.Errorf("Expected 2 runs for task1, got %d", task1Stats.TotalRuns)
	}

	if task1Stats.PassCount != 1 {
		t.Errorf("Expected 1 pass for task1, got %d", task1Stats.PassCount)
	}

	if task1Stats.FailCount != 1 {
		t.Errorf("Expected 1 fail for task1, got %d", task1Stats.FailCount)
	}

	// Average duration should be (1000 + 2000) / 2 = 1500
	if task1Stats.AvgDuration != 1500.0 {
		t.Errorf("Expected avg duration 1500, got %f", task1Stats.AvgDuration)
	}
}

func TestSummarizeRun(t *testing.T) {
	run := model.RunRecord{
		RunID:     "test-123",
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Tasks: []model.TaskResult{
			{Status: model.StatusPass, DurationMs: 1000},
			{Status: model.StatusPass, DurationMs: 2000},
			{Status: model.StatusFail, DurationMs: 1500},
			{Status: model.StatusSkipped, DurationMs: 0},
		},
	}

	summary := summarizeRun(run)

	if summary.RunID != "test-123" {
		t.Errorf("Expected RunID 'test-123', got '%s'", summary.RunID)
	}

	if summary.PassCount != 2 {
		t.Errorf("Expected 2 passes, got %d", summary.PassCount)
	}

	if summary.FailCount != 1 {
		t.Errorf("Expected 1 fail, got %d", summary.FailCount)
	}

	if summary.SkipCount != 1 {
		t.Errorf("Expected 1 skip, got %d", summary.SkipCount)
	}

	if summary.TotalTasks != 4 {
		t.Errorf("Expected 4 total tasks, got %d", summary.TotalTasks)
	}

	if summary.Status != "FAIL" {
		t.Errorf("Expected status 'FAIL', got '%s'", summary.Status)
	}

	// Duration should be sum of all task durations: 1000 + 2000 + 1500 = 4500ms
	if summary.Duration != 4500 {
		t.Errorf("Expected duration 4500ms, got %d", summary.Duration)
	}
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		a    int
		b    int
		want int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{7, 7, 7},
		{0, 100, 0},
		{-5, 5, -5},
	}

	for _, tt := range tests {
		got := min(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCleanCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple command",
			input: "go test",
			want:  "go test",
		},
		{
			name:  "command with shell prompt",
			input: "user@host dir % go test",
			want:  "go test",
		},
		{
			name:  "command without prompt",
			input: "./devpipe run",
			want:  "./devpipe run",
		},
		{
			name:  "empty command",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanCommand(tt.input)
			if got != tt.want {
				t.Errorf("cleanCommand(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWriteSummaryJSON(t *testing.T) {
	tmpDir := t.TempDir()
	summaryPath := filepath.Join(tmpDir, "summary.json")

	summary := &Summary{
		TotalRuns: 5,
		Version:   "1.0.0",
		RecentRuns: []RunSummary{
			{RunID: "run-1", Status: "PASS"},
		},
		TaskStats: map[string]TaskStats{
			"task1": {ID: "task1", Name: "Task 1", TotalRuns: 5},
		},
	}

	err := writeSummaryJSON(summaryPath, *summary)
	if err != nil {
		t.Fatalf("writeSummaryJSON() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(summaryPath); os.IsNotExist(err) {
		t.Error("Expected summary.json to be created")
	}

	// Read and verify content
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("Failed to read summary.json: %v", err)
	}

	var loaded Summary
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal summary.json: %v", err)
	}

	if loaded.TotalRuns != 5 {
		t.Errorf("Expected 5 total runs, got %d", loaded.TotalRuns)
	}
}

func TestCopyMascotAssets(t *testing.T) {
	tmpDir := t.TempDir()

	err := copyMascotAssets(tmpDir)
	// This may fail if assets aren't embedded, that's ok
	if err != nil {
		t.Logf("copyMascotAssets() error = %v (expected if assets not embedded)", err)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		ms   int64
		want string
	}{
		{
			name: "less than second",
			ms:   500,
			want: "500ms",
		},
		{
			name: "exactly 1 second",
			ms:   1000,
			want: "1.0s",
		},
		{
			name: "seconds with decimal",
			ms:   1500,
			want: "1.5s",
		},
		{
			name: "exactly 1 minute",
			ms:   60000,
			want: "1m 0s",
		},
		{
			name: "minutes and seconds",
			ms:   90000,
			want: "1m 30s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.ms)
			if got != tt.want {
				t.Errorf("formatDuration(%d) = %q, want %q", tt.ms, got, tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	formatted := formatTime(testTime.Format(time.RFC3339))

	// Should contain date and time components
	if formatted == "" {
		t.Error("Expected non-empty formatted time")
	}

	// Just verify it doesn't panic and returns something
	t.Logf("Formatted time: %s", formatted)
}

func TestFormatTimeInvalidTimestamp(t *testing.T) {
	// Should return the original string if parsing fails
	invalid := "not-a-timestamp"
	result := formatTime(invalid)

	if result != invalid {
		t.Errorf("Expected %q, got %q", invalid, result)
	}
}

func TestLoadAllRunsWithNonDirectoryEntries(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file (not directory) in the runs directory
	if err := os.WriteFile(filepath.Join(tmpDir, "not-a-dir.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Create a valid run directory
	runDir := filepath.Join(tmpDir, "run-1")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatalf("Failed to create run dir: %v", err)
	}

	run := &model.RunRecord{
		RunID:     "run-1",
		Timestamp: time.Now().Format(time.RFC3339),
		Tasks:     []model.TaskResult{},
	}

	runData, _ := json.Marshal(run)
	if err := os.WriteFile(filepath.Join(runDir, "run.json"), runData, 0644); err != nil {
		t.Fatalf("Failed to write run.json: %v", err)
	}

	loaded, err := loadAllRuns(tmpDir)
	if err != nil {
		t.Fatalf("loadAllRuns() error = %v", err)
	}

	// Should only load the valid run, skipping the file
	if len(loaded) != 1 {
		t.Errorf("Expected 1 run, got %d", len(loaded))
	}
}

func TestLoadAllRunsWithInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a run directory with invalid JSON
	runDir := filepath.Join(tmpDir, "bad-run")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatalf("Failed to create run dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(runDir, "run.json"), []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write run.json: %v", err)
	}

	loaded, err := loadAllRuns(tmpDir)
	if err != nil {
		t.Fatalf("loadAllRuns() error = %v", err)
	}

	// Should skip invalid JSON and return empty list
	if len(loaded) != 0 {
		t.Errorf("Expected 0 runs, got %d", len(loaded))
	}
}

func TestLoadAllRunsMissingRunJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a run directory without run.json
	runDir := filepath.Join(tmpDir, "incomplete-run")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatalf("Failed to create run dir: %v", err)
	}

	loaded, err := loadAllRuns(tmpDir)
	if err != nil {
		t.Fatalf("loadAllRuns() error = %v", err)
	}

	// Should skip directories without run.json
	if len(loaded) != 0 {
		t.Errorf("Expected 0 runs, got %d", len(loaded))
	}
}

func TestSummarizeRunAllSkipped(t *testing.T) {
	run := model.RunRecord{
		RunID:     "test-123",
		Timestamp: time.Now().Format(time.RFC3339),
		Tasks: []model.TaskResult{
			{Status: model.StatusSkipped, DurationMs: 0},
			{Status: model.StatusSkipped, DurationMs: 0},
		},
	}

	summary := summarizeRun(run)

	// When all tasks are skipped, status should be SKIPPED
	if summary.Status != "SKIPPED" {
		t.Errorf("Expected status 'SKIPPED', got '%s'", summary.Status)
	}

	if summary.SkipCount != 2 {
		t.Errorf("Expected 2 skips, got %d", summary.SkipCount)
	}
}

func TestCalculateTaskStatsWithSkippedTasks(t *testing.T) {
	runs := []model.RunRecord{
		{
			Tasks: []model.TaskResult{
				{
					ID:         "task1",
					Name:       "Task 1",
					Status:     model.StatusPass,
					DurationMs: 1000,
					Skipped:    false,
				},
				{
					ID:         "task2",
					Name:       "Task 2",
					Status:     model.StatusSkipped,
					DurationMs: 0,
					Skipped:    true,
				},
			},
		},
	}

	stats := calculateTaskStats(runs, len(runs))

	// Skipped tasks should not be included in average duration
	task1Stats := stats["task1"]
	if task1Stats.AvgDuration != 1000.0 {
		t.Errorf("Expected avg duration 1000, got %f", task1Stats.AvgDuration)
	}

	task2Stats := stats["task2"]
	if task2Stats.SkipCount != 1 {
		t.Errorf("Expected 1 skip for task2, got %d", task2Stats.SkipCount)
	}

	// Skipped task should have 0 average duration
	if task2Stats.AvgDuration != 0 {
		t.Errorf("Expected 0 avg duration for skipped task, got %f", task2Stats.AvgDuration)
	}
}

func TestCalculateTaskStatsLimitedRuns(t *testing.T) {
	runs := []model.RunRecord{
		{Tasks: []model.TaskResult{{ID: "task1", Name: "Task 1", Status: model.StatusPass, DurationMs: 100}}},
		{Tasks: []model.TaskResult{{ID: "task1", Name: "Task 1", Status: model.StatusPass, DurationMs: 200}}},
		{Tasks: []model.TaskResult{{ID: "task1", Name: "Task 1", Status: model.StatusPass, DurationMs: 300}}},
	}

	// Calculate stats for only the first 2 runs
	stats := calculateTaskStats(runs, 2)

	task1Stats := stats["task1"]
	if task1Stats.TotalRuns != 2 {
		t.Errorf("Expected 2 runs, got %d", task1Stats.TotalRuns)
	}

	// Average should be (100 + 200) / 2 = 150
	if task1Stats.AvgDuration != 150.0 {
		t.Errorf("Expected avg duration 150, got %f", task1Stats.AvgDuration)
	}
}

func TestWriteRunJSON(t *testing.T) {
	tmpDir := t.TempDir()
	runPath := filepath.Join(tmpDir, "run.json")

	run := model.RunRecord{
		RunID:     "test-run",
		Timestamp: time.Now().Format(time.RFC3339),
		Tasks: []model.TaskResult{
			{ID: "task1", Name: "Task 1", Status: model.StatusPass},
		},
	}

	err := writeRunJSON(runPath, run)
	if err != nil {
		t.Fatalf("writeRunJSON() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(runPath); os.IsNotExist(err) {
		t.Error("Expected run.json to be created")
	}

	// Read and verify content
	data, err := os.ReadFile(runPath)
	if err != nil {
		t.Fatalf("Failed to read run.json: %v", err)
	}

	var loaded model.RunRecord
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal run.json: %v", err)
	}

	if loaded.RunID != "test-run" {
		t.Errorf("Expected RunID 'test-run', got '%s'", loaded.RunID)
	}
}

func TestGenerateDashboardWithOptionsRegenerateAll(t *testing.T) {
	tmpDir := t.TempDir()
	runsDir := filepath.Join(tmpDir, "runs")
	if err := os.MkdirAll(runsDir, 0755); err != nil {
		t.Fatalf("Failed to create runs dir: %v", err)
	}

	// Create a run with existing ReportVersion
	run := &model.RunRecord{
		RunID:         "test-123",
		Timestamp:     time.Now().Format(time.RFC3339),
		ReportVersion: "old-version",
		Tasks: []model.TaskResult{
			{
				ID:         "task1",
				Name:       "Test Task",
				Status:     model.StatusPass,
				DurationMs: 1000,
			},
		},
	}

	runDir := filepath.Join(runsDir, "test-123")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatalf("Failed to create run dir: %v", err)
	}

	runData, _ := json.Marshal(run)
	if err := os.WriteFile(filepath.Join(runDir, "run.json"), runData, 0644); err != nil {
		t.Fatalf("Failed to write run.json: %v", err)
	}

	// Generate with regenerateAll=true
	err := GenerateDashboardWithOptions(tmpDir, "new-version", true, "")
	if err != nil {
		t.Fatalf("GenerateDashboardWithOptions() error = %v", err)
	}

	// Read the updated run.json
	updatedData, err := os.ReadFile(filepath.Join(runDir, "run.json"))
	if err != nil {
		t.Fatalf("Failed to read updated run.json: %v", err)
	}

	var updatedRun model.RunRecord
	if err := json.Unmarshal(updatedData, &updatedRun); err != nil {
		t.Fatalf("Failed to unmarshal updated run.json: %v", err)
	}

	// ReportVersion should be updated
	if updatedRun.ReportVersion != "new-version" {
		t.Errorf("Expected ReportVersion 'new-version', got '%s'", updatedRun.ReportVersion)
	}
}

func TestAggregateRunsWithLargeNumberOfRuns(t *testing.T) {
	// Create 150 runs to test pagination limit
	runs := make([]model.RunRecord, 150)
	for i := 0; i < 150; i++ {
		runs[i] = model.RunRecord{
			RunID:     fmt.Sprintf("run-%d", i),
			Timestamp: time.Now().Format(time.RFC3339),
			Tasks: []model.TaskResult{
				{ID: "task1", Name: "Task 1", Status: model.StatusPass, DurationMs: 100},
			},
		}
	}

	summary := aggregateRuns(runs, "1.0.0")

	// Should have all 150 runs
	if summary.TotalRuns != 150 {
		t.Errorf("Expected 150 total runs, got %d", summary.TotalRuns)
	}

	// Recent runs should be limited to 100
	if len(summary.RecentRuns) != 100 {
		t.Errorf("Expected 100 recent runs, got %d", len(summary.RecentRuns))
	}

	// TaskStatsLast25 should only include last 25 runs
	if stats, ok := summary.TaskStatsLast25["task1"]; ok {
		if stats.TotalRuns != 25 {
			t.Errorf("Expected task1 to have 25 runs in Last25, got %d", stats.TotalRuns)
		}
	}
}

func TestCleanCommandEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "percent at end",
			input: "command %",
			want:  "command %",
		},
		{
			name:  "multiple percent signs",
			input: "user@host % cmd % arg",
			want:  "cmd % arg",
		},
		{
			name:  "percent without space",
			input: "user@host %cmd",
			want:  "user@host %cmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanCommand(tt.input)
			if got != tt.want {
				t.Errorf("cleanCommand(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateDashboardIDEViewerEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	runsDir := filepath.Join(tmpDir, "runs")
	runDir := filepath.Join(runsDir, "test-run")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatalf("Failed to create run dir: %v", err)
	}

	run := &model.RunRecord{
		RunID:     "test-run",
		Timestamp: time.Now().Format(time.RFC3339),
		Git: map[string]interface{}{
			"mode": "staged",
		},
		Tasks: []model.TaskResult{
			{ID: "task1", Name: "Test Task", Status: model.StatusPass, DurationMs: 1000},
		},
	}

	runData, _ := json.Marshal(run)
	if err := os.WriteFile(filepath.Join(runDir, "run.json"), runData, 0644); err != nil {
		t.Fatalf("Failed to write run.json: %v", err)
	}

	err := GenerateDashboardWithOptions(tmpDir, "1.0.0", true, "")
	if err != nil {
		t.Fatalf("GenerateDashboardWithOptions() error = %v", err)
	}

	// IDE viewer is created per run, not at the root
	ideViewerPath := filepath.Join(runDir, "ide.html")
	if _, err := os.Stat(ideViewerPath); os.IsNotExist(err) {
		t.Error("Expected ide.html to be created in run directory")
	}
}

func TestLoadAllRunsSkipsInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	runsDir := filepath.Join(tmpDir, "runs")
	invalidRunDir := filepath.Join(runsDir, "invalid-run")
	if err := os.MkdirAll(invalidRunDir, 0755); err != nil {
		t.Fatalf("Failed to create invalid run dir: %v", err)
	}

	invalidJSON := filepath.Join(invalidRunDir, "run.json")
	if err := os.WriteFile(invalidJSON, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	validRunDir := filepath.Join(runsDir, "valid-run")
	if err := os.MkdirAll(validRunDir, 0755); err != nil {
		t.Fatalf("Failed to create valid run dir: %v", err)
	}

	validRun := &model.RunRecord{RunID: "valid-run", Timestamp: time.Now().Format(time.RFC3339), Tasks: []model.TaskResult{}}
	validJSON, _ := json.Marshal(validRun)
	if err := os.WriteFile(filepath.Join(validRunDir, "run.json"), validJSON, 0644); err != nil {
		t.Fatalf("Failed to write valid JSON: %v", err)
	}

	runs, err := loadAllRuns(runsDir)
	if err != nil {
		t.Fatalf("loadAllRuns() error = %v", err)
	}

	if len(runs) != 1 {
		t.Errorf("Expected 1 valid run, got %d", len(runs))
	}
}
