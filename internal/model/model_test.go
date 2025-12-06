package model

import (
	"testing"
)

func TestTaskStatus(t *testing.T) {
	tests := []struct {
		name   string
		status TaskStatus
	}{
		{"pending", StatusPending},
		{"running", StatusRunning},
		{"pass", StatusPass},
		{"fail", StatusFail},
		{"skipped", StatusSkipped},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the constants are defined
			if string(tt.status) == "" {
				t.Errorf("Status %s is empty", tt.name)
			}
		})
	}
}

func TestTaskDefinition(t *testing.T) {
	task := TaskDefinition{
		ID:      "test-task",
		Name:    "Test Task",
		Command: "echo test",
		Type:    "test",
		Phase:   "1",
	}

	if task.ID != "test-task" {
		t.Errorf("Expected ID 'test-task', got '%s'", task.ID)
	}

	if task.Phase != "1" {
		t.Errorf("Expected phase '1', got '%s'", task.Phase)
	}
}

func TestEffectiveConfig(t *testing.T) {
	cfg := EffectiveConfig{
		Values: []ConfigValue{
			{
				Key:    "test.key",
				Value:  "test-value",
				Source: "config-file",
			},
			{
				Key:      "test.override",
				Value:    "new-value",
				Source:   "cli-flag",
				Overrode: "old-value",
			},
		},
	}

	if len(cfg.Values) != 2 {
		t.Errorf("Expected 2 config values, got %d", len(cfg.Values))
	}

	// Find the override
	var override *ConfigValue
	for i := range cfg.Values {
		if cfg.Values[i].Key == "test.override" {
			override = &cfg.Values[i]
			break
		}
	}

	if override == nil {
		t.Fatal("Expected to find override config value")
	}

	if override.Source != "cli-flag" {
		t.Errorf("Expected source 'cli-flag', got '%s'", override.Source)
	}

	if override.Overrode != "old-value" {
		t.Errorf("Expected overrode 'old-value', got '%s'", override.Overrode)
	}
}
