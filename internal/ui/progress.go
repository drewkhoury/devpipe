package ui

import (
	"time"
)

// StageProgress represents the progress of a single stage
type StageProgress struct {
	ID               string
	Name             string
	Group            string
	Status           string
	EstimatedSeconds int
	ElapsedSeconds   float64
	StartTime        time.Time
}

// CalculateStageProgress returns the progress percentage for a stage (0-100)
func CalculateStageProgress(elapsed float64, estimated int) float64 {
	if estimated == 0 {
		return 0
	}
	
	progress := (elapsed / float64(estimated)) * 100
	if progress > 100 {
		return 100
	}
	return progress
}

// CalculateOverallProgress calculates overall pipeline progress based on stage weights
func CalculateOverallProgress(stages []StageProgress) float64 {
	if len(stages) == 0 {
		return 0
	}
	
	totalWeight := 0
	completedWeight := 0
	
	for _, stage := range stages {
		weight := stage.EstimatedSeconds
		if weight == 0 {
			weight = 10 // Default weight
		}
		
		totalWeight += weight
		
		switch stage.Status {
		case "PASS", "FAIL", "SKIPPED":
			// Stage complete
			completedWeight += weight
		case "RUNNING":
			// Stage in progress
			stageProgress := CalculateStageProgress(stage.ElapsedSeconds, stage.EstimatedSeconds)
			completedWeight += int(float64(weight) * (stageProgress / 100.0))
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
