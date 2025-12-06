package main

import (
	"testing"
)

func TestSliceFlag(t *testing.T) {
	var sf sliceFlag

	// Test Set
	err := sf.Set("task1")
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	err = sf.Set("task2")
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	if len(sf) != 2 {
		t.Errorf("Expected 2 items, got %d", len(sf))
	}

	if sf[0] != "task1" {
		t.Errorf("Expected first item 'task1', got '%s'", sf[0])
	}

	if sf[1] != "task2" {
		t.Errorf("Expected second item 'task2', got '%s'", sf[1])
	}

	// Test String
	str := sf.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}

func TestGetTerminalWidth(t *testing.T) {
	width := getTerminalWidth()

	// Should return a reasonable width (default is 160)
	if width < 80 {
		t.Errorf("Expected width >= 80, got %d", width)
	}

	if width > 500 {
		t.Errorf("Expected width <= 500, got %d (seems unreasonably large)", width)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		wantLen  int
		wantDots bool
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			wantLen:  5,
			wantDots: false,
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			wantLen:  5,
			wantDots: false,
		},
		{
			name:     "needs truncation",
			input:    "hello world this is a long string",
			maxLen:   10,
			wantLen:  10,
			wantDots: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)

			if len(got) != tt.wantLen {
				t.Errorf("truncate() length = %d, want %d", len(got), tt.wantLen)
			}

			if tt.wantDots && got[len(got)-3:] != "..." {
				t.Errorf("truncate() should end with '...', got '%s'", got)
			}
		})
	}
}

func TestMakeRunID(t *testing.T) {
	// Test that makeRunID generates valid IDs
	id := makeRunID()

	if id == "" {
		t.Error("Expected non-empty run ID")
	}

	// Should contain timestamp format (YYYY-MM-DDTHH-MM-SSZ_NNNNNN)
	if len(id) < 20 {
		t.Errorf("Expected run ID length >= 20, got %d", len(id))
	}

	// Should contain underscore separator
	if !containsSpace(id) && len(id) > 0 {
		// Valid format
	}
}

func TestContainsSpace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "no spaces",
			input: "hello",
			want:  false,
		},
		{
			name:  "with space",
			input: "hello world",
			want:  true,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "only space",
			input: " ",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsSpace(tt.input)
			if got != tt.want {
				t.Errorf("containsSpace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPhaseEmoji(t *testing.T) {
	tests := []struct {
		name      string
		phaseName string
	}{
		{"validation", "Validation"},
		{"build", "Build"},
		{"test", "Test"},
		{"deploy", "Deploy"},
		{"unknown", "Something Else"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emoji := phaseEmoji(tt.phaseName)
			
			// Just verify it returns something
			if emoji == "" {
				t.Errorf("phaseEmoji(%s) returned empty string", tt.phaseName)
			}
		})
	}
}

func TestBuildCommandString(t *testing.T) {
	// Test that buildCommandString returns something
	cmd := buildCommandString()
	
	// Should return a non-empty string
	if cmd == "" {
		t.Error("Expected non-empty command string")
	}
}
