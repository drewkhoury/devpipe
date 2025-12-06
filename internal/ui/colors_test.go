package ui

import (
	"strings"
	"testing"
)

func TestColors(t *testing.T) {
	// Test that color functions return non-empty strings
	c := NewColors(false)
	
	red := c.Red("test")
	green := c.Green("test")
	blue := c.Blue("test")
	
	if red == "" {
		t.Error("Expected non-empty red output")
	}
	if green == "" {
		t.Error("Expected non-empty green output")
	}
	if blue == "" {
		t.Error("Expected non-empty blue output")
	}
	
	// Test with colors disabled
	cNoColor := NewColors(true)
	plain := cNoColor.Red("test")
	if plain == "" {
		t.Error("Expected non-empty output even with colors disabled")
	}
}

func TestStatusSymbol(t *testing.T) {
	c := NewColors(false)
	
	tests := []string{"PASS", "FAIL", "SKIPPED", "RUNNING", "PENDING"}

	for _, status := range tests {
		t.Run(status, func(t *testing.T) {
			got := c.StatusSymbol(status)
			
			// Just verify it returns something non-empty
			if got == "" {
				t.Errorf("StatusSymbol(%s) returned empty string", status)
			}
		})
	}
}

func TestProgressBar(t *testing.T) {
	c := NewColors(true) // No color for easier testing
	
	tests := []struct {
		name    string
		current int
		total   int
		width   int
	}{
		{
			name:    "0 percent",
			current: 0,
			total:   10,
			width:   10,
		},
		{
			name:    "50 percent",
			current: 5,
			total:   10,
			width:   10,
		},
		{
			name:    "100 percent",
			current: 10,
			total:   10,
			width:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := c.ProgressBar(tt.current, tt.total, tt.width)
			
			// Just verify it returns something
			if bar == "" && tt.total > 0 {
				t.Error("Expected non-empty progress bar")
			}
		})
	}
}

// stripAnsi removes ANSI escape codes from a string
func stripAnsi(s string) string {
	// Simple implementation - just remove common ANSI codes
	s = strings.ReplaceAll(s, "\033[0m", "")
	s = strings.ReplaceAll(s, "\033[32m", "")
	s = strings.ReplaceAll(s, "\033[31m", "")
	s = strings.ReplaceAll(s, "\033[33m", "")
	s = strings.ReplaceAll(s, "\033[34m", "")
	s = strings.ReplaceAll(s, "\033[36m", "")
	s = strings.ReplaceAll(s, "\033[90m", "")
	s = strings.ReplaceAll(s, "\033[1m", "")
	return s
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		ms   int64
	}{
		{
			name: "milliseconds",
			ms:   500,
		},
		{
			name: "seconds",
			ms:   1500,
		},
		{
			name: "minutes",
			ms:   65000,
		},
		{
			name: "zero",
			ms:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.ms)
			
			// Just verify it returns something
			if got == "" {
				t.Errorf("FormatDuration(%d) returned empty string", tt.ms)
			}
		})
	}
}

func TestMoreColorFunctions(t *testing.T) {
	c := NewColors(false)

	// Test all color functions
	if c.Yellow("test") == "" {
		t.Error("Expected non-empty yellow output")
	}

	if c.Cyan("test") == "" {
		t.Error("Expected non-empty cyan output")
	}

	if c.Bold("test") == "" {
		t.Error("Expected non-empty bold output")
	}

	if c.Gray("test") == "" {
		t.Error("Expected non-empty gray output")
	}
}

func TestStatusColor(t *testing.T) {
	c := NewColors(false)

	tests := []string{"PASS", "FAIL", "RUNNING", "PENDING", "SKIPPED", "unknown"}

	for _, status := range tests {
		t.Run(status, func(t *testing.T) {
			colored := c.StatusColor(status, "test")
			
			if colored == "" {
				t.Errorf("StatusColor(%s, test) returned empty string", status)
			}
		})
	}
}

func TestRenderer(t *testing.T) {
	// Test Renderer creation
	r := NewRenderer(UIModeBasic, false, false)

	if r == nil {
		t.Fatal("Expected non-nil renderer")
	}

	// Test color methods
	if r.Green("test") == "" {
		t.Error("Expected non-empty green output")
	}

	if r.Red("test") == "" {
		t.Error("Expected non-empty red output")
	}

	if r.Gray("test") == "" {
		t.Error("Expected non-empty gray output")
	}

	// Test StatusColor wrapper
	if r.StatusColor("PASS") == "" {
		t.Error("Expected non-empty status color output")
	}
}
