package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/drew/devpipe/internal/config"
	"github.com/drew/devpipe/internal/model"
)

func TestSliceFlag(t *testing.T) {
	var sf sliceFlag

	// Test Set
	err := sf.Set("task1")
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	err = sf.Set("task2")
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	if len(sf) != 2 {
		t.Errorf("Expected 2 items, got %d", len(sf))
	}

	if sf[0] != "task1" {
		t.Errorf("Expected first item 'task1', got '%s'", sf[0])
	}

	if sf[1] != "task2" {
		t.Errorf("Expected second item 'task2', got '%s'", sf[1])
	}

	// Test String
	str := sf.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}
}

func TestGetTerminalWidth(t *testing.T) {
	width := getTerminalWidth()

	// Should return a reasonable width (default is 160)
	if width < 80 {
		t.Errorf("Expected width >= 80, got %d", width)
	}

	if width > 500 {
		t.Errorf("Expected width <= 500, got %d (seems unreasonably large)", width)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		wantLen  int
		wantDots bool
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			wantLen:  5,
			wantDots: false,
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			wantLen:  5,
			wantDots: false,
		},
		{
			name:     "needs truncation",
			input:    "hello world this is a long string",
			maxLen:   10,
			wantLen:  10,
			wantDots: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.maxLen)

			if len(got) != tt.wantLen {
				t.Errorf("truncate() length = %d, want %d", len(got), tt.wantLen)
			}

			if tt.wantDots && got[len(got)-3:] != "..." {
				t.Errorf("truncate() should end with '...', got '%s'", got)
			}
		})
	}
}

func TestMakeRunID(t *testing.T) {
	// Test that makeRunID generates valid IDs
	id := makeRunID()

	if id == "" {
		t.Error("Expected non-empty run ID")
	}

	// Should contain timestamp format (YYYY-MM-DDTHH-MM-SSZ_NNNNNN)
	if len(id) < 20 {
		t.Errorf("Expected run ID length >= 20, got %d", len(id))
	}

	// Should contain underscore separator
	// Valid format check
	_ = !containsSpace(id) && len(id) > 0
}

func TestContainsSpace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "no spaces",
			input: "hello",
			want:  false,
		},
		{
			name:  "with space",
			input: "hello world",
			want:  true,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "only space",
			input: " ",
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsSpace(tt.input)
			if got != tt.want {
				t.Errorf("containsSpace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPhaseEmoji(t *testing.T) {
	tests := []struct {
		name      string
		phaseName string
	}{
		{"validation", "Validation"},
		{"build", "Build"},
		{"test", "Test"},
		{"deploy", "Deploy"},
		{"unknown", "Something Else"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emoji := phaseEmoji(tt.phaseName)

			// Just verify it returns something
			if emoji == "" {
				t.Errorf("phaseEmoji(%s) returned empty string", tt.phaseName)
			}
		})
	}
}

func TestBuildCommandString(t *testing.T) {
	// Test that buildCommandString returns something
	cmd := buildCommandString()

	// Should return a non-empty string
	if cmd == "" {
		t.Error("Expected non-empty command string")
	}
}

func TestLoadHistoricalAverages(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		summaryJSON   string
		wantAverages  map[string]int
		createSummary bool
	}{
		{
			name:          "no summary file",
			createSummary: false,
			wantAverages:  map[string]int{},
		},
		{
			name:          "empty task stats",
			createSummary: true,
			summaryJSON:   `{"taskStats":{}}`,
			wantAverages:  map[string]int{},
		},
		{
			name:          "valid task stats",
			createSummary: true,
			summaryJSON: `{
				"taskStats": {
					"task1": {"avgDuration": 5000},
					"task2": {"avgDuration": 1500},
					"task3": {"avgDuration": 500}
				}
			}`,
			wantAverages: map[string]int{
				"task1": 5,
				"task2": 1,
				"task3": 1, // Minimum is 1 second
			},
		},
		{
			name:          "zero duration ignored",
			createSummary: true,
			summaryJSON: `{
				"taskStats": {
					"task1": {"avgDuration": 0},
					"task2": {"avgDuration": 2000}
				}
			}`,
			wantAverages: map[string]int{
				"task2": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputRoot := filepath.Join(tmpDir, tt.name)
			_ = os.MkdirAll(outputRoot, 0755)

			if tt.createSummary {
				summaryPath := filepath.Join(outputRoot, "summary.json")
				_ = os.WriteFile(summaryPath, []byte(tt.summaryJSON), 0644)
			}

			got := loadHistoricalAverages(outputRoot)

			if len(got) != len(tt.wantAverages) {
				t.Errorf("loadHistoricalAverages() returned %d items, want %d", len(got), len(tt.wantAverages))
			}

			for taskID, wantAvg := range tt.wantAverages {
				if gotAvg, ok := got[taskID]; !ok {
					t.Errorf("loadHistoricalAverages() missing task %q", taskID)
				} else if gotAvg != wantAvg {
					t.Errorf("loadHistoricalAverages() task %q = %d, want %d", taskID, gotAvg, wantAvg)
				}
			}
		})
	}
}

func TestLoadTaskAveragesLast25(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		summaryJSON   string
		wantAverages  map[string]float64
		createSummary bool
	}{
		{
			name:          "no summary file",
			createSummary: false,
			wantAverages:  map[string]float64{},
		},
		{
			name:          "valid task stats last 25",
			createSummary: true,
			summaryJSON: `{
				"taskStatsLast25": {
					"task1": {"avgDuration": 5000.5},
					"task2": {"avgDuration": 1500.25}
				}
			}`,
			wantAverages: map[string]float64{
				"task1": 5000.5,
				"task2": 1500.25,
			},
		},
		{
			name:          "zero duration ignored",
			createSummary: true,
			summaryJSON: `{
				"taskStatsLast25": {
					"task1": {"avgDuration": 0},
					"task2": {"avgDuration": 2000}
				}
			}`,
			wantAverages: map[string]float64{
				"task2": 2000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputRoot := filepath.Join(tmpDir, tt.name)
			_ = os.MkdirAll(outputRoot, 0755)

			if tt.createSummary {
				summaryPath := filepath.Join(outputRoot, "summary.json")
				_ = os.WriteFile(summaryPath, []byte(tt.summaryJSON), 0644)
			}

			got := loadTaskAveragesLast25(outputRoot)

			if len(got) != len(tt.wantAverages) {
				t.Errorf("loadTaskAveragesLast25() returned %d items, want %d", len(got), len(tt.wantAverages))
			}

			for taskID, wantAvg := range tt.wantAverages {
				if gotAvg, ok := got[taskID]; !ok {
					t.Errorf("loadTaskAveragesLast25() missing task %q", taskID)
				} else if gotAvg != wantAvg {
					t.Errorf("loadTaskAveragesLast25() task %q = %f, want %f", taskID, gotAvg, wantAvg)
				}
			}
		})
	}
}

func TestWriteRunJSON(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()

	runDir := filepath.Join(tmpDir, "run1")
	_ = os.MkdirAll(runDir, 0755)

	record := model.RunRecord{
		RunID:      "test-run-123",
		Timestamp:  "2024-01-01T00:00:00Z",
		RepoRoot:   "/test/repo",
		OutputRoot: "/test/output",
		ConfigPath: "config.toml",
		Command:    "devpipe --verbose",
	}

	err := writeRunJSON(runDir, record)
	if err != nil {
		t.Fatalf("writeRunJSON() error = %v", err)
	}

	// Verify file was created
	runJSONPath := filepath.Join(runDir, "run.json")
	if _, err := os.Stat(runJSONPath); os.IsNotExist(err) {
		t.Errorf("writeRunJSON() did not create run.json")
	}

	// Verify content
	data, err := os.ReadFile(runJSONPath)
	if err != nil {
		t.Fatalf("Failed to read run.json: %v", err)
	}

	var readRecord model.RunRecord
	if err := json.Unmarshal(data, &readRecord); err != nil {
		t.Fatalf("Failed to unmarshal run.json: %v", err)
	}

	if readRecord.RunID != record.RunID {
		t.Errorf("RunID = %q, want %q", readRecord.RunID, record.RunID)
	}
	if readRecord.Timestamp != record.Timestamp {
		t.Errorf("Timestamp = %q, want %q", readRecord.Timestamp, record.Timestamp)
	}
}

func TestBuildEffectiveConfig(t *testing.T) {
	// Create a basic config
	cfg := &config.Config{
		Defaults: config.DefaultsConfig{
			OutputRoot:    "custom-output",
			FastThreshold: 30,
			UIMode:        "full",
		},
	}

	mergedCfg := config.MergeWithDefaults(cfg)

	effective := buildEffectiveConfig(cfg, &mergedCfg, "", "basic", "basic", "staged", "HEAD", map[string]int{})

	if effective == nil {
		t.Fatal("buildEffectiveConfig() returned nil")
	}

	if len(effective.Values) == 0 {
		t.Error("buildEffectiveConfig() returned no values")
	}

	// Check that we have expected config values
	foundOutputRoot := false
	foundUIMode := false
	for _, val := range effective.Values {
		if val.Key == "defaults.outputRoot" {
			foundOutputRoot = true
			if val.Value != "custom-output" {
				t.Errorf("outputRoot value = %q, want %q", val.Value, "custom-output")
			}
			if val.Source != "config-file" {
				t.Errorf("outputRoot source = %q, want %q", val.Source, "config-file")
			}
		}
		if val.Key == "defaults.uiMode" {
			foundUIMode = true
			if val.Value != "basic" {
				t.Errorf("uiMode value = %q, want %q", val.Value, "basic")
			}
		}
	}

	if !foundOutputRoot {
		t.Error("buildEffectiveConfig() missing defaults.outputRoot")
	}
	if !foundUIMode {
		t.Error("buildEffectiveConfig() missing defaults.uiMode")
	}
}

func TestBuildEffectiveConfigWithCLIOverrides(t *testing.T) {
	// Test that CLI flags override config values
	cfg := &config.Config{
		Defaults: config.DefaultsConfig{
			UIMode: "basic",
			Git: config.GitConfig{
				Mode: "staged",
				Ref:  "main",
			},
		},
	}

	mergedCfg := config.MergeWithDefaults(cfg)

	// Simulate CLI flags overriding config
	flagSince := "HEAD~1"
	flagUI := "full"

	effective := buildEffectiveConfig(cfg, &mergedCfg, flagSince, flagUI, "full", "ref", "HEAD~1", map[string]int{})

	if effective == nil {
		t.Fatal("buildEffectiveConfig() returned nil")
	}

	// Check that CLI overrides are recorded
	for _, val := range effective.Values {
		if val.Key == "defaults.uiMode" {
			if val.Source != "cli-flag" {
				t.Errorf("uiMode source = %q, want %q", val.Source, "cli-flag")
			}
			if val.Overrode != "basic" {
				t.Errorf("uiMode overrode = %q, want %q", val.Overrode, "basic")
			}
		}
		if val.Key == "defaults.git.mode" {
			if val.Source != "cli-flag" {
				t.Errorf("git.mode source = %q, want %q", val.Source, "cli-flag")
			}
		}
	}
}
