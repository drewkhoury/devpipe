package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/drew/devpipe/internal/model"
)

// Summary holds aggregated data across all runs
type Summary struct {
	TotalRuns      int            `json:"totalRuns"`
	RecentRuns     []RunSummary   `json:"recentRuns"`
	StageStats     map[string]StageStats `json:"stageStats"`
	LastGenerated  string         `json:"lastGenerated"`
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
	TotalStages int   `json:"totalStages"`
}

// StageStats holds statistics for a specific stage across runs
type StageStats struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	TotalRuns     int     `json:"totalRuns"`
	PassCount     int     `json:"passCount"`
	FailCount     int     `json:"failCount"`
	SkipCount     int     `json:"skipCount"`
	AvgDuration   float64 `json:"avgDuration"`
	LastStatus    string  `json:"lastStatus"`
}

// GenerateDashboard reads all runs and generates summary.json and report.html
func GenerateDashboard(outputRoot string) error {
	runsDir := filepath.Join(outputRoot, "runs")
	
	// Read all run.json files
	runs, err := loadAllRuns(runsDir)
	if err != nil {
		return fmt.Errorf("failed to load runs: %w", err)
	}
	
	// Generate summary
	summary := aggregateRuns(runs)
	
	// Write summary.json
	summaryPath := filepath.Join(outputRoot, "summary.json")
	if err := writeSummaryJSON(summaryPath, summary); err != nil {
		return fmt.Errorf("failed to write summary.json: %w", err)
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
			// Don't fail if one detail page fails
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
func aggregateRuns(runs []model.RunRecord) Summary {
	summary := Summary{
		TotalRuns:     len(runs),
		RecentRuns:    []RunSummary{},
		StageStats:    make(map[string]StageStats),
		LastGenerated: time.Now().UTC().Format(time.RFC3339),
	}
	
	// Track stage statistics
	stageDurations := make(map[string][]int64)
	
	// Process each run
	for i, run := range runs {
		// Add to recent runs (limit to 20)
		if i < 20 {
			runSummary := summarizeRun(run)
			summary.RecentRuns = append(summary.RecentRuns, runSummary)
		}
		
		// Aggregate stage stats
		for _, stage := range run.Stages {
			stats, exists := summary.StageStats[stage.ID]
			if !exists {
				stats = StageStats{
					ID:   stage.ID,
					Name: stage.Name,
				}
			}
			
			stats.TotalRuns++
			switch stage.Status {
			case model.StatusPass:
				stats.PassCount++
			case model.StatusFail:
				stats.FailCount++
			case model.StatusSkipped:
				stats.SkipCount++
			}
			
			// Track duration for average
			if !stage.Skipped {
				stageDurations[stage.ID] = append(stageDurations[stage.ID], stage.DurationMs)
			}
			
			// Update last status (from most recent run)
			if i == 0 {
				stats.LastStatus = string(stage.Status)
			}
			
			summary.StageStats[stage.ID] = stats
		}
	}
	
	// Calculate average durations
	for id, durations := range stageDurations {
		if len(durations) > 0 {
			var sum int64
			for _, d := range durations {
				sum += d
			}
			stats := summary.StageStats[id]
			stats.AvgDuration = float64(sum) / float64(len(durations))
			summary.StageStats[id] = stats
		}
	}
	
	return summary
}

// summarizeRun creates a RunSummary from a RunRecord
func summarizeRun(run model.RunRecord) RunSummary {
	summary := RunSummary{
		RunID:       run.RunID,
		Timestamp:   run.Timestamp,
		TotalStages: len(run.Stages),
	}
	
	anyFailed := false
	var totalDuration int64
	
	for _, stage := range run.Stages {
		totalDuration += stage.DurationMs
		
		switch stage.Status {
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
	} else if summary.SkipCount == summary.TotalStages {
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
