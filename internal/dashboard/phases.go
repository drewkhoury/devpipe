package dashboard

import (
	"github.com/drew/devpipe/internal/config"
	"github.com/drew/devpipe/internal/model"
)

// PhaseGroup represents a phase with its tasks
type PhaseGroup struct {
	ID        string
	Name      string
	Desc      string
	Tasks     []TaskWithDesc
	Status    string // "PASS" or "FAIL"
	TotalMs   int64
	TaskCount int
}

// TaskWithDesc is an alias for TaskResult (desc is now in TaskResult)
type TaskWithDesc = model.TaskResult

// ParsePhasesFromConfig groups tasks by their Phase field from execution
func ParsePhasesFromConfig(configPath string, tasks []model.TaskResult) ([]PhaseGroup, error) {
	// Load config to get phase descriptions and IDs
	_, _, phaseInfoMap, _, err := config.LoadConfig(configPath)
	phaseDescMap := make(map[string]string)
	phaseIDMap := make(map[string]string)
	if err == nil {
		// Build maps: phase name -> description and phase name -> ID
		for _, phaseInfo := range phaseInfoMap {
			phaseDescMap[phaseInfo.Name] = phaseInfo.Desc
			phaseIDMap[phaseInfo.Name] = phaseInfo.ID
		}
	}

	// Group tasks by phase in execution order
	phaseMap := make(map[string]*PhaseGroup)
	phaseOrder := []string{}

	for _, task := range tasks {
		phaseName := task.Phase
		if phaseName == "" {
			phaseName = "Tasks"
		}

		// Create phase if it doesn't exist
		if _, exists := phaseMap[phaseName]; !exists {
			phaseMap[phaseName] = &PhaseGroup{
				ID:     phaseIDMap[phaseName],
				Name:   phaseName,
				Desc:   phaseDescMap[phaseName],
				Tasks:  []model.TaskResult{},
				Status: "PASS",
			}
			phaseOrder = append(phaseOrder, phaseName)
		}

		// Add task to phase
		phase := phaseMap[phaseName]
		phase.Tasks = append(phase.Tasks, task)
		phase.TotalMs += task.DurationMs
		phase.TaskCount++
		if task.Status == model.StatusFail {
			phase.Status = "FAIL"
		}
	}

	// Build final phase list in order
	var phases []PhaseGroup
	for _, phaseName := range phaseOrder {
		phases = append(phases, *phaseMap[phaseName])
	}

	return phases, nil
}
