package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// AnimatedStageTracker tracks live progress of stages
type AnimatedStageTracker struct {
	stages       []StageProgress
	mu           sync.RWMutex
	done         chan struct{}
	renderer     *Renderer
	headerLines  int
	firstRender  bool
	logLines     []string
	maxLogLines  int
	termHeight   int
	animLines    int
	refreshMs    int
}

// NewAnimatedStageTracker creates a new animated stage tracker
func NewAnimatedStageTracker(renderer *Renderer, stages []StageProgress, headerLines int, refreshMs int) *AnimatedStageTracker {
	termHeight := GetTerminalHeight()
	
	// Calculate animation lines more accurately
	var animLines int
	if renderer.mode == UIModeFull {
		// Full mode: overall progress (2 lines) + grouped stages
		animLines = 2 // Overall progress + blank
		
		// Count lines per group
		groups := make(map[string]int)
		for _, stage := range stages {
			groups[stage.Group]++
		}
		
		// Each group: header (1) + stages (N) + footer (1) + blank (1)
		for _, count := range groups {
			animLines += 3 + count // header + stages + footer + blank
		}
	} else {
		// Basic mode: progress bar + blank + stages + blank
		animLines = 2 + len(stages) + 1
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
	
	return &AnimatedStageTracker{
		stages:      stages,
		done:        make(chan struct{}),
		renderer:    renderer,
		headerLines: headerLines,
		firstRender: true,
		logLines:    []string{},
		maxLogLines: maxLogLines,
		termHeight:  termHeight,
		animLines:   animLines,
		refreshMs:   refreshMs,
	}
}

// Start begins the animation loop
func (a *AnimatedStageTracker) Start() error {
	// Test if terminal supports animation
	if err := a.testRender(); err != nil {
		return fmt.Errorf("animation not supported: %w", err)
	}
	
	go a.animationLoop()
	return nil
}

// Stop stops the animation loop
func (a *AnimatedStageTracker) Stop() {
	// Always restore cursor, even if something goes wrong
	defer fmt.Print("\033[?25h") // Show cursor again
	
	close(a.done)
	time.Sleep(50 * time.Millisecond) // Let final updates propagate
	a.render() // Force final render with all stages complete
	time.Sleep(300 * time.Millisecond) // Show final 100% state
}

// testRender tests if terminal supports ANSI escape codes
func (a *AnimatedStageTracker) testRender() error {
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

// UpdateStage updates a stage's progress
func (a *AnimatedStageTracker) UpdateStage(id string, status string, elapsed float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	for i := range a.stages {
		if a.stages[i].ID == id {
			a.stages[i].Status = status
			a.stages[i].ElapsedSeconds = elapsed
			break
		}
	}
}

// AddLogLine adds a log line to the display
func (a *AnimatedStageTracker) AddLogLine(line string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.logLines = append(a.logLines, line)
	
	// Keep only the last maxLogLines
	if len(a.logLines) > a.maxLogLines {
		a.logLines = a.logLines[len(a.logLines)-a.maxLogLines:]
	}
}

// animationLoop continuously updates the display
func (a *AnimatedStageTracker) animationLoop() {
	// Use configured refresh rate
	ticker := time.NewTicker(time.Duration(a.refreshMs) * time.Millisecond)
	defer ticker.Stop()
	
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
func (a *AnimatedStageTracker) render() {
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
		
		// Move cursor to animation start
		linesToMove := a.calculateLines()
		fmt.Printf("\033[%dA", linesToMove) // Move up
		
		// Clear each line of the animation
		for i := 0; i < linesToMove; i++ {
			fmt.Print("\033[2K") // Clear entire line
			if i < linesToMove-1 {
				fmt.Print("\n") // Move to next line
			}
		}
		
		// Move back to start of animation area
		fmt.Printf("\033[%dA", linesToMove-1)
		
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
func (a *AnimatedStageTracker) calculateLines() int {
	if a.renderer.mode == UIModeFull {
		// Full mode: overall progress + grouped stages + log box (fixed size)
		lines := 2 // Overall progress bar + blank line
		
		// Group stages
		groups := make(map[string][]StageProgress)
		for _, stage := range a.stages {
			groups[stage.Group] = append(groups[stage.Group], stage)
		}
		
		for _, stages := range groups {
			lines += 2 + len(stages) + 1 // Header + stages + footer + blank
		}
		
		// Add FIXED log box lines (always reserve maxLogLines)
		lines += 1 + a.maxLogLines // "─── Output ───" + max log lines
		
		return lines
	} else {
		// Basic mode: progress bar + blank + stages + blank + log header + FIXED log lines
		return 2 + len(a.stages) + 1 + 1 + a.maxLogLines
	}
}

// renderBasicMode renders the basic animated mode
func (a *AnimatedStageTracker) renderBasicMode() {
	// Calculate overall progress
	overallProgress := CalculateOverallProgress(a.stages)
	
	// Render progress bar
	barWidth := 40
	if a.renderer.width > 80 {
		barWidth = 60
	}
	
	completed := 0
	for _, stage := range a.stages {
		if stage.Status == "PASS" || stage.Status == "FAIL" || stage.Status == "SKIPPED" {
			completed++
		}
	}
	
	bar := a.renderer.colors.ProgressBar(int(overallProgress), 100, barWidth)
	fmt.Printf("%s (%d/%d stages)\n\n", bar, completed, len(a.stages))
	
	// Render stage list
	for _, stage := range a.stages {
		symbol := a.renderer.colors.StatusSymbol(stage.Status)
		
		switch stage.Status {
		case "PASS", "FAIL", "SKIPPED":
			statusText := a.renderer.colors.StatusColor(stage.Status, stage.Status)
			fmt.Printf("%s %-15s %s\n", symbol, stage.ID, statusText)
		case "RUNNING":
			progress := CalculateStageProgress(stage.ElapsedSeconds, stage.EstimatedSeconds)
			progressText := fmt.Sprintf("%.0f%%", progress)
			fmt.Printf("%s %-15s %s %s\n", symbol, stage.ID, 
				a.renderer.colors.Blue("running..."), 
				a.renderer.colors.Gray(progressText))
		case "PENDING":
			fmt.Printf("%s %-15s %s\n", symbol, stage.ID, a.renderer.colors.Gray("pending"))
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
func (a *AnimatedStageTracker) renderFullMode() {
	// Calculate overall progress
	overallProgress := CalculateOverallProgress(a.stages)
	
	// Render overall progress bar
	barWidth := 40
	if a.renderer.width > 60 {
		barWidth = 60
	}
	
	bar := a.renderer.colors.ProgressBar(int(overallProgress), 100, barWidth)
	fmt.Printf("Overall: %s\n\n", bar)
	
	// Group stages by group
	groups := make(map[string][]StageProgress)
	groupOrder := []string{}
	seen := make(map[string]bool)
	
	for _, stage := range a.stages {
		if !seen[stage.Group] {
			groupOrder = append(groupOrder, stage.Group)
			seen[stage.Group] = true
		}
		groups[stage.Group] = append(groups[stage.Group], stage)
	}
	
	// Render each group
	for _, groupName := range groupOrder {
		stages := groups[groupName]
		
		// Group header
		headerText := fmt.Sprintf("─ %s ", strings.Title(groupName))
		padding := strings.Repeat("─", a.renderer.width-len(headerText)-2)
		fmt.Printf("┌%s%s┐\n", headerText, padding)
		
		// Stages in group
		for _, stage := range stages {
			symbol := a.renderer.colors.StatusSymbol(stage.Status)
			
			switch stage.Status {
			case "PASS":
				duration := FormatDuration(int64(stage.ElapsedSeconds * 1000))
				fmt.Printf("│ %s %-12s %s\n", symbol, stage.ID, 
					a.renderer.colors.Green(duration))
			case "FAIL":
				duration := FormatDuration(int64(stage.ElapsedSeconds * 1000))
				fmt.Printf("│ %s %-12s %s\n", symbol, stage.ID, 
					a.renderer.colors.Red(duration))
			case "SKIPPED":
				fmt.Printf("│ %s %-12s %s\n", symbol, stage.ID, 
					a.renderer.colors.Yellow("skipped"))
			case "RUNNING":
				progress := CalculateStageProgress(stage.ElapsedSeconds, stage.EstimatedSeconds)
				elapsed := FormatDuration(int64(stage.ElapsedSeconds * 1000))
				estimated := FormatDuration(int64(stage.EstimatedSeconds * 1000))
				
				// Mini progress bar
				miniBarWidth := 12
				miniBar := a.renderer.colors.ProgressBar(int(progress), 100, miniBarWidth)
				
				fmt.Printf("│ %s %-12s %s / %s   %s\n", symbol, stage.ID, 
					elapsed, estimated, miniBar)
			case "PENDING":
				fmt.Printf("│ %s %-12s %s\n", symbol, stage.ID, 
					a.renderer.colors.Gray("pending"))
			}
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
