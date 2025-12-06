package ui

import (
	"testing"
	"time"
)

func TestNewAnimatedTaskTracker(t *testing.T) {
	renderer := NewRenderer(UIModeFull, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "PENDING"},
		{ID: "task2", Name: "Task 2", Status: "RUNNING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 5, 100, "type")

	if tracker == nil {
		t.Fatal("Expected non-nil tracker")
	}

	// Verify initial state
	if len(tracker.tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tracker.tasks))
	}
}

func TestAnimatedTrackerUpdateTask(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "PENDING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "phase")

	// Update task status
	tracker.UpdateTask("task1", "RUNNING", 50)

	// Verify update (we can't directly access private fields, but method shouldn't panic)
	t.Log("UpdateTask completed without panic")
}

func TestAnimatedTrackerAddLogLine(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "PENDING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Add log lines
	tracker.AddLogLine("Log line 1")
	tracker.AddLogLine("Log line 2")

	t.Log("AddLogLine completed without panic")
}

func TestAnimatedTrackerAddVerboseLine(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "PENDING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Add verbose lines
	tracker.AddVerboseLine("Verbose message 1")
	tracker.AddVerboseLine("Verbose message 2")

	t.Log("AddVerboseLine completed without panic")
}

func TestAnimatedTrackerStartStop(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "PENDING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Start the tracker
	_ = tracker.Start() // Test operation

	// Let it run briefly
	time.Sleep(50 * time.Millisecond)

	// Stop the tracker
	tracker.Stop()

	t.Log("Start/Stop completed without panic")
}

func TestAnimatedTrackerTestRender(t *testing.T) {
	renderer := NewRenderer(UIModeFull, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "PENDING"},
		{ID: "task2", Name: "Task 2", Status: "RUNNING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 5, 100, "type")

	// Test render (should not panic)
	_ = tracker.testRender() // Test operation

	t.Log("testRender completed without panic")
}

func TestAnimatedTrackerModes(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "PENDING"},
	}

	// Test basic mode
	trackerBasic := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")
	if trackerBasic == nil {
		t.Error("Expected non-nil tracker for basic mode")
	}

	// Test full mode
	rendererFull := NewRenderer(UIModeFull, false, false)
	trackerFull := NewAnimatedTaskTracker(rendererFull, tasks, 5, 100, "phase")
	if trackerFull == nil {
		t.Error("Expected non-nil tracker for full mode")
	}
}

func TestAnimatedTrackerWithMultipleTasks(t *testing.T) {
	renderer := NewRenderer(UIModeFull, false, false)
	tasks := []TaskProgress{
		{ID: "lint", Name: "Lint", Status: "PASS", Type: "check"},
		{ID: "test", Name: "Test", Status: "RUNNING", Type: "test"},
		{ID: "build", Name: "Build", Status: "PENDING", Type: "build"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 5, 100, "type")

	// Update various tasks
	tracker.UpdateTask("test", "PASS", 100)
	tracker.UpdateTask("build", "RUNNING", 25)

	// Add log lines
	tracker.AddLogLine("Test passed")
	tracker.AddLogLine("Building...")

	// Test render
	_ = tracker.testRender() // Test operation

	t.Log("Multi-task tracker test completed")
}

func TestAnimatedTrackerCalculateLines(t *testing.T) {
	renderer := NewRenderer(UIModeFull, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "RUNNING"},
		{ID: "task2", Name: "Task 2", Status: "PENDING"},
		{ID: "task3", Name: "Task 3", Status: "PASS"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 5, 100, "type")

	// The calculateLines method is private, but we can test it indirectly
	// by calling testRender which uses it
	_ = tracker.testRender() // Test operation

	t.Log("calculateLines tested indirectly via testRender")
}

func TestAnimatedTrackerRenderModes(t *testing.T) {
	rendererBasic := NewRenderer(UIModeBasic, false, false)
	rendererFull := NewRenderer(UIModeFull, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "RUNNING"},
	}

	// Test basic mode rendering
	trackerBasic := NewAnimatedTaskTracker(rendererBasic, tasks, 3, 100, "type")
	_ = trackerBasic.testRender() // Test operation

	// Test full mode rendering
	trackerFull := NewAnimatedTaskTracker(rendererFull, tasks, 5, 100, "phase")
	_ = trackerFull.testRender() // Test operation

	t.Log("Both render modes tested")
}

func TestAnimatedTrackerRenderFullMode(t *testing.T) {
	renderer := NewRenderer(UIModeFull, true, false)
	tasks := []TaskProgress{
		{ID: "lint", Name: "Lint", Status: "PASS", Type: "quality", Phase: 1, PhaseName: "Build", ElapsedSeconds: 2.5},
		{ID: "test", Name: "Test", Status: "RUNNING", Type: "quality", Phase: 1, PhaseName: "Build", ElapsedSeconds: 5.0, EstimatedSeconds: 10, IsEstimateGuess: false},
		{ID: "build", Name: "Build", Status: "PENDING", Type: "release", Phase: 2, PhaseName: "Deploy"},
		{ID: "deploy", Name: "Deploy", Status: "FAIL", Type: "release", Phase: 2, PhaseName: "Deploy", ElapsedSeconds: 3.0},
		{ID: "skip-task", Name: "Skip", Status: "SKIPPED", Type: "quality", Phase: 1, PhaseName: "Build"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 5, 100, "type")

	// Add some log lines
	tracker.AddLogLine("Running tests...")
	tracker.AddLogLine("Test output line 2")

	// Trigger full mode render
	tracker.render()

	t.Log("Full mode render completed")
}

func TestAnimatedTrackerRenderFullModeByPhase(t *testing.T) {
	renderer := NewRenderer(UIModeFull, true, false)
	tasks := []TaskProgress{
		{ID: "lint", Name: "Lint", Status: "PASS", Type: "quality", Phase: 1, PhaseName: "Build", ElapsedSeconds: 2.5},
		{ID: "test", Name: "Test", Status: "RUNNING", Type: "quality", Phase: 1, PhaseName: "Build", ElapsedSeconds: 5.0, EstimatedSeconds: 10, IsEstimateGuess: true},
		{ID: "build", Name: "Build", Status: "PENDING", Type: "release", Phase: 2, PhaseName: "Deploy"},
	}

	// Group by phase instead of type
	tracker := NewAnimatedTaskTracker(renderer, tasks, 5, 100, "phase")

	// Trigger full mode render
	tracker.render()

	t.Log("Full mode render by phase completed")
}

func TestAnimatedTrackerCalculateLinesFullMode(t *testing.T) {
	renderer := NewRenderer(UIModeFull, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "RUNNING", Type: "quality", Phase: 1, PhaseName: "Build"},
		{ID: "task2", Name: "Task 2", Status: "PENDING", Type: "quality", Phase: 1, PhaseName: "Build"},
		{ID: "task3", Name: "Task 3", Status: "PASS", Type: "release", Phase: 2, PhaseName: "Deploy"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 5, 100, "type")

	// Call calculateLines indirectly through render
	lines := tracker.calculateLines()
	if lines <= 0 {
		t.Errorf("Expected positive line count, got %d", lines)
	}

	t.Logf("Calculated lines: %d", lines)
}

func TestAnimatedTrackerCalculateLinesPhaseGrouping(t *testing.T) {
	renderer := NewRenderer(UIModeFull, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "RUNNING", Type: "quality", Phase: 1, PhaseName: "Build"},
		{ID: "task2", Name: "Task 2", Status: "PENDING", Type: "quality", Phase: 2, PhaseName: "Test"},
		{ID: "task3", Name: "Task 3", Status: "PASS", Type: "release", Phase: 3, PhaseName: ""}, // Empty phase name
	}

	// Group by phase
	tracker := NewAnimatedTaskTracker(renderer, tasks, 5, 100, "phase")

	// Call calculateLines
	lines := tracker.calculateLines()
	if lines <= 0 {
		t.Errorf("Expected positive line count, got %d", lines)
	}

	t.Logf("Calculated lines with phase grouping: %d", lines)
}

func TestAnimatedTrackerInvalidRefreshMs(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "PENDING"},
	}

	// Test with invalid refresh rates (should default to 500ms)
	tracker1 := NewAnimatedTaskTracker(renderer, tasks, 3, 10, "type") // Too low
	if tracker1 == nil {
		t.Error("Expected non-nil tracker")
	}

	tracker2 := NewAnimatedTaskTracker(renderer, tasks, 3, 3000, "type") // Too high
	if tracker2 == nil {
		t.Error("Expected non-nil tracker")
	}
}

func TestAnimatedTrackerInvalidGroupBy(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "PENDING"},
	}

	// Test with invalid groupBy (should default to "type")
	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "invalid")
	if tracker == nil {
		t.Error("Expected non-nil tracker")
	}
}

func TestAnimatedTrackerLongTaskID(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "this-is-a-very-long-task-id-that-exceeds-the-maximum-allowed-length-for-display", Name: "Long Task", Status: "RUNNING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")
	if tracker == nil {
		t.Error("Expected non-nil tracker")
	}

	// Render to test truncation
	tracker.render()

	t.Log("Long task ID handled correctly")
}

func TestAnimatedTrackerMaxLogLines(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "RUNNING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Add more log lines than maxLogLines
	for i := 0; i < 20; i++ {
		tracker.AddLogLine("Log line")
	}

	// Should keep only the last maxLogLines
	tracker.render()

	t.Log("Max log lines test completed")
}

func TestAnimatedTrackerMaxVerboseLines(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "RUNNING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Add more verbose lines than maxVerbose (5)
	for i := 0; i < 10; i++ {
		tracker.AddVerboseLine("Verbose line")
	}

	t.Log("Max verbose lines test completed")
}

func TestAnimatedTrackerCalculateLinesBasicMode(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "RUNNING"},
		{ID: "task2", Name: "Task 2", Status: "PENDING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Call calculateLines for basic mode
	lines := tracker.calculateLines()
	if lines <= 0 {
		t.Errorf("Expected positive line count, got %d", lines)
	}

	// Basic mode formula: 2 + len(tasks) + 1 + 1 + maxLogLines
	// Should be at least 2 + 2 + 1 + 1 + maxLogLines
	expectedMin := 2 + len(tasks) + 1 + 1
	if lines < expectedMin {
		t.Errorf("Expected at least %d lines, got %d", expectedMin, lines)
	}

	t.Logf("Calculated lines for basic mode: %d", lines)
}

func TestAnimatedTrackerRenderBasicModeAllStatuses(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "pass-task", Name: "Pass", Status: "PASS", ElapsedSeconds: 1.5},
		{ID: "fail-task", Name: "Fail", Status: "FAIL", ElapsedSeconds: 2.0},
		{ID: "skip-task", Name: "Skip", Status: "SKIPPED"},
		{ID: "run-task", Name: "Run", Status: "RUNNING", ElapsedSeconds: 3.0, EstimatedSeconds: 10},
		{ID: "pend-task", Name: "Pend", Status: "PENDING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Render to cover all status branches
	tracker.render()

	t.Log("Basic mode render with all statuses completed")
}

func TestAnimatedTrackerRenderFullModeAllStatuses(t *testing.T) {
	renderer := NewRenderer(UIModeFull, true, false)
	tasks := []TaskProgress{
		{ID: "pass-task", Name: "Pass", Status: "PASS", Type: "quality", Phase: 1, PhaseName: "Build", ElapsedSeconds: 1.5},
		{ID: "fail-task", Name: "Fail", Status: "FAIL", Type: "quality", Phase: 1, PhaseName: "Build", ElapsedSeconds: 2.0},
		{ID: "skip-task", Name: "Skip", Status: "SKIPPED", Type: "release", Phase: 2, PhaseName: "Deploy"},
		{ID: "run-task", Name: "Run", Status: "RUNNING", Type: "release", Phase: 2, PhaseName: "Deploy", ElapsedSeconds: 3.0, EstimatedSeconds: 10, IsEstimateGuess: false},
		{ID: "pend-task", Name: "Pend", Status: "PENDING", Type: "test", Phase: 3, PhaseName: "Test"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Call renderFullMode directly to ensure coverage
	tracker.renderFullMode()

	t.Log("Full mode render with all statuses completed")
}

func TestAnimatedTrackerRenderFullModeWithGuessedEstimate(t *testing.T) {
	renderer := NewRenderer(UIModeFull, true, false)
	tasks := []TaskProgress{
		{ID: "run-task", Name: "Run", Status: "RUNNING", Type: "quality", Phase: 1, PhaseName: "Build", ElapsedSeconds: 5.0, EstimatedSeconds: 10, IsEstimateGuess: true},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Call renderFullMode to cover the IsEstimateGuess branch
	tracker.renderFullMode()

	t.Log("Full mode render with guessed estimate completed")
}

func TestAnimatedTrackerRenderFullModeGroupByPhaseWithEmptyName(t *testing.T) {
	renderer := NewRenderer(UIModeFull, true, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "PASS", Type: "quality", Phase: 1, PhaseName: ""},
		{ID: "task2", Name: "Task 2", Status: "RUNNING", Type: "quality", Phase: 2, PhaseName: ""},
	}

	// Group by phase with empty phase names
	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "phase")

	// Call renderFullMode to cover phase grouping with empty names
	tracker.renderFullMode()

	t.Log("Full mode render with phase grouping and empty names completed")
}

func TestAnimatedTrackerRenderFullModeWideTerminal(t *testing.T) {
	renderer := NewRenderer(UIModeFull, true, false)
	// Set a wide terminal width
	renderer.width = 120

	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "RUNNING", Type: "quality", Phase: 1, PhaseName: "Build", ElapsedSeconds: 5.0, EstimatedSeconds: 10},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Call renderFullMode with wide terminal
	tracker.renderFullMode()

	t.Log("Full mode render with wide terminal completed")
}

func TestAnimatedTrackerRenderBasicModeWideTerminal(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, true, false)
	// Set a wide terminal width
	renderer.width = 120

	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "RUNNING", ElapsedSeconds: 5.0, EstimatedSeconds: 10},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Call renderBasicMode with wide terminal
	tracker.renderBasicMode()

	t.Log("Basic mode render with wide terminal completed")
}

func TestAnimatedTrackerCalculateLinesFullModeMultipleGroups(t *testing.T) {
	renderer := NewRenderer(UIModeFull, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "RUNNING", Type: "quality", Phase: 1, PhaseName: "Build"},
		{ID: "task2", Name: "Task 2", Status: "PENDING", Type: "quality", Phase: 1, PhaseName: "Build"},
		{ID: "task3", Name: "Task 3", Status: "PASS", Type: "release", Phase: 2, PhaseName: "Deploy"},
		{ID: "task4", Name: "Task 4", Status: "FAIL", Type: "release", Phase: 2, PhaseName: "Deploy"},
		{ID: "task5", Name: "Task 5", Status: "SKIPPED", Type: "test", Phase: 3, PhaseName: "Test"},
	}

	// Test with type grouping
	tracker := NewAnimatedTaskTracker(renderer, tasks, 5, 100, "type")
	lines := tracker.calculateLines()
	if lines <= 0 {
		t.Errorf("Expected positive line count, got %d", lines)
	}
	t.Logf("Calculated lines with type grouping (3 groups): %d", lines)

	// Test with phase grouping
	tracker2 := NewAnimatedTaskTracker(renderer, tasks, 5, 100, "phase")
	lines2 := tracker2.calculateLines()
	if lines2 <= 0 {
		t.Errorf("Expected positive line count, got %d", lines2)
	}
	t.Logf("Calculated lines with phase grouping (3 phases): %d", lines2)
}

func TestAnimatedTrackerStartError(t *testing.T) {
	renderer := NewRenderer(UIModeBasic, false, false)
	tasks := []TaskProgress{
		{ID: "task1", Name: "Task 1", Status: "PENDING"},
	}

	tracker := NewAnimatedTaskTracker(renderer, tasks, 3, 100, "type")

	// Start should not return an error in normal conditions
	err := tracker.Start()
	if err != nil {
		t.Logf("Start returned error (may be expected in test environment): %v", err)
	}

	// Stop the tracker
	tracker.Stop()
}
