package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/drew/devpipe/assets"
	"github.com/drew/devpipe/internal/model"
)

// Summary holds aggregated data across all runs
type Summary struct {
	TotalRuns       int                  `json:"totalRuns"`
	RecentRuns      []RunSummary         `json:"recentRuns"`
	TaskStats       map[string]TaskStats `json:"taskStats"`       // All runs
	TaskStatsRecent map[string]TaskStats `json:"taskStatsRecent"` // Most recent run only
	TaskStatsLast25 map[string]TaskStats `json:"taskStatsLast25"` // Last 25 runs
	LastGenerated   string               `json:"lastGenerated"`
	Username        string               `json:"username"`
	Greeting        string               `json:"greeting"`
	Version         string               `json:"version"`
}

// RunSummary is a condensed view of a single run
type RunSummary struct {
	RunID      string `json:"runId"`
	Timestamp  string `json:"timestamp"`
	Status     string `json:"status"` // "PASS", "FAIL", "PARTIAL"
	Duration   int64  `json:"duration"`
	PassCount  int    `json:"passCount"`
	FailCount  int    `json:"failCount"`
	SkipCount  int    `json:"skipCount"`
	TotalTasks int    `json:"totalTasks"`
	Command    string `json:"command"` // Full command line that was executed
}

// TaskStats holds statistics for a specific task across runs
type TaskStats struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	TotalRuns   int     `json:"totalRuns"`
	PassCount   int     `json:"passCount"`
	FailCount   int     `json:"failCount"`
	SkipCount   int     `json:"skipCount"`
	AvgDuration float64 `json:"avgDuration"`
	LastStatus  string  `json:"lastStatus"`
}

// GenerateDashboard reads all runs and generates summary.json and report.html
func GenerateDashboard(outputRoot string) error {
	return GenerateDashboardWithVersion(outputRoot, "dev")
}

// GenerateDashboardWithVersion generates dashboard with version info
func GenerateDashboardWithVersion(outputRoot, version string) error {
	runsDir := filepath.Join(outputRoot, "runs")

	// Read all run.json files
	runs, err := loadAllRuns(runsDir)
	if err != nil {
		return fmt.Errorf("failed to load runs: %w", err)
	}

	// Aggregate data
	summary := aggregateRuns(runs, version)

	// Write summary.json
	summaryPath := filepath.Join(outputRoot, "summary.json")
	if err := writeSummaryJSON(summaryPath, summary); err != nil {
		return fmt.Errorf("failed to write summary.json: %w", err)
	}

	// Copy mascot image to output directory
	if err := copyMascotAssets(outputRoot); err != nil {
		// Don't fail if mascot copy fails, just warn
		fmt.Fprintf(os.Stderr, "WARNING: failed to copy mascot assets: %v\n", err)
	}

	// Generate HTML dashboard
	htmlPath := filepath.Join(outputRoot, "report.html")
	if err := writeHTMLDashboard(htmlPath, summary); err != nil {
		return fmt.Errorf("failed to write report.html: %w", err)
	}

	// Generate individual run detail pages
	for _, run := range runs {
		detailPath := filepath.Join(outputRoot, "runs", run.RunID, "report.html")
		if err := writeRunDetailHTML(detailPath, run); err != nil {
			// Don't fail if one detail page fails, but log it
			fmt.Fprintf(os.Stderr, "WARNING: failed to generate report for run %s: %v\n", run.RunID, err)
			continue
		}
	}

	return nil
}

// loadAllRuns reads all run.json files from the runs directory
func loadAllRuns(runsDir string) ([]model.RunRecord, error) {
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.RunRecord{}, nil
		}
		return nil, err
	}

	var runs []model.RunRecord
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		runJSONPath := filepath.Join(runsDir, entry.Name(), "run.json")
		data, err := os.ReadFile(runJSONPath)
		if err != nil {
			continue // Skip if can't read
		}

		var run model.RunRecord
		if err := json.Unmarshal(data, &run); err != nil {
			continue // Skip if can't parse
		}

		runs = append(runs, run)
	}

	// Sort by timestamp (newest first)
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].Timestamp > runs[j].Timestamp
	})

	return runs, nil
}

// aggregateRuns creates a summary from all runs
func aggregateRuns(runs []model.RunRecord, version string) Summary {
	// Get username
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME") // Windows fallback
	}
	if username == "" {
		username = "friend"
	}
	
	// Random greeting
	greetings := []string{
		"Hello",
		"Hi",
		"Hey",
		"Howdy",
		"Greetings",
		"Welcome back",
		"Good to see you",
		"Hey there",
		"Ahoy",
		"Yo",
	}
	greeting := greetings[time.Now().Unix()%int64(len(greetings))]
	
	summary := Summary{
		TotalRuns:       len(runs),
		RecentRuns:      []RunSummary{},
		TaskStats:       make(map[string]TaskStats),
		TaskStatsRecent: make(map[string]TaskStats),
		TaskStatsLast25: make(map[string]TaskStats),
		LastGenerated:   time.Now().UTC().Format(time.RFC3339),
		Username:        username,
		Greeting:        greeting,
		Version:         version,
	}

	// Add recent runs (limit to 100 for pagination)
	for i, run := range runs {
		if i < 100 {
			runSummary := summarizeRun(run)
			summary.RecentRuns = append(summary.RecentRuns, runSummary)
		}
	}

	// Calculate task stats for different ranges
	summary.TaskStats = calculateTaskStats(runs, len(runs))                // All runs
	summary.TaskStatsRecent = calculateTaskStats(runs, 1)                  // Most recent run
	summary.TaskStatsLast25 = calculateTaskStats(runs, min(25, len(runs))) // Last 25 runs

	return summary
}

// calculateTaskStats aggregates task statistics for a given number of recent runs
func calculateTaskStats(runs []model.RunRecord, numRuns int) map[string]TaskStats {
	taskStats := make(map[string]TaskStats)
	taskDurations := make(map[string][]int64)

	// Process only the specified number of runs
	for i := 0; i < numRuns && i < len(runs); i++ {
		run := runs[i]

		for _, task := range run.Tasks {
			stats, exists := taskStats[task.ID]
			if !exists {
				stats = TaskStats{
					ID:   task.ID,
					Name: task.Name,
				}
			}

			stats.TotalRuns++
			switch task.Status {
			case model.StatusPass:
				stats.PassCount++
			case model.StatusFail:
				stats.FailCount++
			case model.StatusSkipped:
				stats.SkipCount++
			}

			// Track duration for average
			if !task.Skipped {
				taskDurations[task.ID] = append(taskDurations[task.ID], task.DurationMs)
			}

			// Update last status (from most recent run in this range)
			if i == 0 {
				stats.LastStatus = string(task.Status)
			}

			taskStats[task.ID] = stats
		}
	}

	// Calculate average durations
	for id, durations := range taskDurations {
		if len(durations) > 0 {
			var sum int64
			for _, d := range durations {
				sum += d
			}
			stats := taskStats[id]
			stats.AvgDuration = float64(sum) / float64(len(durations))
			taskStats[id] = stats
		}
	}

	return taskStats
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// summarizeRun creates a RunSummary from a RunRecord
func summarizeRun(run model.RunRecord) RunSummary {
	summary := RunSummary{
		RunID:      run.RunID,
		Timestamp:  run.Timestamp,
		TotalTasks: len(run.Tasks),
		Command:    cleanCommand(run.Command),
	}

	anyFailed := false
	var totalDuration int64

	for _, task := range run.Tasks {
		totalDuration += task.DurationMs

		switch task.Status {
		case model.StatusPass:
			summary.PassCount++
		case model.StatusFail:
			summary.FailCount++
			anyFailed = true
		case model.StatusSkipped:
			summary.SkipCount++
		}
	}

	summary.Duration = totalDuration

	if anyFailed {
		summary.Status = "FAIL"
	} else if summary.SkipCount == summary.TotalTasks {
		summary.Status = "SKIPPED"
	} else {
		summary.Status = "PASS"
	}

	return summary
}

// writeSummaryJSON writes the summary to a JSON file
func writeSummaryJSON(path string, summary Summary) error {
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// cleanCommand removes shell prompt cruft from old command strings
// Old format: "drew@drews-MBP.attlocal.net devpipe % ./devpipe --config ..."
// New format: "./devpipe --config ..."
func cleanCommand(cmd string) string {
	// Look for the pattern: username@hostname directory % command
	// We want to extract just the command part after the %
	for i := 0; i < len(cmd); i++ {
		if cmd[i] == '%' && i+1 < len(cmd) && cmd[i+1] == ' ' {
			// Found "% ", return everything after it
			return cmd[i+2:]
		}
	}
	// No cruft found, return as-is
	return cmd
}

// copyMascotAssets writes the embedded mascot image to the output directory
func copyMascotAssets(outputRoot string) error {
	mascotDir := filepath.Join(outputRoot, "mascot")
	if err := os.MkdirAll(mascotDir, 0755); err != nil {
		return fmt.Errorf("failed to create mascot directory: %w", err)
	}

	mascotPath := filepath.Join(mascotDir, "squirrel-blank-eyes-transparent-cropped.png")
	if err := os.WriteFile(mascotPath, assets.MascotImage, 0644); err != nil {
		return fmt.Errorf("failed to write mascot image: %w", err)
	}

	return nil
}
