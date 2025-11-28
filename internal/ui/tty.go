package ui

import (
	"os"
	"golang.org/x/term"
)

// IsTTY returns true if the given file descriptor is a terminal
func IsTTY(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}

// GetTerminalWidth returns the width of the terminal, or 80 if not a TTY
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // Default width
	}
	if width < 40 {
		return 40 // Minimum width
	}
	return width
}

// IsColorEnabled returns true if color output should be enabled
func IsColorEnabled() bool {
	// Check NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	
	// Check if stdout is a TTY
	return IsTTY(os.Stdout.Fd())
}
