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

func TestRunTask_Success(t *testing.T) {
	runDir := t.TempDir()
	logDir := filepath.Join(runDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	renderer := ui.NewRenderer(ui.UIModeBasic, false, false)

	task := model.TaskDefinition{
		ID:      "success-task",
		Name:    "Success Task",
		Command: "echo 'test output'",
		Workdir: runDir,
	}

	taskDone := make(chan struct{})
	res, buf, err := runTask(task, runDir, logDir, false, false, renderer, nil, &sync.Mutex{}, nil, taskDone)
	if err != nil {
		t.Fatalf("runTask returned error: %v", err)
	}

	if res.Status != model.StatusPass {
		t.Fatalf("expected status PASS, got %s", res.Status)
	}
	if res.ExitCode == nil || *res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %v", res.ExitCode)
	}
	if res.DurationMs <= 0 {
		t.Fatalf("expected positive duration, got %d", res.DurationMs)
	}
	if buf == nil {
		t.Fatalf("expected non-nil buffer")
	}

	// Verify log file was created
	logPath := filepath.Join(logDir, "success-task.log")
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log file not created: %v", err)
	}
}

func TestRunTask_Failure(t *testing.T) {
	runDir := t.TempDir()
	logDir := filepath.Join(runDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	renderer := ui.NewRenderer(ui.UIModeBasic, false, false)

	task := model.TaskDefinition{
		ID:      "fail-task",
		Name:    "Failing Task",
		Command: "exit 42",
		Workdir: runDir,
	}

	taskDone := make(chan struct{})
	res, buf, err := runTask(task, runDir, logDir, false, false, renderer, nil, &sync.Mutex{}, nil, taskDone)

	// runTask returns error when command fails
	if err == nil {
		t.Fatalf("expected error from failing command, got nil")
	}

	if res.Status != model.StatusFail {
		t.Fatalf("expected status FAIL, got %s", res.Status)
	}
	if res.ExitCode == nil || *res.ExitCode != 42 {
		t.Fatalf("expected exit code 42, got %v", res.ExitCode)
	}
	if buf == nil {
		t.Fatalf("expected non-nil buffer")
	}
}

func TestRunTask_CustomWorkdir(t *testing.T) {
	runDir := t.TempDir()
	logDir := filepath.Join(runDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	// Create a custom workdir with a test file
	customWorkdir := filepath.Join(runDir, "custom")
	if err := os.MkdirAll(customWorkdir, 0o755); err != nil {
		t.Fatalf("failed to create custom workdir: %v", err)
	}
	testFile := filepath.Join(customWorkdir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	renderer := ui.NewRenderer(ui.UIModeBasic, false, false)

	task := model.TaskDefinition{
		ID:      "workdir-task",
		Name:    "Custom Workdir Task",
		Command: "test -f test.txt",
		Workdir: customWorkdir,
	}

	taskDone := make(chan struct{})
	res, _, err := runTask(task, runDir, logDir, false, false, renderer, nil, &sync.Mutex{}, nil, taskDone)
	if err != nil {
		t.Fatalf("runTask returned error: %v", err)
	}

	if res.Status != model.StatusPass {
		t.Fatalf("expected status PASS (file should exist in custom workdir), got %s", res.Status)
	}
}

func TestRunTask_OutputBuffering_NonAnimated(t *testing.T) {
	runDir := t.TempDir()
	logDir := filepath.Join(runDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	renderer := ui.NewRenderer(ui.UIModeBasic, false, false)

	task := model.TaskDefinition{
		ID:      "output-task",
		Name:    "Output Task",
		Command: "echo 'line1' && echo 'line2'",
		Workdir: runDir,
	}

	taskDone := make(chan struct{})
	res, buf, err := runTask(task, runDir, logDir, false, false, renderer, nil, &sync.Mutex{}, nil, taskDone)
	if err != nil {
		t.Fatalf("runTask returned error: %v", err)
	}

	if res.Status != model.StatusPass {
		t.Fatalf("expected status PASS, got %s", res.Status)
	}

	// In non-animated mode, buffer should be minimal (just status messages)
	if buf == nil {
		t.Fatalf("expected non-nil buffer")
	}
}

func TestRunTask_LogFileCreation(t *testing.T) {
	runDir := t.TempDir()
	logDir := filepath.Join(runDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	renderer := ui.NewRenderer(ui.UIModeBasic, false, false)

	task := model.TaskDefinition{
		ID:      "log-task",
		Name:    "Log Task",
		Command: "echo 'logged output'",
		Workdir: runDir,
	}

	taskDone := make(chan struct{})
	res, _, err := runTask(task, runDir, logDir, false, false, renderer, nil, &sync.Mutex{}, nil, taskDone)
	if err != nil {
		t.Fatalf("runTask returned error: %v", err)
	}

	if res.Status != model.StatusPass {
		t.Fatalf("expected status PASS, got %s", res.Status)
	}

	// Verify log file exists and contains output
	logPath := filepath.Join(logDir, "log-task.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Fatalf("log file is empty")
	}
}

func TestRunTask_VerboseMode(t *testing.T) {
	runDir := t.TempDir()
	logDir := filepath.Join(runDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("failed to create log dir: %v", err)
	}

	renderer := ui.NewRenderer(ui.UIModeBasic, false, false)

	task := model.TaskDefinition{
		ID:      "verbose-task",
		Name:    "Verbose Task",
		Command: "echo 'verbose output'",
		Workdir: runDir,
	}

	taskDone := make(chan struct{})
	res, _, err := runTask(task, runDir, logDir, false, true, renderer, nil, &sync.Mutex{}, nil, taskDone)
	if err != nil {
		t.Fatalf("runTask returned error: %v", err)
	}

	if res.Status != model.StatusPass {
		t.Fatalf("expected status PASS, got %s", res.Status)
	}
}

func TestParseTaskMetrics_JUnit(t *testing.T) {
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}

	task := model.TaskDefinition{
		ID:            "junit-task",
		Workdir:       projectRoot,
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

func TestParseTaskMetrics_SARIF(t *testing.T) {
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}

	task := model.TaskDefinition{
		ID:            "sarif-task",
		Workdir:       projectRoot,
		MetricsFormat: "sarif",
		MetricsPath:   "testdata/sarif-sample.json",
	}

	m := parseTaskMetrics(task, false)
	if m == nil {
		t.Fatalf("expected non-nil metrics for valid SARIF file")
	}
	if m.Kind != "security" {
		t.Fatalf("expected metrics kind 'security', got %q", m.Kind)
	}
	if m.SummaryFormat != "sarif" {
		t.Fatalf("expected summary format 'sarif', got %q", m.SummaryFormat)
	}
}

func TestParseTaskMetrics_Artifact(t *testing.T) {
	tempDir := t.TempDir()

	// Create a dummy artifact file
	artifactPath := filepath.Join(tempDir, "artifact.txt")
	if err := os.WriteFile(artifactPath, []byte("test artifact"), 0o644); err != nil {
		t.Fatalf("failed to create artifact file: %v", err)
	}

	task := model.TaskDefinition{
		ID:            "artifact-task",
		Workdir:       tempDir,
		MetricsFormat: "artifact",
		MetricsPath:   "artifact.txt",
	}

	m := parseTaskMetrics(task, false)
	if m == nil {
		t.Fatalf("expected non-nil metrics for artifact format")
	}
	if m.Kind != "artifact" {
		t.Fatalf("expected metrics kind 'artifact', got %q", m.Kind)
	}
	if m.SummaryFormat != "artifact" {
		t.Fatalf("expected summary format 'artifact', got %q", m.SummaryFormat)
	}

	// Verify path is stored in data
	path, ok := m.Data["path"].(string)
	if !ok {
		t.Fatalf("expected path in metrics data")
	}
	if path != artifactPath {
		t.Fatalf("expected path %q, got %q", artifactPath, path)
	}
}

func TestParseTaskMetrics_UnknownFormat(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file (format doesn't matter, we're testing unknown format handling)
	filePath := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	task := model.TaskDefinition{
		ID:            "unknown-format-task",
		Workdir:       tempDir,
		MetricsFormat: "unknown-format",
		MetricsPath:   "test.txt",
	}

	m := parseTaskMetrics(task, false)
	if m != nil {
		t.Fatalf("expected nil metrics for unknown format, got %#v", m)
	}
}

func TestParseTaskMetrics_MalformedJUnit(t *testing.T) {
	tempDir := t.TempDir()

	// Create malformed JUnit XML
	malformedPath := filepath.Join(tempDir, "malformed.xml")
	malformedContent := []byte("<not-valid-junit>this is not valid XML</broken>")
	if err := os.WriteFile(malformedPath, malformedContent, 0o644); err != nil {
		t.Fatalf("failed to create malformed file: %v", err)
	}

	task := model.TaskDefinition{
		ID:            "malformed-junit-task",
		Workdir:       tempDir,
		MetricsFormat: "junit",
		MetricsPath:   "malformed.xml",
	}

	m := parseTaskMetrics(task, false)
	if m != nil {
		t.Fatalf("expected nil metrics for malformed JUnit XML, got %#v", m)
	}
}

func TestParseTaskMetrics_MalformedSARIF(t *testing.T) {
	tempDir := t.TempDir()

	// Create malformed SARIF JSON
	malformedPath := filepath.Join(tempDir, "malformed.sarif")
	malformedContent := []byte(`{"not": "valid", "sarif": [broken json}`)
	if err := os.WriteFile(malformedPath, malformedContent, 0o644); err != nil {
		t.Fatalf("failed to create malformed file: %v", err)
	}

	task := model.TaskDefinition{
		ID:            "malformed-sarif-task",
		Workdir:       tempDir,
		MetricsFormat: "sarif",
		MetricsPath:   "malformed.sarif",
	}

	m := parseTaskMetrics(task, false)
	if m != nil {
		t.Fatalf("expected nil metrics for malformed SARIF, got %#v", m)
	}
}

func TestParseTaskMetrics_AbsolutePath(t *testing.T) {
	tempDir := t.TempDir()

	// Create artifact with absolute path
	artifactPath := filepath.Join(tempDir, "absolute-artifact.txt")
	if err := os.WriteFile(artifactPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	task := model.TaskDefinition{
		ID:            "absolute-path-task",
		Workdir:       "/some/other/dir", // Different workdir
		MetricsFormat: "artifact",
		MetricsPath:   artifactPath, // Absolute path
	}

	m := parseTaskMetrics(task, false)
	if m == nil {
		t.Fatalf("expected non-nil metrics for absolute path")
	}

	// Verify absolute path is used correctly
	path, ok := m.Data["path"].(string)
	if !ok {
		t.Fatalf("expected path in metrics data")
	}
	if path != artifactPath {
		t.Fatalf("expected absolute path %q, got %q", artifactPath, path)
	}
}

func TestParseTaskMetrics_RelativePath(t *testing.T) {
	tempDir := t.TempDir()

	// Create artifact with relative path
	artifactPath := filepath.Join(tempDir, "relative-artifact.txt")
	if err := os.WriteFile(artifactPath, []byte("test"), 0o644); err != nil {
		t.Fatalf("failed to create artifact: %v", err)
	}

	task := model.TaskDefinition{
		ID:            "relative-path-task",
		Workdir:       tempDir,
		MetricsFormat: "artifact",
		MetricsPath:   "relative-artifact.txt", // Relative path
	}

	m := parseTaskMetrics(task, false)
	if m == nil {
		t.Fatalf("expected non-nil metrics for relative path")
	}

	// Verify relative path is resolved correctly
	path, ok := m.Data["path"].(string)
	if !ok {
		t.Fatalf("expected path in metrics data")
	}
	if path != artifactPath {
		t.Fatalf("expected resolved path %q, got %q", artifactPath, path)
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
