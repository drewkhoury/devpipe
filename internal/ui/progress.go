package ui

import (
	"time"
)

// TaskProgress represents the progress of a single task
type TaskProgress struct {
	ID               string
	Name             string
	Type             string // Type of task (quality, correctness, release)
	Phase            int    // Phase number (1, 2, 3, etc.)
	PhaseName        string // Display name for the phase
	Status           string
	EstimatedSeconds int
	IsEstimateGuess  bool // True if estimate is a default guess
	ElapsedSeconds   float64
	StartTime        time.Time
}

// CalculateTaskProgress returns the progress percentage for a task (0-100)
func CalculateTaskProgress(elapsed float64, estimated int) float64 {
	if estimated == 0 {
		return 0
	}
	
	progress := (elapsed / float64(estimated)) * 100
	if progress > 100 {
		return 100
	}
	return progress
}

// CalculateOverallProgress calculates overall pipeline progress based on task weights
func CalculateOverallProgress(tasks []TaskProgress) float64 {
	if len(tasks) == 0 {
		return 0
	}
	
	totalWeight := 0
	completedWeight := 0
	
	for _, task := range tasks {
		weight := task.EstimatedSeconds
		if weight == 0 {
			weight = 10 // Default weight
		}
		
		totalWeight += weight
		
		switch task.Status {
		case "PASS", "FAIL", "SKIPPED":
			// Task complete
			completedWeight += weight
		case "RUNNING":
			// Task in progress
			taskProgress := CalculateTaskProgress(task.ElapsedSeconds, task.EstimatedSeconds)
			completedWeight += int(float64(weight) * (taskProgress / 100.0))
		case "PENDING":
			// Not started yet
			completedWeight += 0
		}
	}
	
	if totalWeight == 0 {
		return 0
	}
	
	return (float64(completedWeight) / float64(totalWeight)) * 100
}

// FormatDuration formats a duration in milliseconds to a human-readable string
func FormatDuration(ms int64) string {
	if ms < 1000 {
		return "0s"
	}
	
	seconds := ms / 1000
	
	if seconds < 60 {
		return formatSeconds(seconds)
	}
	
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	
	if minutes < 60 {
		if remainingSeconds == 0 {
			return formatMinutes(minutes)
		}
		return formatMinutes(minutes) + " " + formatSeconds(remainingSeconds)
	}
	
	hours := minutes / 60
	remainingMinutes := minutes % 60
	
	if remainingMinutes == 0 {
		return formatHours(hours)
	}
	return formatHours(hours) + " " + formatMinutes(remainingMinutes)
}

func formatSeconds(s int64) string {
	if s == 1 {
		return "1s"
	}
	return formatInt(s) + "s"
}

func formatMinutes(m int64) string {
	if m == 1 {
		return "1m"
	}
	return formatInt(m) + "m"
}

func formatHours(h int64) string {
	if h == 1 {
		return "1h"
	}
	return formatInt(h) + "h"
}

func formatInt(n int64) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0' + n/10)) + string(rune('0' + n%10))
}
