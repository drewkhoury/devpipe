package ui

import (
	"os"
	"golang.org/x/term"
)

// IsTTY checks if the given file descriptor is a terminal
func IsTTY(fd uintptr) bool {
	return term.IsTerminal(int(fd))
}

// GetTerminalWidth returns the terminal width, or 80 if not a TTY
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width < 40 {
		return 80
	}
	return width
}

// GetTerminalHeight returns the terminal height, or 24 if not a TTY
func GetTerminalHeight() int {
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || height < 10 {
		return 24
	}
	return height
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
