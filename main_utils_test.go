package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/drew/devpipe/internal/config"
	"github.com/drew/devpipe/internal/model"
)

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		wantLen  int
		wantLast string
	}{
		{
			name:    "empty text",
			text:    "",
			width:   80,
			wantLen: 0,
		},
		{
			name:    "short text",
			text:    "hello world",
			width:   80,
			wantLen: 1,
		},
		{
			name:     "long text needs wrapping",
			text:     "This is a very long line that should be wrapped when it exceeds the specified width limit",
			width:    40,
			wantLen:  3, // Should wrap into multiple lines
			wantLast: "limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.text, tt.width)

			if len(got) != tt.wantLen {
				t.Errorf("wrapText() returned %d lines, want %d", len(got), tt.wantLen)
			}

			if tt.wantLast != "" && len(got) > 0 {
				lastLine := got[len(got)-1]
				if !strings.Contains(lastLine, tt.wantLast) {
					t.Errorf("Last line should contain %q, got %q", tt.wantLast, lastLine)
				}
			}
		})
	}
}

func TestPrintVersion(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printVersion()

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should contain version info
	if !strings.Contains(output, "devpipe version") {
		t.Error("Expected output to contain 'devpipe version'")
	}
}

func TestPrintHelp(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printHelp()

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should contain help information
	expectedStrings := []string{
		"devpipe",
		"USAGE:",
		"FLAGS:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q", expected)
		}
	}
}

func TestFilterTasks_InvalidOnlyExits(t *testing.T) {
	if os.Getenv("DEVPIPE_TEST_INVALID_ONLY") == "1" {
		tasks := []model.TaskDefinition{{ID: "task1"}}
		// This should call os.Exit(1) inside filterTasks
		_ = filterTasks(tasks, "does-not-exist", sliceFlag{}, false, 0, false)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFilterTasks_InvalidOnlyExits")
	cmd.Env = append(os.Environ(), "DEVPIPE_TEST_INVALID_ONLY=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected non-zero exit code, got nil error")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 0 {
			t.Fatalf("expected non-zero exit code, got 0")
		}
	} else {
		t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
	}
}

func TestFilterTasks(t *testing.T) {
	tasks := []model.TaskDefinition{
		{ID: "task1", Name: "Task 1", EstimatedSeconds: 5},
		{ID: "task2", Name: "Task 2", EstimatedSeconds: 15},
		{ID: "task3", Name: "Task 3", EstimatedSeconds: 3},
	}

	tests := []struct {
		name          string
		only          string
		skip          sliceFlag
		fast          bool
		fastThreshold int
		wantCount     int
	}{
		{
			name:      "no filters",
			only:      "",
			skip:      sliceFlag{},
			fast:      false,
			wantCount: 3,
		},
		{
			name:      "only one task",
			only:      "task1",
			skip:      sliceFlag{},
			fast:      false,
			wantCount: 1,
		},
		{
			name:      "skip one task",
			only:      "",
			skip:      sliceFlag{"task2"},
			fast:      false,
			wantCount: 2,
		},
		{
			name:      "only multiple tasks comma separated",
			only:      "task1,task3",
			skip:      sliceFlag{},
			fast:      false,
			wantCount: 2,
		},
		{
			name:      "only multiple tasks with spaces and duplicates",
			only:      " task1 , task2 , task1 ",
			skip:      sliceFlag{},
			fast:      false,
			wantCount: 2,
		},
		{
			name:      "only with skip removes one of requested",
			only:      "task1,task2,task3",
			skip:      sliceFlag{"task2"},
			fast:      false,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterTasks(tasks, tt.only, tt.skip, tt.fast, tt.fastThreshold, false)

			if len(filtered) != tt.wantCount {
				t.Errorf("filterTasks() returned %d tasks, want %d", len(filtered), tt.wantCount)
			}

			// Verify only filter
			if tt.only != "" {
				// Build requested set from comma-separated list (mirrors filterTasks logic)
				rawIDs := strings.Split(tt.only, ",")
				requested := make(map[string]struct{})
				for _, raw := range rawIDs {
					id := strings.TrimSpace(raw)
					if id == "" {
						continue
					}
					requested[id] = struct{}{}
				}

				for _, task := range filtered {
					if _, ok := requested[task.ID]; !ok {
						t.Errorf("Task %q was not requested by only=%q", task.ID, tt.only)
					}
				}
			}

			// Verify skip filter
			for _, task := range filtered {
				for _, skipID := range tt.skip {
					if task.ID == skipID {
						t.Errorf("Task %q should have been skipped", skipID)
					}
				}
			}
		})
	}
}

func TestGroupTasksIntoPhases(t *testing.T) {
	tests := []struct {
		name       string
		tasks      []model.TaskDefinition
		phaseNames map[string]config.PhaseInfo
		wantPhases int
	}{
		{
			name:       "empty tasks",
			tasks:      []model.TaskDefinition{},
			phaseNames: map[string]config.PhaseInfo{},
			wantPhases: 0,
		},
		{
			name: "single phase",
			tasks: []model.TaskDefinition{
				{ID: "task1", Phase: "build"},
				{ID: "task2", Phase: "build"},
			},
			phaseNames: map[string]config.PhaseInfo{
				"build": {ID: "build", Name: "Build"},
			},
			wantPhases: 1,
		},
		{
			name: "multiple phases",
			tasks: []model.TaskDefinition{
				{ID: "task1", Phase: "build"},
				{ID: "task2", Phase: "build", Wait: true},
				{ID: "task3", Phase: "test"},
			},
			phaseNames: map[string]config.PhaseInfo{
				"build": {ID: "build", Name: "Build"},
				"test":  {ID: "test", Name: "Test"},
			},
			wantPhases: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phases := groupTasksIntoPhases(tt.tasks, tt.phaseNames)

			if len(phases) != tt.wantPhases {
				t.Errorf("groupTasksIntoPhases() returned %d phases, want %d", len(phases), tt.wantPhases)
			}

			// Verify all tasks are in phases
			totalTasks := 0
			for _, phase := range phases {
				totalTasks += len(phase.Tasks)
			}

			if totalTasks != len(tt.tasks) {
				t.Errorf("Total tasks in phases = %d, want %d", totalTasks, len(tt.tasks))
			}
		})
	}
}

func TestGroupTasksIntoPhasesWithWait(t *testing.T) {
	tasks := []model.TaskDefinition{
		{ID: "task1", Phase: "build"},
		{ID: "task2", Phase: "build", Wait: true},
		{ID: "task3", Phase: "test"},
		{ID: "task4", Phase: "test"},
	}

	phaseNames := map[string]config.PhaseInfo{
		"build": {ID: "build", Name: "Build"},
		"test":  {ID: "test", Name: "Test"},
	}

	phases := groupTasksIntoPhases(tasks, phaseNames)

	if len(phases) != 2 {
		t.Fatalf("Expected 2 phases, got %d", len(phases))
	}

	// First phase should have 2 tasks (task1 and task2)
	if len(phases[0].Tasks) != 2 {
		t.Errorf("First phase should have 2 tasks, got %d", len(phases[0].Tasks))
	}

	// Second phase should have 2 tasks (task3 and task4)
	if len(phases[1].Tasks) != 2 {
		t.Errorf("Second phase should have 2 tasks, got %d", len(phases[1].Tasks))
	}
}
