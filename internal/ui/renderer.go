package ui

import (
	"fmt"
	"os"
	"strings"
)

// UIMode represents the UI rendering mode
type UIMode string

// UI mode constants
const (
	UIModeBasic UIMode = "basic"
	UIModeFull  UIMode = "full"
)

// Renderer handles UI rendering
type Renderer struct {
	mode        UIMode
	colors      *Colors
	width       int
	isTTY       bool
	animated    bool
	tracker     *AnimatedTaskTracker // Reference to tracker for verbose output
	pipelineLog *os.File             // Log file for verbose output
}

// NewRenderer creates a new UI renderer
func NewRenderer(mode UIMode, enableColors bool, animated bool) *Renderer {
	isTTY := IsTTY(uintptr(1)) // stdout
	width := GetTerminalWidth()

	// Force basic mode if not a TTY
	if !isTTY && mode != UIModeBasic {
		mode = UIModeBasic
	}

	// Disable colors if not a TTY or explicitly disabled
	if !isTTY {
		enableColors = false
	}

	// Disable animation if not a TTY
	if !isTTY {
		animated = false
	}

	return &Renderer{
		mode:     mode,
		colors:   NewColors(enableColors),
		width:    width,
		isTTY:    isTTY,
		animated: animated,
	}
}

// IsAnimated returns true if animated mode is enabled
func (r *Renderer) IsAnimated() bool {
	return r.animated
}

// RenderHeader renders the pipeline header
func (r *Renderer) RenderHeader(runID, repoRoot string, gitMode string, changedFiles int) {
	switch r.mode {
	case UIModeFull:
		r.renderFullHeader(runID, repoRoot, gitMode, changedFiles)
	default:
		r.renderBasicHeader(runID, repoRoot, gitMode, changedFiles)
	}
}

func (r *Renderer) renderFullHeader(runID, repoRoot string, gitMode string, changedFiles int) {
	line := strings.Repeat("═", r.width-2)
	fmt.Printf("╔%s╗\n", line)
	fmt.Printf("║ %s%-*s%s║\n",
		r.colors.Bold("devpipe run "+runID),
		r.width-len("devpipe run "+runID)-3,
		"",
		"")
	fmt.Printf("║ Repo: %-*s║\n", r.width-9, repoRoot)
	if gitMode != "" {
		info := fmt.Sprintf("Git: %s | Files: %d", gitMode, changedFiles)
		fmt.Printf("║ %-*s║\n", r.width-3, info)
	}
	fmt.Printf("╚%s╝\n", line)
	fmt.Println()
}

func (r *Renderer) renderBasicHeader(runID, repoRoot string, gitMode string, changedFiles int) {
	fmt.Printf("devpipe run %s\n", runID)
	fmt.Printf("Repo root: %s\n", repoRoot)
	if gitMode != "" {
		fmt.Printf("Git mode: %s\n", gitMode)
		fmt.Printf("Changed files: %d\n", changedFiles)
	}
	fmt.Println() // Blank line after header
}

// truncateTaskID truncates a task ID to maxLen, adding "..." if needed
func truncateTaskID(id string, maxLen int) string {
	if len(id) <= maxLen {
		return id
	}
	return id[:maxLen-3] + "..."
}

// RenderTaskStart renders when a task starts
func (r *Renderer) RenderTaskStart(id, command string, verbose bool) {
	// In animated mode, don't print anything yet
	if r.animated {
		return
	}

	taskID := truncateTaskID(id, 15)
	if verbose {
		fmt.Printf("[%-15s] %s    %s\n", taskID, r.colors.Blue("RUN"), command)
	} else {
		fmt.Printf("[%-15s] %s\n", taskID, r.colors.Blue("RUN"))
	}
}

// RenderTaskComplete renders when a task completes
func (r *Renderer) RenderTaskComplete(id, status string, exitCode *int, durationMs int64, verbose bool) {
	// In animated mode, don't print anything (animation handles it)
	if r.animated {
		return
	}

	taskID := truncateTaskID(id, 15)
	symbol := r.colors.StatusSymbol(status)
	statusText := r.colors.StatusColor(status, status)

	if verbose && exitCode != nil {
		fmt.Printf("[%-15s] %s %s (exit %d, %dms)\n", taskID, symbol, statusText, *exitCode, durationMs)
	} else {
		fmt.Printf("[%-15s] %s %s (%dms)\n", taskID, symbol, statusText, durationMs)
	}
	fmt.Println() // Blank line after task
}

// RenderTaskSkipped renders when a task is skipped
func (r *Renderer) RenderTaskSkipped(id, reason string, verbose bool) {
	// In animated mode, don't print anything (animation handles it)
	if r.animated {
		return
	}

	taskID := truncateTaskID(id, 15)
	symbol := r.colors.StatusSymbol("SKIPPED")
	if verbose {
		fmt.Printf("[%-15s] %s %s (%s)\n", taskID, symbol, r.colors.Yellow("SKIPPED"), reason)
	} else {
		fmt.Printf("[%-15s] %s %s\n", taskID, symbol, r.colors.Yellow("SKIPPED"))
	}
	fmt.Println() // Blank line after skipped task
}

// RenderSummary renders the final summary
func (r *Renderer) RenderSummary(results []TaskSummary, anyFailed bool, totalMs int64) {
	// Add blank line before summary if animated (animation already on screen)
	if r.animated {
		fmt.Println()
	}

	// Calculate max task ID width (same logic as animated view)
	maxIDWidth := 12
	for _, result := range results {
		taskLen := len(result.ID)
		if taskLen > 45 {
			taskLen = 45
		}
		if taskLen > maxIDWidth {
			maxIDWidth = taskLen
		}
	}

	fmt.Println(r.colors.Bold("Summary:"))

	for _, result := range results {
		symbol := r.colors.StatusSymbol(result.Status)
		statusText := r.colors.StatusColor(result.Status, fmt.Sprintf("%-10s", result.Status))
		seconds := float64(result.DurationMs) / 1000.0
		durationText := fmt.Sprintf("%.2fs (%dms)", seconds, result.DurationMs)

		annotation := ""
		if result.AutoFixed {
			annotation = " " + r.colors.Gray("[auto-fixed]")
		}

		taskID := truncateTaskID(result.ID, 45)
		fmt.Printf("  %s %-*s %s %s%s\n", symbol, maxIDWidth, taskID, statusText, durationText, annotation)
	}

	// Show total pipeline duration
	fmt.Println()
	totalSeconds := float64(totalMs) / 1000.0
	fmt.Printf("Total: %.2fs (%dms)\n", totalSeconds, totalMs)

	fmt.Println()
	if anyFailed {
		fmt.Println(r.colors.Red("devpipe: one or more tasks failed"))
	} else {
		fmt.Println(r.colors.Green("devpipe: all tasks passed or were skipped"))
	}
	fmt.Println() // Blank line at very end
}

// TaskSummary represents a task result for the summary
type TaskSummary struct {
	ID         string
	Status     string
	DurationMs int64
	AutoFixed  bool
}

// RenderProgress renders a progress bar (for full mode)
func (r *Renderer) RenderProgress(current, total int) {
	if r.mode == UIModeBasic {
		return
	}

	barWidth := 40
	if r.width > 80 {
		barWidth = 60
	}

	bar := r.colors.ProgressBar(current, total, barWidth)
	fmt.Printf("\n%s (%d/%d stages)\n\n", bar, current, total)
}

// CreateAnimatedTracker creates an animated task tracker
func (r *Renderer) CreateAnimatedTracker(tasks []TaskProgress, headerLines int, refreshMs int, groupBy string) *AnimatedTaskTracker {
	if !r.animated {
		return nil
	}
	return NewAnimatedTaskTracker(r, tasks, headerLines, refreshMs, groupBy)
}

// Blue returns the string formatted in blue color
func (r *Renderer) Blue(s string) string {
	return r.colors.Blue(s)
}

// Green returns the string formatted in green color
func (r *Renderer) Green(s string) string {
	return r.colors.Green(s)
}

// Red returns the string formatted in red color
func (r *Renderer) Red(s string) string {
	return r.colors.Red(s)
}

// Yellow returns the string formatted in yellow color
func (r *Renderer) Yellow(s string) string {
	return r.colors.Yellow(s)
}

// Cyan returns the string formatted in cyan color
func (r *Renderer) Cyan(s string) string {
	return r.colors.Cyan(s)
}

// Gray returns the string formatted in gray color
func (r *Renderer) Gray(s string) string {
	return r.colors.Gray(s)
}

// StatusColor returns the appropriate color for a given status string
func (r *Renderer) StatusColor(status string) string {
	return r.colors.StatusColor(status, status)
}

// SetTracker sets the animated tracker reference
func (r *Renderer) SetTracker(tracker *AnimatedTaskTracker) {
	r.tracker = tracker
}

// SetPipelineLog sets the pipeline log file for verbose output
func (r *Renderer) SetPipelineLog(log *os.File) {
	r.pipelineLog = log
}

// Verbose outputs a verbose message with [verbose] prefix
// Always writes to pipeline.log if available
// Only displays on screen if verbose flag is enabled
func (r *Renderer) Verbose(verbose bool, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	// Always write to pipeline.log (without color codes)
	if r.pipelineLog != nil {
		plainLine := fmt.Sprintf("[%-15s] %s\n", "verbose", msg)
		_, _ = r.pipelineLog.WriteString(plainLine) // Best effort logging
	}

	// Only output to console/tracker if verbose flag is enabled
	if verbose {
		// Pad "verbose" to 15 chars BEFORE applying color, so alignment works
		paddedVerbose := fmt.Sprintf("%-15s", "verbose")
		line := fmt.Sprintf("[%s] %s", r.colors.Gray(paddedVerbose), msg)

		if r.tracker != nil {
			r.tracker.AddLogLine(line)
		} else {
			fmt.Println(line)
		}
	}
}
