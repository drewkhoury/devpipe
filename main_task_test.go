package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/drew/devpipe/internal/config"
	"github.com/drew/devpipe/internal/model"
	"github.com/drew/devpipe/internal/ui"
)

func TestRunTask_DryRun(t *testing.T) {
	runDir := t.TempDir()
	logDir := filepath.Join(runDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	renderer := ui.NewRenderer(ui.UIModeBasic, false, false)

	task := model.TaskDefinition{
		ID:   "dry-run-task",
		Name: "Dry Run Task",
	}

	res, buf, err := runTask(task, runDir, logDir, true, false, renderer, nil, &sync.Mutex{}, nil, make(chan struct{}))
	if err != nil {
		t.Fatalf("runTask returned error in dry-run: %v", err)
	}

	if res.Status != model.StatusSkipped {
		t.Fatalf("expected status SKIPPED, got %s", res.Status)
	}
	if !res.Skipped || res.SkipReason != "dry-run" {
		t.Fatalf("expected dry-run skip, got skipped=%v reason=%q", res.Skipped, res.SkipReason)
	}
	if buf == nil {
		t.Fatalf("expected non-nil buffer")
	}
	if buf.Len() != 0 {
		t.Fatalf("expected empty buffer for dry-run, got %d bytes", buf.Len())
	}
}

func TestParseTaskMetrics_JUnit(t *testing.T) {
	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}

	task := model.TaskDefinition{
		ID:            "junit-task",
		Workdir:       repoRoot,
		MetricsFormat: "junit",
		MetricsPath:   "testdata/junit-single-suite.xml",
	}

	m := parseTaskMetrics(task, false)
	if m == nil {
		t.Fatalf("expected non-nil metrics for valid junit file")
	}
	if m.Kind != "test" {
		t.Fatalf("expected metrics kind 'test', got %q", m.Kind)
	}
	if m.SummaryFormat != "junit" {
		t.Fatalf("expected summary format 'junit', got %q", m.SummaryFormat)
	}
}

func TestParseTaskMetrics_FileNotFound(t *testing.T) {
	task := model.TaskDefinition{
		ID:            "missing-metrics",
		Workdir:       t.TempDir(),
		MetricsFormat: "junit",
		MetricsPath:   "does-not-exist.xml",
	}

	m := parseTaskMetrics(task, true)
	if m != nil {
		t.Fatalf("expected nil metrics when file is missing, got %#v", m)
	}
}

func TestCopyConfigToRun_WithExistingFile(t *testing.T) {
	runDir := t.TempDir()

	cfgDir := t.TempDir()
	configPath := filepath.Join(cfgDir, "config.toml")
	content := []byte("[defaults]\noutputRoot = \"out\"\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	merged := config.MergeWithDefaults(&config.Config{})

	if err := copyConfigToRun(runDir, configPath, &merged); err != nil {
		t.Fatalf("copyConfigToRun returned error: %v", err)
	}

	dest := filepath.Join(runDir, "config.toml")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("failed to read copied config: %v", err)
	}
	if string(data) != string(content) {
		t.Fatalf("copied config mismatch: got %q, want %q", string(data), string(content))
	}
}

func TestCopyConfigToRun_NoFile_WritesJSON(t *testing.T) {
	runDir := t.TempDir()

	merged := config.MergeWithDefaults(&config.Config{})

	if err := copyConfigToRun(runDir, "non-existent-config.toml", &merged); err != nil {
		t.Fatalf("copyConfigToRun returned error: %v", err)
	}

	jsonPath := filepath.Join(runDir, "config.json")
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("expected config.json to be written, got error: %v", err)
	}
}
