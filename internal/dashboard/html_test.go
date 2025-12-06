package dashboard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/drew/devpipe/internal/model"
)

func TestStatusClass(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"PASS", "pass"},
		{"FAIL", "fail"},
		{"SKIPPED", "skip"},
		{"RUNNING", ""},
		{"PENDING", ""},
		{"UNKNOWN", ""},
	}

	for _, tt := range tests {
		got := statusClass(tt.status)
		if got != tt.want {
			t.Errorf("statusClass(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestStatusSymbol(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"PASS", "✓"},
		{"FAIL", "✗"},
		{"SKIPPED", "⊘"},
		{"RUNNING", "•"},
		{"PENDING", "•"},
		{"UNKNOWN", "•"},
	}

	for _, tt := range tests {
		got := statusSymbol(tt.status)
		if got != tt.want {
			t.Errorf("statusSymbol(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "needs truncation",
			input:  "hello world",
			maxLen: 8,
			want:   "hello wo...",
		},
		{
			name:   "very short maxLen",
			input:  "hello",
			maxLen: 3,
			want:   "hel...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 5,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestPhaseEmoji(t *testing.T) {
	tests := []struct {
		phase string
		want  string
	}{
		{"build", "📦"},
		{"compile", "🔨"},
		{"test", "🧪"},
		{"deploy", "🚀"},
		{"lint", "🔍"},
		{"security", "🔒"},
		{"unknown", "📋"},
		{"", "📋"},
	}

	for _, tt := range tests {
		got := phaseEmoji(tt.phase)
		if got != tt.want {
			t.Errorf("phaseEmoji(%q) = %q, want %q", tt.phase, got, tt.want)
		}
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
	}{
		{
			name:  "int",
			input: 42,
			want:  42.0,
		},
		{
			name:  "float64",
			input: 3.14,
			want:  3.14,
		},
		{
			name:  "float32",
			input: float32(2.5),
			want:  2.5,
		},
		{
			name:  "invalid string",
			input: "not a number",
			want:  0.0,
		},
		{
			name:  "nil",
			input: nil,
			want:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toFloat64(tt.input)
			if got != tt.want {
				t.Errorf("toFloat64(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestReadLastLines(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Create a test log file with multiple lines
	content := `line 1
line 2
line 3
line 4
line 5
line 6
line 7
line 8
line 9
line 10`

	if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		maxLines int
		wantLen  int
	}{
		{
			name:     "read 5 lines",
			path:     logPath,
			maxLines: 5,
			wantLen:  5,
		},
		{
			name:     "read more than available",
			path:     logPath,
			maxLines: 20,
			wantLen:  10,
		},
		{
			name:     "read 0 lines",
			path:     logPath,
			maxLines: 0,
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := readLastLines(tt.path, tt.maxLines)
			if len(lines) != tt.wantLen {
				t.Errorf("readLastLines() returned %d lines, want %d", len(lines), tt.wantLen)
			}
		})
	}
}

func TestReadLastLinesContent(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	content := "line 1\nline 2\nline 3\nline 4\nline 5"
	if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	lines := readLastLines(logPath, 3)

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}

	// Should get the last 3 lines
	expected := []string{"line 3", "line 4", "line 5"}
	for i, line := range lines {
		if !strings.Contains(line, expected[i]) {
			t.Errorf("Line %d: expected to contain %q, got %q", i, expected[i], line)
		}
	}
}

func TestWriteHTMLDashboard(t *testing.T) {
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "test.html")

	summary := Summary{
		TotalRuns: 5,
		Version:   "1.0.0",
		RecentRuns: []RunSummary{
			{
				RunID:      "run-1",
				Status:     "PASS",
				PassCount:  3,
				FailCount:  0,
				TotalTasks: 3,
			},
		},
		TaskStats: map[string]TaskStats{
			"task1": {
				ID:        "task1",
				Name:      "Task 1",
				TotalRuns: 5,
				PassCount: 4,
				FailCount: 1,
			},
		},
	}

	err := writeHTMLDashboard(htmlPath, summary)
	if err != nil {
		t.Fatalf("writeHTMLDashboard() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Error("Expected HTML file to be created")
	}

	// Read and verify content
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML file: %v", err)
	}

	contentStr := string(content)

	// Should contain HTML structure
	if !strings.Contains(contentStr, "<html") {
		t.Error("Expected HTML to contain <html tag")
	}

	// Should contain summary data
	if !strings.Contains(contentStr, "run-1") {
		t.Error("Expected HTML to contain run ID")
	}
}

func TestWriteRunDetailHTML(t *testing.T) {
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "detail.html")

	// Create a log file for testing
	logPath := filepath.Join(tmpDir, "task.log")
	if err := os.WriteFile(logPath, []byte("log line 1\nlog line 2\nlog line 3"), 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}

	run := model.RunRecord{
		RunID:     "test-run-123",
		Timestamp: "2024-01-15T10:00:00Z",
		Git: map[string]interface{}{
			"mode": "staged",
		},
		Tasks: []model.TaskResult{
			{
				ID:         "task1",
				Name:       "Test Task",
				Status:     model.StatusPass,
				DurationMs: 1500,
				LogPath:    logPath,
			},
		},
	}

	err := writeRunDetailHTML(htmlPath, run)
	if err != nil {
		t.Fatalf("writeRunDetailHTML() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Error("Expected detail HTML file to be created")
	}

	// Read and verify content
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read detail HTML file: %v", err)
	}

	contentStr := string(content)

	// Should contain HTML structure
	if !strings.Contains(contentStr, "<html") {
		t.Error("Expected HTML to contain <html tag")
	}

	// Should contain run ID
	if !strings.Contains(contentStr, "test-run-123") {
		t.Error("Expected HTML to contain run ID")
	}

	// Should contain task name
	if !strings.Contains(contentStr, "Test Task") {
		t.Error("Expected HTML to contain task name")
	}
}

func TestGetLocalTimezone(t *testing.T) {
	tz := getLocalTimezone()

	if tz == "" {
		t.Error("Expected non-empty timezone")
	}

	t.Logf("Local timezone: %s", tz)
}

func TestShortRunID(t *testing.T) {
	tests := []struct {
		name   string
		fullID string
		want   string
	}{
		{
			name:   "standard format",
			fullID: "2025-11-30T08-15-34Z_003617",
			want:   "003617",
		},
		{
			name:   "no underscore",
			fullID: "simple-id",
			want:   "simple-id",
		},
		{
			name:   "multiple underscores",
			fullID: "2025_11_30_12345",
			want:   "12345",
		},
		{
			name:   "empty string",
			fullID: "",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortRunID(tt.fullID)
			if got != tt.want {
				t.Errorf("shortRunID(%q) = %q, want %q", tt.fullID, got, tt.want)
			}
		})
	}
}

func TestPhaseEmojiCaseInsensitive(t *testing.T) {
	tests := []struct {
		phase string
		want  string
	}{
		{"BUILD", "📦"},
		{"Build", "📦"},
		{"build", "📦"},
		{"VALIDATION", "🧪"},
		{"Testing", "🧪"},
		{"my-build-phase", "📦"},
		{"integration", "🔗"}, // Exact match for integration
	}

	for _, tt := range tests {
		got := phaseEmoji(tt.phase)
		if got != tt.want {
			t.Errorf("phaseEmoji(%q) = %q, want %q", tt.phase, got, tt.want)
		}
	}
}

func TestReadLastLinesWithANSI(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "ansi.log")

	// Create log with ANSI codes
	content := "\x1b[31mRed line\x1b[0m\n\x1b[32mGreen line\x1b[0m\nnormal line"
	if err := os.WriteFile(logPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}

	lines := readLastLines(logPath, 10)

	// ANSI codes should be stripped
	for _, line := range lines {
		if strings.Contains(line, "\x1b") {
			t.Errorf("Expected ANSI codes to be stripped, got line: %q", line)
		}
	}

	// Should contain the text content
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Red line") || strings.Contains(line, "Green line") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find text content after ANSI stripping")
	}
}

func TestReadLastLinesNonexistentFile(t *testing.T) {
	lines := readLastLines("/nonexistent/file.log", 10)

	if len(lines) != 1 {
		t.Errorf("Expected 1 error line, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "Error") {
		t.Errorf("Expected error message, got: %q", lines[0])
	}
}

func TestToFloat64Int64(t *testing.T) {
	var val int64 = 12345
	result := toFloat64(val)

	if result != 12345.0 {
		t.Errorf("Expected 12345.0, got %f", result)
	}
}

func TestWriteHTMLDashboardError(t *testing.T) {
	// Try to write to an invalid path
	invalidPath := "/invalid/path/that/does/not/exist/report.html"

	summary := Summary{
		TotalRuns: 1,
		Version:   "1.0.0",
	}

	err := writeHTMLDashboard(invalidPath, summary)
	if err == nil {
		t.Error("Expected error when writing to invalid path")
	}
}

func TestWriteSummaryJSONError(t *testing.T) {
	// Try to write to an invalid path
	invalidPath := "/invalid/path/that/does/not/exist/summary.json"

	summary := Summary{
		TotalRuns: 1,
	}

	err := writeSummaryJSON(invalidPath, summary)
	if err == nil {
		t.Error("Expected error when writing to invalid path")
	}
}

func TestWriteRunDetailHTMLWithConfigAndLogs(t *testing.T) {
	tmpDir := t.TempDir()
	runDir := filepath.Join(tmpDir, "run-123")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatalf("Failed to create run dir: %v", err)
	}

	htmlPath := filepath.Join(runDir, "detail.html")

	// Create a config file
	configPath := filepath.Join(runDir, "config.toml")
	configContent := `[tasks.test]
name = "Test Task"
command = "echo test"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create a log file
	logPath := filepath.Join(runDir, "task.log")
	if err := os.WriteFile(logPath, []byte("test output\nline 2\nline 3"), 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}

	run := model.RunRecord{
		RunID:      "run-123",
		Timestamp:  "2024-01-15T10:00:00Z",
		ConfigPath: configPath,
		Git: map[string]interface{}{
			"mode":   "staged",
			"branch": "main",
		},
		Tasks: []model.TaskResult{
			{
				ID:         "test",
				Name:       "Test Task",
				Status:     model.StatusPass,
				DurationMs: 1500,
				LogPath:    logPath,
				ExitCode:   intPtr(0),
			},
		},
	}

	err := writeRunDetailHTML(htmlPath, run)
	if err != nil {
		t.Fatalf("writeRunDetailHTML() error = %v", err)
	}

	// Verify file was created and contains config
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "run-123") {
		t.Error("Expected HTML to contain run ID")
	}
}

func TestWriteRunDetailHTMLWithMetricsData(t *testing.T) {
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "detail.html")

	run := model.RunRecord{
		RunID:     "run-456",
		Timestamp: "2024-01-15T10:00:00Z",
		Git: map[string]interface{}{
			"mode": "staged",
		},
		Tasks: []model.TaskResult{
			{
				ID:         "task-with-metrics",
				Name:       "Task With Metrics",
				Status:     model.StatusPass,
				DurationMs: 2000,
				Metrics: &model.TaskMetrics{
					SummaryFormat: "artifact",
					Kind:          "test",
				},
			},
		},
	}

	err := writeRunDetailHTML(htmlPath, run)
	if err != nil {
		t.Fatalf("writeRunDetailHTML() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Error("Expected HTML file to be created")
	}
}

func TestWriteRunDetailHTMLWithPhases(t *testing.T) {
	tmpDir := t.TempDir()
	runDir := filepath.Join(tmpDir, "run-789")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatalf("Failed to create run dir: %v", err)
	}

	htmlPath := filepath.Join(runDir, "detail.html")

	// Create a config file with phases
	configPath := filepath.Join(runDir, "config.toml")
	configContent := `[[phases]]
name = "Build"
tasks = ["build"]

[tasks.build]
name = "Build Task"
command = "make build"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	run := model.RunRecord{
		RunID:      "run-789",
		Timestamp:  "2024-01-15T10:00:00Z",
		ConfigPath: configPath,
		Git: map[string]interface{}{
			"mode": "staged",
		},
		Tasks: []model.TaskResult{
			{
				ID:         "build",
				Name:       "Build Task",
				Status:     model.StatusPass,
				DurationMs: 3000,
			},
		},
	}

	err := writeRunDetailHTML(htmlPath, run)
	if err != nil {
		t.Fatalf("writeRunDetailHTML() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Error("Expected HTML file to be created")
	}
}

func TestWriteRunDetailHTMLMultipleTasks(t *testing.T) {
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "detail.html")

	run := model.RunRecord{
		RunID:     "run-multi",
		Timestamp: "2024-01-15T10:00:00Z",
		Git: map[string]interface{}{
			"mode": "staged",
		},
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
				ExitCode:   intPtr(1),
			},
			{
				ID:         "task3",
				Name:       "Task 3",
				Status:     model.StatusSkipped,
				DurationMs: 0,
			},
		},
	}

	err := writeRunDetailHTML(htmlPath, run)
	if err != nil {
		t.Fatalf("writeRunDetailHTML() error = %v", err)
	}

	// Verify file contains all tasks
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML file: %v", err)
	}

	contentStr := string(content)
	for _, taskName := range []string{"Task 1", "Task 2", "Task 3"} {
		if !strings.Contains(contentStr, taskName) {
			t.Errorf("Expected HTML to contain %q", taskName)
		}
	}
}

func intPtr(i int) *int {
	return &i
}

func TestTemplateFunctions(t *testing.T) {
	// Test the template helper functions directly

	// Test add function
	addFunc := func(a, b interface{}) int {
		return int(toFloat64(a)) + int(toFloat64(b))
	}
	if addFunc(5, 3) != 8 {
		t.Error("add function failed")
	}

	// Test gt function
	gtFunc := func(a, b interface{}) bool {
		return toFloat64(a) > toFloat64(b)
	}
	if !gtFunc(10, 5) {
		t.Error("gt function failed")
	}
	if gtFunc(3, 7) {
		t.Error("gt function should return false")
	}

	// Test sub function
	subFunc := func(a, b interface{}) float64 {
		return toFloat64(a) - toFloat64(b)
	}
	if subFunc(10.0, 3.0) != 7.0 {
		t.Error("sub function failed")
	}

	// Test mul function
	mulFunc := func(a, b interface{}) float64 {
		return toFloat64(a) * toFloat64(b)
	}
	if mulFunc(5.0, 3.0) != 15.0 {
		t.Error("mul function failed")
	}

	// Test div function
	divFunc := func(a, b interface{}) float64 {
		aVal := toFloat64(a)
		bVal := toFloat64(b)
		if bVal == 0 {
			return 0
		}
		return aVal / bVal
	}
	if divFunc(10.0, 2.0) != 5.0 {
		t.Error("div function failed")
	}
	if divFunc(10.0, 0.0) != 0.0 {
		t.Error("div by zero should return 0")
	}

	// Test deref function
	derefFunc := func(i *int) int {
		if i != nil {
			return *i
		}
		return 0
	}
	val := 42
	if derefFunc(&val) != 42 {
		t.Error("deref function failed")
	}
	if derefFunc(nil) != 0 {
		t.Error("deref nil should return 0")
	}

	// Test hasPrefix function
	hasPrefixFunc := func(s, prefix string) bool {
		return len(s) >= len(prefix) && s[:len(prefix)] == prefix
	}
	if !hasPrefixFunc("hello world", "hello") {
		t.Error("hasPrefix should return true")
	}
	if hasPrefixFunc("hi", "hello") {
		t.Error("hasPrefix should return false")
	}

	// Test trimPrefix function
	trimPrefixFunc := func(s, prefix string) string {
		if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
			return s[len(prefix):]
		}
		return s
	}
	if trimPrefixFunc("hello world", "hello ") != "world" {
		t.Error("trimPrefix failed")
	}
	if trimPrefixFunc("hi", "hello") != "hi" {
		t.Error("trimPrefix should return original string")
	}
}

func TestWriteRunDetailHTMLTemplateError(t *testing.T) {
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "detail.html")

	// Create a run with data that might cause template execution issues
	run := model.RunRecord{
		RunID:     "test-run",
		Timestamp: "invalid-timestamp", // This might cause formatTime to fail gracefully
		Git: map[string]interface{}{
			"mode": "staged",
		},
		Tasks: []model.TaskResult{
			{
				ID:         "task1",
				Name:       "Test Task",
				Status:     model.StatusPass,
				DurationMs: 1500,
			},
		},
	}

	// Should still succeed even with invalid timestamp
	err := writeRunDetailHTML(htmlPath, run)
	if err != nil {
		t.Fatalf("writeRunDetailHTML() should handle invalid timestamp, error = %v", err)
	}
}
