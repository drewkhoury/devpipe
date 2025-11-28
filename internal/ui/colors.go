package ui

import "fmt"

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[94m"  // Bright blue - more readable on dark backgrounds
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[90m"
	ColorBold   = "\033[1m"
)

// ColorFunc wraps text with color codes if colors are enabled
type ColorFunc func(string) string

// Colors holds all color functions
type Colors struct {
	enabled bool
}

// NewColors creates a new Colors instance
func NewColors(enabled bool) *Colors {
	return &Colors{enabled: enabled}
}

// Red returns red colored text
func (c *Colors) Red(s string) string {
	if !c.enabled {
		return s
	}
	return ColorRed + s + ColorReset
}

// Green returns green colored text
func (c *Colors) Green(s string) string {
	if !c.enabled {
		return s
	}
	return ColorGreen + s + ColorReset
}

// Yellow returns yellow colored text
func (c *Colors) Yellow(s string) string {
	if !c.enabled {
		return s
	}
	return ColorYellow + s + ColorReset
}

// Blue returns blue colored text
func (c *Colors) Blue(s string) string {
	if !c.enabled {
		return s
	}
	return ColorBlue + s + ColorReset
}

// Gray returns gray colored text
func (c *Colors) Gray(s string) string {
	if !c.enabled {
		return s
	}
	return ColorGray + s + ColorReset
}

// Bold returns bold text
func (c *Colors) Bold(s string) string {
	if !c.enabled {
		return s
	}
	return ColorBold + s + ColorReset
}

// StatusColor returns colored text based on status
func (c *Colors) StatusColor(status string, text string) string {
	switch status {
	case "PASS":
		return c.Green(text)
	case "FAIL":
		return c.Red(text)
	case "SKIPPED":
		return c.Yellow(text)
	case "RUNNING":
		return c.Blue(text)
	case "PENDING":
		return c.Gray(text)
	default:
		return text
	}
}

// StatusSymbol returns a colored symbol for the status
func (c *Colors) StatusSymbol(status string) string {
	switch status {
	case "PASS":
		return c.Green("✓")
	case "FAIL":
		return c.Red("✗")
	case "SKIPPED":
		return c.Yellow("⊘")
	case "RUNNING":
		return c.Blue("⚙")
	case "PENDING":
		return c.Gray("⋯")
	default:
		return " "
	}
}

// ProgressBar creates a simple progress bar
func (c *Colors) ProgressBar(current, total, width int) string {
	if total == 0 {
		return ""
	}
	
	percent := float64(current) / float64(total)
	filled := int(percent * float64(width))
	
	if filled > width {
		filled = width
	}
	
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	
	percentText := fmt.Sprintf(" %3.0f%%", percent*100)
	
	if c.enabled {
		if percent >= 1.0 {
			return c.Green(bar) + c.Green(percentText)
		} else if percent >= 0.5 {
			return c.Blue(bar) + percentText
		} else {
			return c.Gray(bar) + percentText
		}
	}
	
	return bar + percentText
}
