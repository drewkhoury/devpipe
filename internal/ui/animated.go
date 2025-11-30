package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// AnimatedTaskTracker tracks live progress of tasks
type AnimatedTaskTracker struct {
	tasks        []TaskProgress
	mu           sync.RWMutex
	done         chan struct{}
	renderer     *Renderer
	headerLines  int
	firstRender  bool
	logLines     []string
	maxLogLines  int
	verboseLines []string // Verbose output lines
	maxVerbose   int      // Max verbose lines to show (5)
	termHeight   int
	animLines    int
	refreshMs    int
	groupBy      string // "type" or "phase"
	maxIDWidth   int    // Calculated once at init for consistent alignment
}

// NewAnimatedTaskTracker creates a new animated task tracker
func NewAnimatedTaskTracker(renderer *Renderer, tasks []TaskProgress, headerLines int, refreshMs int, groupBy string) *AnimatedTaskTracker {
	termHeight := GetTerminalHeight()

	// Calculate animation lines more accurately
	var animLines int
	if renderer.mode == UIModeFull {
		// Full mode: overall progress (2 lines) + grouped tasks
		animLines = 2 // Overall progress + blank

		// Count lines per group
		groups := make(map[string]int)
		for _, task := range tasks {
			groups[task.Type]++
		}

		// Each group: header (1) + tasks (N) + footer (1) + blank (1)
		for _, count := range groups {
			animLines += 3 + count // header + stages + footer + blank
		}
	} else {
		// Basic mode: progress bar + blank + tasks + blank
		animLines = 2 + len(tasks) + 1
	}

	// Reserve space for logs (rest of terminal minus header, animation, log header, and summary)
	// animLines + 1 (for "─── Output ───") + maxLogLines + 6 (summary)
	maxLogLines := termHeight - headerLines - animLines - 1 - 6 // 1 for log header, 6 for summary
	if maxLogLines < 3 {
		maxLogLines = 3
	}

	// Validate refreshMs (20ms to 2000ms allowed)
	if refreshMs < 20 || refreshMs > 2000 {
		refreshMs = 500 // Default to 500ms if invalid
	}

	// Validate groupBy
	if groupBy != "type" && groupBy != "phase" {
		groupBy = "type" // Default to type if invalid
	}

	// Calculate max task ID width once for consistent alignment
	// Cap at 45 to prevent very long task names from breaking layout
	maxIDWidth := 12
	for _, task := range tasks {
		taskLen := len(task.ID)
		if taskLen > 45 {
			taskLen = 45
		}
		if taskLen > maxIDWidth {
			maxIDWidth = taskLen
		}
	}

	return &AnimatedTaskTracker{
		tasks:        tasks,
		done:         make(chan struct{}),
		renderer:     renderer,
		headerLines:  headerLines,
		firstRender:  true,
		logLines:     []string{},
		maxLogLines:  maxLogLines,
		verboseLines: []string{},
		maxVerbose:   5, // Fixed 5 lines for verbose output
		termHeight:   termHeight,
		animLines:    animLines,
		refreshMs:    refreshMs,
		groupBy:      groupBy,
		maxIDWidth:   maxIDWidth,
	}
}

// Start begins the animation loop
func (a *AnimatedTaskTracker) Start() error {
	// Test if terminal supports animation
	if err := a.testRender(); err != nil {
		return fmt.Errorf("animation not supported: %w", err)
	}

	// Start animation loop (it will do the initial render)
	go a.animationLoop()

	// Give the animation loop a moment to start and do initial render
	time.Sleep(50 * time.Millisecond)

	return nil
}

// Stop stops the animation loop
func (a *AnimatedTaskTracker) Stop() {
	// Always restore cursor, even if something goes wrong
	defer fmt.Print("\033[?25h") // Show cursor again

	close(a.done)

	fmt.Println() // Add newline after animation
}

// testRender tests if terminal supports ANSI escape codes
func (a *AnimatedTaskTracker) testRender() error {
	defer func() {
		if r := recover(); r != nil {
			// Rendering panicked
		}
	}()

	// Try basic ANSI sequences
	fmt.Print("\033[s") // Save cursor position
	fmt.Print("\033[u") // Restore cursor position

	return nil
}

// UpdateTask updates a task's progress
func (a *AnimatedTaskTracker) UpdateTask(id string, status string, elapsed float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for i := range a.tasks {
		if a.tasks[i].ID == id {
			a.tasks[i].Status = status
			a.tasks[i].ElapsedSeconds = elapsed
			break
		}
	}
}

// AddLogLine adds a log line to the display
func (a *AnimatedTaskTracker) AddLogLine(line string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logLines = append(a.logLines, line)

	// Keep only the last maxLogLines
	if len(a.logLines) > a.maxLogLines {
		a.logLines = a.logLines[len(a.logLines)-a.maxLogLines:]
	}
}

// AddVerboseLine adds a verbose log line to the display
func (a *AnimatedTaskTracker) AddVerboseLine(line string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.verboseLines = append(a.verboseLines, line)

	// Keep only the last maxVerbose lines (5)
	if len(a.verboseLines) > a.maxVerbose {
		a.verboseLines = a.verboseLines[len(a.verboseLines)-a.maxVerbose:]
	}
}

// animationLoop continuously updates the display
func (a *AnimatedTaskTracker) animationLoop() {
	// Do initial render immediately
	a.render()

	// Use configured refresh rate
	ticker := time.NewTicker(time.Duration(a.refreshMs) * time.Millisecond)
	defer ticker.Stop()

	// Do another render after a short delay to catch fast tasks
	time.Sleep(100 * time.Millisecond)
	a.render()

	for {
		select {
		case <-a.done:
			return
		case <-ticker.C:
			a.render()
		}
	}
}

// render draws the current state
func (a *AnimatedTaskTracker) render() {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.firstRender {
		// First render: hide cursor and print normally
		fmt.Print("\033[?25l") // Hide cursor
		a.firstRender = false
		if a.renderer.mode == UIModeFull {
			a.renderFullMode()
		} else {
			a.renderBasicMode()
		}
	} else {
		// Subsequent renders: move to top, clear animation area, redraw

		// Move cursor to start of animation area and clear all lines
		linesToMove := a.calculateLines()

		// Move cursor up to start of animation area
		if linesToMove > 0 {
			fmt.Printf("\033[%dA", linesToMove)
		}

		// Move to beginning of line and clear everything below
		fmt.Print("\r")     // Move to start of line
		fmt.Print("\033[J") // Clear from cursor to end of screen

		// Render animation
		if a.renderer.mode == UIModeFull {
			a.renderFullMode()
		} else {
			a.renderBasicMode()
		}
	}
}

// calculateLines calculates how many lines we need to clear
// This should return a FIXED number based on maxLogLines, not actual log count
func (a *AnimatedTaskTracker) calculateLines() int {
	if a.renderer.mode == UIModeFull {
		// Full mode: overall progress + grouped stages + log box (fixed size)
		lines := 2 // Overall progress bar + blank line

		// Group tasks using the same logic as rendering
		groups := make(map[string][]TaskProgress)
		for _, task := range a.tasks {
			var groupKey string
			if a.groupBy == "phase" {
				// Use the phase name if available, otherwise fall back to "Phase N"
				if task.PhaseName != "" {
					groupKey = task.PhaseName
				} else {
					groupKey = fmt.Sprintf("Phase %d", task.Phase)
				}
			} else {
				groupKey = task.Type
			}
			groups[groupKey] = append(groups[groupKey], task)
		}

		for _, taskList := range groups {
			lines += 2 + len(taskList) + 1 // Header + tasks + footer + blank
		}

		// Add FIXED log box lines (always reserve maxLogLines)
		lines += 1 + a.maxLogLines // "─── Output ───" + max log lines

		return lines
	} else {
		// Basic mode: progress bar + blank + tasks + blank + log header + FIXED log lines
		return 2 + len(a.tasks) + 1 + 1 + a.maxLogLines
	}
}

// renderBasicMode renders the basic animated mode
func (a *AnimatedTaskTracker) renderBasicMode() {
	// Calculate overall progress
	overallProgress := CalculateOverallProgress(a.tasks)

	// Render progress bar
	barWidth := 40
	if a.renderer.width > 80 {
		barWidth = 60
	}

	completed := 0
	for _, task := range a.tasks {
		if task.Status == "PASS" || task.Status == "FAIL" || task.Status == "SKIPPED" {
			completed++
		}
	}

	bar := a.renderer.colors.ProgressBar(int(overallProgress), 100, barWidth)
	fmt.Printf("%s (%d/%d tasks)\n\n", bar, completed, len(a.tasks))

	// Render task list
	for _, task := range a.tasks {
		symbol := a.renderer.colors.StatusSymbol(task.Status)

		// Truncate task ID if needed (max 45 chars)
		taskID := task.ID
		if len(taskID) > 45 {
			taskID = taskID[:42] + "..."
		}

		switch task.Status {
		case "PASS", "FAIL", "SKIPPED":
			statusText := a.renderer.colors.StatusColor(task.Status, task.Status)
			fmt.Printf("%s %-*s %s\n", symbol, a.maxIDWidth, taskID, statusText)
		case "RUNNING":
			progress := CalculateTaskProgress(task.ElapsedSeconds, task.EstimatedSeconds)
			progressText := fmt.Sprintf("%.0f%%", progress)
			fmt.Printf("%s %-*s %s %s\n", symbol, a.maxIDWidth, taskID,
				a.renderer.colors.Blue("running..."),
				a.renderer.colors.Gray(progressText))
		case "PENDING":
			fmt.Printf("%s %-*s %s\n", symbol, a.maxIDWidth, taskID, a.renderer.colors.Gray("pending"))
		}
	}

	// Render log box (fixed size)
	fmt.Println()
	fmt.Println(a.renderer.colors.Bold("─── Output ───"))

	// Render log lines (pad with empty lines to maintain fixed size)
	for i := 0; i < a.maxLogLines; i++ {
		if i < len(a.logLines) {
			fmt.Println(a.logLines[i])
		} else {
			fmt.Println() // Empty line to maintain fixed height
		}
	}
}

// renderFullMode renders the full animated mode with grouped stages
func (a *AnimatedTaskTracker) renderFullMode() {
	// Calculate overall progress
	overallProgress := CalculateOverallProgress(a.tasks)

	// Render overall progress bar
	barWidth := 40
	if a.renderer.width > 60 {
		barWidth = 60
	}

	bar := a.renderer.colors.ProgressBar(int(overallProgress), 100, barWidth)
	fmt.Printf("Overall: %s\n\n", bar)

	// Group tasks by type or phase
	groups := make(map[string][]TaskProgress)
	groupOrder := []string{}
	seen := make(map[string]bool)

	for _, task := range a.tasks {
		var groupKey string
		if a.groupBy == "phase" {
			// Use the phase name if available, otherwise fall back to "Phase N"
			if task.PhaseName != "" {
				groupKey = task.PhaseName
			} else {
				groupKey = fmt.Sprintf("Phase %d", task.Phase)
			}
		} else {
			groupKey = task.Type
		}

		if !seen[groupKey] {
			groupOrder = append(groupOrder, groupKey)
			seen[groupKey] = true
		}
		groups[groupKey] = append(groups[groupKey], task)
	}

	// Render each group
	for _, groupName := range groupOrder {
		taskList := groups[groupName]

		// Group header
		headerText := fmt.Sprintf("─ %s ", strings.Title(groupName))
		padding := strings.Repeat("─", a.renderer.width-len(headerText)-2)
		fmt.Printf("┌%s%s┐\n", headerText, padding)

		// Tasks in group
		for _, task := range taskList {
			symbol := a.renderer.colors.StatusSymbol(task.Status)

			// Truncate task ID if needed (max 45 chars)
			taskID := task.ID
			if len(taskID) > 45 {
				taskID = taskID[:42] + "..."
			}

			// Build the content line
			var content string
			switch task.Status {
			case "PASS":
				duration := FormatDuration(int64(task.ElapsedSeconds * 1000))
				content = fmt.Sprintf("%s %-*s %s", symbol, a.maxIDWidth, taskID,
					a.renderer.colors.Green(duration))
			case "FAIL":
				duration := FormatDuration(int64(task.ElapsedSeconds * 1000))
				content = fmt.Sprintf("%s %-*s %s", symbol, a.maxIDWidth, taskID,
					a.renderer.colors.Red(duration))
			case "SKIPPED":
				content = fmt.Sprintf("%s %-*s %s", symbol, a.maxIDWidth, taskID,
					a.renderer.colors.Yellow("skipped"))
			case "RUNNING":
				progress := CalculateTaskProgress(task.ElapsedSeconds, task.EstimatedSeconds)
				elapsed := FormatDuration(int64(task.ElapsedSeconds * 1000))
				estimated := FormatDuration(int64(task.EstimatedSeconds * 1000))
				if task.IsEstimateGuess {
					estimated = estimated + "?"
				}

				// Mini progress bar
				miniBarWidth := 12
				miniBar := a.renderer.colors.ProgressBar(int(progress), 100, miniBarWidth)

				content = fmt.Sprintf("%s %-*s %s / %s   %s", symbol, a.maxIDWidth, taskID,
					elapsed, estimated, miniBar)
			case "PENDING":
				content = fmt.Sprintf("%s %-*s %s", symbol, a.maxIDWidth, taskID,
					a.renderer.colors.Gray("pending"))
			}

			// Print with consistent width
			fmt.Printf("│ %s\n", content)
		}

		// Group footer
		fmt.Printf("└%s┘\n\n", strings.Repeat("─", a.renderer.width-2))
	}

	// Render log box (fixed size)
	fmt.Println(a.renderer.colors.Bold("─── Output ───"))

	// Render log lines (pad with empty lines to maintain fixed size)
	for i := 0; i < a.maxLogLines; i++ {
		if i < len(a.logLines) {
			fmt.Println(a.logLines[i])
		} else {
			fmt.Println() // Empty line to maintain fixed height
		}
	}
}
