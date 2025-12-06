package ui

import (
	"testing"
	"time"
)

func TestCalculateTaskProgress(t *testing.T) {
	tests := []struct {
		name      string
		elapsed   float64
		estimated int
		want      float64
	}{
		{
			name:      "zero estimated",
			elapsed:   10.0,
			estimated: 0,
			want:      0,
		},
		{
			name:      "50% complete",
			elapsed:   5.0,
			estimated: 10,
			want:      50.0,
		},
		{
			name:      "100% complete",
			elapsed:   10.0,
			estimated: 10,
			want:      100.0,
		},
		{
			name:      "over 100% capped",
			elapsed:   15.0,
			estimated: 10,
			want:      100.0,
		},
		{
			name:      "just started",
			elapsed:   0.5,
			estimated: 10,
			want:      5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTaskProgress(tt.elapsed, tt.estimated)
			if got != tt.want {
				t.Errorf("CalculateTaskProgress(%v, %v) = %v, want %v",
					tt.elapsed, tt.estimated, got, tt.want)
			}
		})
	}
}

func TestCalculateOverallProgress(t *testing.T) {
	tests := []struct {
		name  string
		tasks []TaskProgress
		want  float64
	}{
		{
			name:  "empty tasks",
			tasks: []TaskProgress{},
			want:  0,
		},
		{
			name: "all completed",
			tasks: []TaskProgress{
				{Status: "PASS"},
				{Status: "PASS"},
				{Status: "FAIL"},
			},
			want: 100.0,
		},
		{
			name: "all pending",
			tasks: []TaskProgress{
				{Status: "PENDING"},
				{Status: "PENDING"},
			},
			want: 0,
		},
		{
			name: "mixed statuses",
			tasks: []TaskProgress{
				{Status: "PASS"},
				{Status: "RUNNING", ElapsedSeconds: 5.0, EstimatedSeconds: 10},
				{Status: "PENDING"},
			},
			want: 50.0, // (100 + 50 + 0) / 3 = 50
		},
		{
			name: "one running at 75%",
			tasks: []TaskProgress{
				{Status: "RUNNING", ElapsedSeconds: 7.5, EstimatedSeconds: 10},
			},
			want: 75.0,
		},
		{
			name: "skipped counts as complete",
			tasks: []TaskProgress{
				{Status: "SKIPPED"},
				{Status: "PASS"},
			},
			want: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateOverallProgress(tt.tasks)
			if got != tt.want {
				t.Errorf("CalculateOverallProgress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDurationProgress(t *testing.T) {
	tests := []struct {
		name string
		ms   int64
		want string
	}{
		{
			name: "less than 1 second",
			ms:   500,
			want: "0s",
		},
		{
			name: "exactly 1 second",
			ms:   1000,
			want: "1s",
		},
		{
			name: "multiple seconds",
			ms:   5000,
			want: "5s",
		},
		{
			name: "59 seconds",
			ms:   59000,
			want: "59s",
		},
		{
			name: "exactly 1 minute",
			ms:   60000,
			want: "1m",
		},
		{
			name: "1 minute 30 seconds",
			ms:   90000,
			want: "1m 30s",
		},
		{
			name: "5 minutes",
			ms:   300000,
			want: "5m",
		},
		{
			name: "exactly 1 hour",
			ms:   3600000,
			want: "1h",
		},
		{
			name: "1 hour 30 minutes",
			ms:   5400000,
			want: "1h 30m",
		},
		{
			name: "2 hours",
			ms:   7200000,
			want: "2h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.ms)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %v, want %v", tt.ms, got, tt.want)
			}
		})
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{5, "5"},
		{9, "9"},
		{10, "10"},
		{15, "15"},
		{59, "59"},
		{99, "99"},
	}

	for _, tt := range tests {
		got := formatInt(tt.input)
		if got != tt.want {
			t.Errorf("formatInt(%d) = %s, want %s", tt.input, got, tt.want)
		}
	}
}

func TestTaskProgress(t *testing.T) {
	// Test TaskProgress struct creation
	tp := TaskProgress{
		ID:               "test-task",
		Name:             "Test Task",
		Type:             "quality",
		Phase:            1,
		PhaseName:        "Build",
		Status:           "RUNNING",
		EstimatedSeconds: 30,
		IsEstimateGuess:  false,
		ElapsedSeconds:   15.0,
		StartTime:        time.Now(),
	}

	if tp.ID != "test-task" {
		t.Errorf("Expected ID 'test-task', got '%s'", tp.ID)
	}

	if tp.Status != "RUNNING" {
		t.Errorf("Expected Status 'RUNNING', got '%s'", tp.Status)
	}

	// Calculate progress for this task
	progress := CalculateTaskProgress(tp.ElapsedSeconds, tp.EstimatedSeconds)
	if progress != 50.0 {
		t.Errorf("Expected 50%% progress, got %v", progress)
	}
}

func TestCalculateOverallProgressCapped(t *testing.T) {
	// Test that overall progress is capped at 100%
	tasks := []TaskProgress{
		{Status: "PASS"},
		{Status: "RUNNING", ElapsedSeconds: 200.0, EstimatedSeconds: 10}, // Over 100%
	}

	progress := CalculateOverallProgress(tasks)
	if progress > 100.0 {
		t.Errorf("Expected progress <= 100%%, got %v", progress)
	}
}

func TestCalculateOverallProgressUnknownStatus(t *testing.T) {
	// Test with unknown status (should be treated like PENDING)
	tasks := []TaskProgress{
		{Status: "PASS"},
		{Status: "UNKNOWN"}, // Unknown status
		{Status: "PENDING"},
	}

	progress := CalculateOverallProgress(tasks)
	// Should be 33.33% (1 out of 3 completed)
	expected := 100.0 / 3.0
	if progress < expected-1 || progress > expected+1 {
		t.Errorf("Expected progress around %.2f%%, got %v", expected, progress)
	}
}

func TestFormatDurationEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		ms   int64
		want string
	}{
		{
			name: "zero milliseconds",
			ms:   0,
			want: "0s",
		},
		{
			name: "1 millisecond",
			ms:   1,
			want: "0s",
		},
		{
			name: "999 milliseconds",
			ms:   999,
			want: "0s",
		},
		{
			name: "1 hour 1 minute",
			ms:   3660000,
			want: "1h 1m",
		},
		{
			name: "2 hours 30 minutes",
			ms:   9000000,
			want: "2h 30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.ms)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %v, want %v", tt.ms, got, tt.want)
			}
		})
	}
}
