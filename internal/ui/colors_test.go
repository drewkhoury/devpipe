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

func TestGetTerminalHeight(t *testing.T) {
	height := GetTerminalHeight()

	// Should return a reasonable value (default 24 if not detectable)
	if height <= 0 {
		t.Errorf("Expected positive height, got %d", height)
	}

	if height < 10 || height > 1000 {
		t.Logf("Unusual terminal height: %d (might be default)", height)
	}
}

func TestIsColorEnabled(t *testing.T) {
	// Test that IsColorEnabled returns a boolean
	enabled := IsColorEnabled()

	// Just verify it returns without error
	_ = enabled
	t.Logf("IsColorEnabled() = %v", enabled)
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

func TestColorsWithColorsEnabled(t *testing.T) {
	c := NewColors(true)

	// Test all color functions with colors enabled
	red := c.Red("test")
	if !strings.Contains(red, "\033[") {
		t.Error("Expected ANSI codes in red output when colors enabled")
	}

	green := c.Green("test")
	if !strings.Contains(green, "\033[") {
		t.Error("Expected ANSI codes in green output when colors enabled")
	}

	yellow := c.Yellow("test")
	if !strings.Contains(yellow, "\033[") {
		t.Error("Expected ANSI codes in yellow output when colors enabled")
	}

	blue := c.Blue("test")
	if !strings.Contains(blue, "\033[") {
		t.Error("Expected ANSI codes in blue output when colors enabled")
	}

	cyan := c.Cyan("test")
	if !strings.Contains(cyan, "\033[") {
		t.Error("Expected ANSI codes in cyan output when colors enabled")
	}

	gray := c.Gray("test")
	if !strings.Contains(gray, "\033[") {
		t.Error("Expected ANSI codes in gray output when colors enabled")
	}

	bold := c.Bold("test")
	if !strings.Contains(bold, "\033[") {
		t.Error("Expected ANSI codes in bold output when colors enabled")
	}
}

func TestColorsWithColorsDisabled(t *testing.T) {
	c := NewColors(false)

	// Test all color functions with colors disabled
	red := c.Red("test")
	if red != "test" {
		t.Errorf("Expected plain text when colors disabled, got %q", red)
	}

	green := c.Green("test")
	if green != "test" {
		t.Errorf("Expected plain text when colors disabled, got %q", green)
	}

	yellow := c.Yellow("test")
	if yellow != "test" {
		t.Errorf("Expected plain text when colors disabled, got %q", yellow)
	}

	blue := c.Blue("test")
	if blue != "test" {
		t.Errorf("Expected plain text when colors disabled, got %q", blue)
	}

	cyan := c.Cyan("test")
	if cyan != "test" {
		t.Errorf("Expected plain text when colors disabled, got %q", cyan)
	}

	gray := c.Gray("test")
	if gray != "test" {
		t.Errorf("Expected plain text when colors disabled, got %q", gray)
	}

	bold := c.Bold("test")
	if bold != "test" {
		t.Errorf("Expected plain text when colors disabled, got %q", bold)
	}
}

func TestStatusSymbolDefault(t *testing.T) {
	c := NewColors(false)

	// Test default case (unknown status)
	symbol := c.StatusSymbol("UNKNOWN")
	if symbol != " " {
		t.Errorf("Expected space for unknown status, got %q", symbol)
	}
}

func TestProgressBarEdgeCases(t *testing.T) {
	c := NewColors(true)

	// Test zero total
	bar := c.ProgressBar(5, 0, 10)
	if bar != "" {
		t.Error("Expected empty bar when total is 0")
	}

	// Test current > total (should cap at width)
	bar2 := c.ProgressBar(150, 100, 10)
	if bar2 == "" {
		t.Error("Expected non-empty bar")
	}

	// Test 100% completion (should be green)
	bar3 := c.ProgressBar(100, 100, 10)
	if !strings.Contains(bar3, "\033[") {
		t.Error("Expected colored output for 100% completion")
	}

	// Test 50% completion (should be blue)
	bar4 := c.ProgressBar(50, 100, 10)
	if !strings.Contains(bar4, "\033[") {
		t.Error("Expected colored output for 50% completion")
	}

	// Test < 50% completion (should be gray)
	bar5 := c.ProgressBar(25, 100, 10)
	if !strings.Contains(bar5, "\033[") {
		t.Error("Expected colored output for 25% completion")
	}
}

func TestProgressBarWithoutColors(t *testing.T) {
	c := NewColors(false)

	// Test progress bar without colors
	bar := c.ProgressBar(50, 100, 10)
	if bar == "" {
		t.Error("Expected non-empty bar")
	}

	// Should not contain ANSI codes
	if strings.Contains(bar, "\033[") {
		t.Error("Expected no ANSI codes when colors disabled")
	}
}
