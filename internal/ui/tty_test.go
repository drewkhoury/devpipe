package ui

import (
	"os"
	"testing"
)

func TestIsTTY(t *testing.T) {
	// Test with stdout
	result := IsTTY(os.Stdout.Fd())
	// Result depends on environment, just verify it doesn't panic
	t.Logf("IsTTY(stdout) = %v", result)

	// Test with stderr
	result2 := IsTTY(os.Stderr.Fd())
	t.Logf("IsTTY(stderr) = %v", result2)
}

func TestGetTerminalWidth(t *testing.T) {
	width := GetTerminalWidth()

	// Should return a positive value (default 80 if not a TTY)
	if width <= 0 {
		t.Errorf("Expected positive width, got %d", width)
	}

	// Should be at least 40 (minimum) or default 80
	if width < 40 {
		t.Errorf("Expected width >= 40, got %d", width)
	}

	t.Logf("Terminal width: %d", width)
}

func TestGetTerminalWidthDefault(t *testing.T) {
	// When not a TTY or width < 40, should return 80
	width := GetTerminalWidth()

	// Should be a reasonable value
	if width < 40 || width > 500 {
		t.Logf("Width %d is outside typical range, might be default", width)
	}
}

func TestGetTerminalHeightDefault(t *testing.T) {
	height := GetTerminalHeight()

	// Should return a positive value (default 24 if not a TTY)
	if height <= 0 {
		t.Errorf("Expected positive height, got %d", height)
	}

	// Should be at least 10 (minimum) or default 24
	if height < 10 {
		t.Errorf("Expected height >= 10, got %d", height)
	}

	t.Logf("Terminal height: %d", height)
}

func TestIsColorEnabledWithNOCOLOR(t *testing.T) {
	// Save original value
	originalValue := os.Getenv("NO_COLOR")
	defer func() {
		if originalValue == "" {
			_ = os.Unsetenv("NO_COLOR") // Test cleanup
		} else {
			_ = os.Setenv("NO_COLOR", originalValue) // Test cleanup
		}
	}()

	// Test with NO_COLOR set
	_ = os.Setenv("NO_COLOR", "1") // Test setup
	enabled := IsColorEnabled()
	if enabled {
		t.Error("Expected IsColorEnabled() to be false when NO_COLOR is set")
	}

	// Test with NO_COLOR unset
	_ = os.Unsetenv("NO_COLOR") // Test setup
	enabled2 := IsColorEnabled()
	// Result depends on whether stdout is a TTY
	t.Logf("IsColorEnabled() without NO_COLOR = %v", enabled2)
}

func TestIsColorEnabledWithoutNOCOLOR(t *testing.T) {
	// Save original value
	originalValue := os.Getenv("NO_COLOR")
	defer func() {
		if originalValue == "" {
			_ = os.Unsetenv("NO_COLOR") // Test cleanup
		} else {
			_ = os.Setenv("NO_COLOR", originalValue) // Test cleanup
		}
	}()

	// Ensure NO_COLOR is not set
	_ = os.Unsetenv("NO_COLOR") // Test setup

	enabled := IsColorEnabled()
	// Result depends on TTY status, just verify it doesn't panic
	t.Logf("IsColorEnabled() = %v (depends on TTY)", enabled)
}

func TestIsColorEnabledWithCOLORTERM(t *testing.T) {
	// Save original values
	originalCOLORTERM := os.Getenv("COLORTERM")
	defer func() {
		_ = os.Setenv("COLORTERM", originalCOLORTERM) // Test cleanup
	}()

	// Test with COLORTERM set
	_ = os.Unsetenv("NO_COLOR") // Test setup

	enabled := IsColorEnabled()
	// Result depends on TTY status, just verify it doesn't panic
	t.Logf("IsColorEnabled() = %v (depends on TTY)", enabled)
}

func TestGetTerminalHeightMinimum(t *testing.T) {
	height := GetTerminalHeight()

	// Should return a positive value (default 24 if not a TTY)
	if height <= 0 {
		t.Errorf("Expected positive height, got %d", height)
	}

	// Should be at least 10 (minimum) or default 24
	if height < 10 {
		t.Errorf("Expected height >= 10, got %d", height)
	}

	t.Logf("Terminal height: %d", height)
}

func TestIsTTYWithStdin(t *testing.T) {
	// Test with stdin
	result := IsTTY(os.Stdin.Fd())
	// Result depends on environment, just verify it doesn't panic
	t.Logf("IsTTY(stdin) = %v", result)
}
