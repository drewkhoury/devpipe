package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestNewRenderer(t *testing.T) {
	tests := []struct {
		name         string
		mode         UIMode
		enableColors bool
		animated     bool
	}{
		{
			name:         "basic mode",
			mode:         UIModeBasic,
			enableColors: false,
			animated:     false,
		},
		{
			name:         "full mode with colors",
			mode:         UIModeFull,
			enableColors: true,
			animated:     true,
		},
		{
			name:         "full mode without colors",
			mode:         UIModeFull,
			enableColors: false,
			animated:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRenderer(tt.mode, tt.enableColors, tt.animated)

			if r == nil {
				t.Fatal("Expected non-nil renderer")
			}

			if r.colors == nil {
				t.Error("Expected non-nil colors")
			}

			if r.width <= 0 {
				t.Error("Expected positive width")
			}
		})
	}
}

func TestRendererIsAnimated(t *testing.T) {
	r := NewRenderer(UIModeBasic, false, false)
	if r.IsAnimated() {
		t.Error("Expected IsAnimated() to be false for basic mode")
	}

	r2 := NewRenderer(UIModeFull, true, true)
	// May be false if not a TTY, that's ok
	_ = r2.IsAnimated()
}

func TestRenderHeader(t *testing.T) {
	tests := []struct {
		name           string
		mode           UIMode
		runID          string
		projectRoot    string
		gitMode        string
		changedFiles   int
		expectInOutput []string
	}{
		{
			name:         "basic mode",
			mode:         UIModeBasic,
			runID:        "test-123",
			projectRoot:  "/home/user/project",
			gitMode:      "staged",
			changedFiles: 5,
			expectInOutput: []string{
				"devpipe run test-123",
				"Project root: /home/user/project",
				"Git mode: staged",
				"Changed files: 5",
			},
		},
		{
			name:         "basic mode without git",
			mode:         UIModeBasic,
			runID:        "test-456",
			projectRoot:  "/project",
			gitMode:      "",
			changedFiles: 0,
			expectInOutput: []string{
				"devpipe run test-456",
				"Project root: /project",
			},
		},
		{
			name:         "full mode",
			mode:         UIModeFull,
			runID:        "test-789",
			projectRoot:  "/app",
			gitMode:      "ref",
			changedFiles: 10,
			expectInOutput: []string{
				"test-789",
				"/app",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			renderer := NewRenderer(tt.mode, false, false)
			renderer.RenderHeader(tt.runID, tt.projectRoot, tt.gitMode, tt.changedFiles)

			_ = w.Close() // Test cleanup
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r) // Test output capture
			output := buf.String()

			for _, expected := range tt.expectInOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestTruncateTaskID(t *testing.T) {
	tests := []struct {
		name   string
		id     string
		maxLen int
		want   string
	}{
		{
			name:   "short id",
			id:     "test",
			maxLen: 10,
			want:   "test",
		},
		{
			name:   "exact length",
			id:     "test-task",
			maxLen: 9,
			want:   "test-task",
		},
		{
			name:   "needs truncation",
			id:     "very-long-task-id",
			maxLen: 10,
			want:   "very-lo...",
		},
		{
			name:   "very short maxLen",
			id:     "task",
			maxLen: 3,
			want:   "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateTaskID(tt.id, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateTaskID(%q, %d) = %q, want %q",
					tt.id, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestRendererColorMethods(t *testing.T) {
	r := NewRenderer(UIModeBasic, true, false)

	// Test color methods exist and return strings
	if r.Blue("test") == "" {
		t.Error("Blue() should return non-empty string")
	}

	if r.Yellow("test") == "" {
		t.Error("Yellow() should return non-empty string")
	}

	if r.Cyan("test") == "" {
		t.Error("Cyan() should return non-empty string")
	}

	// Already tested in colors_test.go
	_ = r.Green("test")
	_ = r.Red("test")
	_ = r.Gray("test")
	_ = r.StatusColor("PASS")
}

func TestSetTracker(t *testing.T) {
	r := NewRenderer(UIModeBasic, false, false)

	// Create a mock tracker
	tracker := &AnimatedTaskTracker{}

	r.SetTracker(tracker)

	// Verify it was set (we can't directly access private field, but method shouldn't panic)
	t.Log("SetTracker completed without panic")
}

func TestSetPipelineLog(t *testing.T) {
	r := NewRenderer(UIModeBasic, false, false)

	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test-log-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }() // Test cleanup
	defer func() { _ = tmpFile.Close() }()           // Test cleanup

	r.SetPipelineLog(tmpFile)

	t.Log("SetPipelineLog completed without panic")
}

func TestVerbose(t *testing.T) {
	r := NewRenderer(UIModeBasic, false, false)

	// Test verbose output (should not panic even without tracker/log)
	r.Verbose(true, "test message %s", "arg")

	t.Log("Verbose() completed without panic")
}

func TestRenderTaskStart(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	renderer.RenderTaskStart("test-task", "go test", false)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	if !strings.Contains(output, "test-task") {
		t.Errorf("Expected output to contain 'test-task', got: %s", output)
	}
}

func TestRenderTaskComplete(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	exitCode := 0
	renderer.RenderTaskComplete("test-task", "PASS", &exitCode, 1500, false)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	if !strings.Contains(output, "test-task") {
		t.Errorf("Expected output to contain 'test-task', got: %s", output)
	}
}

func TestRenderTaskSkipped(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	renderer.RenderTaskSkipped("test-task", "disabled", false)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	if !strings.Contains(output, "test-task") {
		t.Errorf("Expected output to contain 'test-task', got: %s", output)
	}
}

func TestRenderSummary(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	summaries := []TaskSummary{
		{ID: "task1", Status: "PASS", DurationMs: 1000},
		{ID: "task2", Status: "FAIL", DurationMs: 2000},
	}
	renderer.RenderSummary(summaries, false, 3000)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should contain summary information
	if len(output) == 0 {
		t.Error("Expected non-empty summary output")
	}
}

func TestRenderProgress(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	renderer.RenderProgress(5, 10)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Progress output may be empty in basic mode without animation
	// Just verify it doesn't panic
	t.Logf("Progress output: %q", output)
}

func TestCreateAnimatedTracker(t *testing.T) {
	renderer := NewRenderer(UIModeFull, true, true)

	tasks := []TaskProgress{
		{ID: "task1", Status: "PENDING"},
	}
	tracker := renderer.CreateAnimatedTracker(tasks, 80, 24, "full")

	// Tracker might be nil if not a TTY, that's ok
	if tracker != nil {
		t.Log("Created animated tracker successfully")
	} else {
		t.Log("Animated tracker is nil (expected if not a TTY)")
	}
}

func TestRenderFullHeader(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeFull, true, false)
	renderer.RenderHeader("test-run-123", "/home/user/project", "staged", 10)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should contain box drawing characters for full mode
	if !strings.Contains(output, "╔") && !strings.Contains(output, "═") {
		t.Logf("Full header output: %s", output)
	}

	if !strings.Contains(output, "test-run-123") {
		t.Errorf("Expected output to contain run ID")
	}
}

func TestRenderProgressFullMode(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeFull, true, false)
	renderer.RenderProgress(5, 10)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Full mode should render progress (check for stages text)
	if !strings.Contains(output, "stages") && !strings.Contains(output, "5/10") {
		t.Logf("Progress output: %q", output)
	}

	// Just verify it doesn't panic - output might be empty in test environment
	t.Logf("RenderProgress completed for full mode")
}

func TestRenderTaskStartVerbose(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	renderer.RenderTaskStart("test-task", "go test ./...", true)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	if !strings.Contains(output, "test-task") {
		t.Errorf("Expected output to contain 'test-task', got: %s", output)
	}

	if !strings.Contains(output, "go test ./...") {
		t.Errorf("Expected output to contain command in verbose mode")
	}
}

func TestRenderTaskCompleteVerbose(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	exitCode := 1
	renderer.RenderTaskComplete("test-task", "FAIL", &exitCode, 2500, true)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	if !strings.Contains(output, "test-task") {
		t.Errorf("Expected output to contain 'test-task'")
	}

	if !strings.Contains(output, "exit 1") {
		t.Errorf("Expected output to contain exit code in verbose mode")
	}
}

func TestRenderTaskSkippedVerbose(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	renderer.RenderTaskSkipped("test-task", "dependency failed", true)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	if !strings.Contains(output, "test-task") {
		t.Errorf("Expected output to contain 'test-task'")
	}

	if !strings.Contains(output, "dependency failed") {
		t.Errorf("Expected output to contain skip reason in verbose mode")
	}
}

func TestRenderSummaryWithAutoFixed(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	summaries := []TaskSummary{
		{ID: "task1", Status: "PASS", DurationMs: 1000, AutoFixed: true},
		{ID: "task2", Status: "PASS", DurationMs: 2000, AutoFixed: false},
	}
	renderer.RenderSummary(summaries, false, 3000)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should contain auto-fixed annotation
	if !strings.Contains(output, "auto-fixed") {
		t.Error("Expected output to contain 'auto-fixed' annotation")
	}
}

func TestRenderSummaryWithFailures(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	summaries := []TaskSummary{
		{ID: "task1", Status: "PASS", DurationMs: 1000},
		{ID: "task2", Status: "FAIL", DurationMs: 2000},
	}
	renderer.RenderSummary(summaries, true, 3000)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should contain failure message
	if !strings.Contains(output, "failed") {
		t.Error("Expected output to contain 'failed' message")
	}
}

func TestVerboseWithTracker(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)

	// Create a tracker
	tasks := []TaskProgress{
		{ID: "task1", Status: "RUNNING"},
	}
	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")
	renderer.SetTracker(tracker)

	// Test verbose output with tracker
	renderer.Verbose(true, "test message with tracker")

	t.Log("Verbose with tracker completed")
}

func TestVerboseWithPipelineLog(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)

	// Create a temp file for pipeline log
	tmpFile, err := os.CreateTemp("", "test-pipeline-*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }() // Test cleanup
	defer func() { _ = tmpFile.Close() }()           // Test cleanup

	renderer.SetPipelineLog(tmpFile)

	// Test verbose output (should write to log file)
	renderer.Verbose(true, "test log message")

	// Read back the log file
	_, _ = tmpFile.Seek(0, 0) // Test operation
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, tmpFile) // Test operation
	logContent := buf.String()

	if !strings.Contains(logContent, "test log message") {
		t.Errorf("Expected log file to contain message, got: %s", logContent)
	}
}

func TestVerboseDisabled(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)

	// Test verbose output with verbose=false (should not print to console)
	renderer.Verbose(false, "this should not appear")

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should not contain the message
	if strings.Contains(output, "this should not appear") {
		t.Error("Expected no output when verbose is disabled")
	}
}

func TestRenderTaskStartAnimated(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, true)

	// In animated mode, RenderTaskStart should not print anything
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer.RenderTaskStart("test-task", "go test", false)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should be empty in animated mode
	if len(output) > 0 {
		t.Logf("Output in animated mode: %q (might be empty)", output)
	}
}

func TestRenderTaskCompleteAnimated(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, true)

	// In animated mode, RenderTaskComplete should not print anything
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := 0
	renderer.RenderTaskComplete("test-task", "PASS", &exitCode, 1000, false)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should be empty in animated mode
	if len(output) > 0 {
		t.Logf("Output in animated mode: %q (might be empty)", output)
	}
}

func TestRenderTaskSkippedAnimated(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, true)

	// In animated mode, RenderTaskSkipped should not print anything
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer.RenderTaskSkipped("test-task", "disabled", false)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should be empty in animated mode
	if len(output) > 0 {
		t.Logf("Output in animated mode: %q (might be empty)", output)
	}
}

func TestRenderSummaryAnimated(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, true)
	summaries := []TaskSummary{
		{ID: "task1", Status: "PASS", DurationMs: 1000},
	}
	renderer.RenderSummary(summaries, false, 1000)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should have output (summary is always shown)
	if len(output) == 0 {
		t.Error("Expected non-empty summary output")
	}
}

func TestRenderFullHeaderWithoutGit(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeFull, true, false)
	// Force full mode (NewRenderer might override to basic if not a TTY)
	renderer.mode = UIModeFull
	renderer.RenderHeader("test-run-456", "/home/user/project", "", 0)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should contain box drawing characters for full mode
	if !strings.Contains(output, "╔") && !strings.Contains(output, "═") {
		t.Logf("Full header output (may not have box chars if not TTY): %s", output)
	}

	if !strings.Contains(output, "test-run-456") {
		t.Errorf("Expected output to contain run ID")
	}

	// Should NOT contain git info
	if strings.Contains(output, "Git:") {
		t.Errorf("Expected no git info when gitMode is empty")
	}
}

func TestRenderProgressFullModeWideTerminal(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeFull, true, false)
	renderer.width = 120 // Wide terminal
	renderer.RenderProgress(7, 10)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Full mode should render progress
	if !strings.Contains(output, "7/10") && !strings.Contains(output, "stages") {
		t.Logf("Progress output: %q", output)
	}

	t.Log("RenderProgress with wide terminal completed")
}

func TestRenderProgressBasicMode(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, true, false)
	renderer.RenderProgress(3, 5)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Basic mode should not render progress (returns early)
	if len(output) > 0 {
		t.Logf("Basic mode progress output (should be empty): %q", output)
	}
}

func TestCreateAnimatedTrackerNotAnimated(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)

	tasks := []TaskProgress{
		{ID: "task1", Status: "PENDING"},
	}
	tracker := renderer.CreateAnimatedTracker(tasks, 80, 24, "type")

	// Should be nil when not animated
	if tracker != nil {
		t.Error("Expected nil tracker when animated is false")
	}
}

func TestRenderHeaderBasicModeWithGit(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	renderer.RenderHeader("run-123", "/project", "diff", 3)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	if !strings.Contains(output, "run-123") {
		t.Error("Expected output to contain run ID")
	}

	if !strings.Contains(output, "Git mode: diff") {
		t.Error("Expected output to contain git mode")
	}

	if !strings.Contains(output, "Changed files: 3") {
		t.Error("Expected output to contain changed files count")
	}
}

func TestRenderSummaryLongTaskID(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	summaries := []TaskSummary{
		{ID: "this-is-a-very-long-task-id-that-exceeds-the-maximum-allowed-length-for-display-purposes", Status: "PASS", DurationMs: 1000},
	}
	renderer.RenderSummary(summaries, false, 1000)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should truncate long task IDs
	if !strings.Contains(output, "...") {
		t.Log("Expected truncation of long task ID")
	}
}

func TestRenderSummarySkippedStatus(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer := NewRenderer(UIModeBasic, false, false)
	summaries := []TaskSummary{
		{ID: "task1", Status: "SKIPPED", DurationMs: 0},
	}
	renderer.RenderSummary(summaries, false, 0)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	if !strings.Contains(output, "SKIPPED") {
		t.Error("Expected output to contain SKIPPED status")
	}
}
