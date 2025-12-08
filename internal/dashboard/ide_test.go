package dashboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteIDEViewer(t *testing.T) {
	tmpDir := t.TempDir()
	runDir := filepath.Join(tmpDir, "run-123")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatalf("Failed to create run dir: %v", err)
	}

	// Create some test files
	if err := os.WriteFile(filepath.Join(runDir, "pipeline.log"), []byte("test log"), 0644); err != nil {
		t.Fatalf("Failed to write pipeline.log: %v", err)
	}

	idePath := filepath.Join(tmpDir, "ide.html")
	err := writeIDEViewer(idePath, "run-123", runDir)
	if err != nil {
		t.Fatalf("writeIDEViewer() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(idePath); os.IsNotExist(err) {
		t.Error("Expected ide.html to be created")
	}

	// Read and verify content
	content, err := os.ReadFile(idePath)
	if err != nil {
		t.Fatalf("Failed to read ide.html: %v", err)
	}

	contentStr := string(content)

	// Should contain HTML structure
	if !strings.Contains(contentStr, "<html") {
		t.Error("Expected HTML to contain <html tag")
	}

	// Should contain run ID
	if !strings.Contains(contentStr, "run-123") {
		t.Error("Expected HTML to contain run ID")
	}

	// Should contain Monaco editor reference
	if !strings.Contains(contentStr, "monaco-editor") {
		t.Error("Expected HTML to contain Monaco editor reference")
	}
}

func TestCollectFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	if err := os.WriteFile(filepath.Join(tmpDir, "pipeline.log"), []byte("pipeline log content"), 0644); err != nil {
		t.Fatalf("Failed to write pipeline.log: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "config.toml"), []byte("[test]\nkey = \"value\""), 0644); err != nil {
		t.Fatalf("Failed to write config.toml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "run.json"), []byte(`{"runId":"test"}`), 0644); err != nil {
		t.Fatalf("Failed to write run.json: %v", err)
	}

	// Create logs directory
	logsDir := filepath.Join(tmpDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		t.Fatalf("Failed to create logs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(logsDir, "task1.log"), []byte("task log"), 0644); err != nil {
		t.Fatalf("Failed to write task log: %v", err)
	}

	// Create outputs directory
	outputsDir := filepath.Join(tmpDir, "outputs")
	if err := os.MkdirAll(outputsDir, 0755); err != nil {
		t.Fatalf("Failed to create outputs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outputsDir, "report.xml"), []byte("<report/>"), 0644); err != nil {
		t.Fatalf("Failed to write artifact: %v", err)
	}

	files := collectFiles(tmpDir)

	// Should have collected all files
	if len(files) < 4 {
		t.Errorf("Expected at least 4 files, got %d", len(files))
	}

	// Verify specific files were collected
	fileMap := make(map[string]FileInfo)
	for _, f := range files {
		fileMap[f.Path] = f
	}

	if _, ok := fileMap["pipeline.log"]; !ok {
		t.Error("Expected pipeline.log to be collected")
	}

	if _, ok := fileMap["config.toml"]; !ok {
		t.Error("Expected config.toml to be collected")
	}

	if _, ok := fileMap["run.json"]; !ok {
		t.Error("Expected run.json to be collected")
	}

	if _, ok := fileMap["logs/task1.log"]; !ok {
		t.Error("Expected logs/task1.log to be collected")
	}

	if _, ok := fileMap["outputs/report.xml"]; !ok {
		t.Error("Expected outputs/report.xml to be collected")
	}
}

func TestCollectFilesWithNestedArtifacts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested outputs directory
	outputsDir := filepath.Join(tmpDir, "outputs", "subdir")
	if err := os.MkdirAll(outputsDir, 0755); err != nil {
		t.Fatalf("Failed to create nested outputs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outputsDir, "nested.txt"), []byte("nested content"), 0644); err != nil {
		t.Fatalf("Failed to write nested artifact: %v", err)
	}

	files := collectFiles(tmpDir)

	// Find the nested artifact
	found := false
	for _, f := range files {
		if strings.Contains(f.Path, "outputs/subdir/nested.txt") {
			found = true
			if f.Content != "nested content" {
				t.Errorf("Expected content 'nested content', got %q", f.Content)
			}
		}
	}

	if !found {
		t.Error("Expected nested artifact to be collected")
	}
}

func TestCollectFilesEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	files := collectFiles(tmpDir)

	// collectFiles may return nil for empty directories, which is acceptable
	// Just verify we don't have any files
	if len(files) != 0 {
		t.Errorf("Expected 0 files, got %d", len(files))
	}
}

func TestCollectFilesWithANSICodes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create log file with ANSI codes
	ansiContent := "\x1b[31mRed text\x1b[0m normal text"
	if err := os.WriteFile(filepath.Join(tmpDir, "pipeline.log"), []byte(ansiContent), 0644); err != nil {
		t.Fatalf("Failed to write pipeline.log: %v", err)
	}

	files := collectFiles(tmpDir)

	if len(files) == 0 {
		t.Fatal("Expected at least one file")
	}

	// ANSI codes should be stripped
	if strings.Contains(files[0].Content, "\x1b") {
		t.Error("Expected ANSI codes to be stripped from content")
	}

	if !strings.Contains(files[0].Content, "Red text") {
		t.Error("Expected text content to be preserved")
	}
}

func TestWriteIDEViewerWithInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	runDir := tmpDir
	idePath := filepath.Join(tmpDir, "ide.html")

	// This should not fail even with no files
	err := writeIDEViewer(idePath, "test-run", runDir)
	if err != nil {
		t.Fatalf("writeIDEViewer() should handle empty file list, error = %v", err)
	}

	// Verify the HTML contains valid JSON
	content, err := os.ReadFile(idePath)
	if err != nil {
		t.Fatalf("Failed to read ide.html: %v", err)
	}

	// Extract the JSON from the HTML
	contentStr := string(content)
	if !strings.Contains(contentStr, "let files = ") {
		t.Error("Expected HTML to contain files variable")
	}

	// Find the JSON array
	start := strings.Index(contentStr, "let files = ")
	if start == -1 {
		t.Fatal("Could not find files variable")
	}
	start += len("let files = ")
	end := strings.Index(contentStr[start:], ";")
	if end == -1 {
		t.Fatal("Could not find end of files variable")
	}
	jsonStr := contentStr[start : start+end]

	// Verify it's valid JSON
	var files []FileInfo
	if err := json.Unmarshal([]byte(jsonStr), &files); err != nil {
		t.Errorf("Expected valid JSON, got error: %v", err)
	}
}

func TestFileInfoStructure(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	testContent := "test content"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a minimal directory structure
	if err := os.WriteFile(filepath.Join(tmpDir, "pipeline.log"), []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to write pipeline.log: %v", err)
	}

	files := collectFiles(tmpDir)

	if len(files) == 0 {
		t.Fatal("Expected at least one file")
	}

	// Verify FileInfo structure
	file := files[0]
	if file.Name == "" {
		t.Error("Expected Name to be set")
	}
	if file.Path == "" {
		t.Error("Expected Path to be set")
	}
	if file.Size < 0 {
		t.Error("Expected Size to be non-negative")
	}
	if file.Content == "" {
		t.Error("Expected Content to be set")
	}
}

func TestWriteIDEViewerError(t *testing.T) {
	// Try to write to an invalid path
	invalidPath := "/invalid/path/that/does/not/exist/ide.html"
	runDir := "/some/run/dir"

	err := writeIDEViewer(invalidPath, "test-run", runDir)
	if err == nil {
		t.Error("Expected error when writing to invalid path")
	}
}

func TestCollectFilesAllFileTypes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create all possible file types
	files := map[string]string{
		"pipeline.log":              "pipeline content",
		"config.toml":               "[test]\nkey = \"value\"",
		"run.json":                  `{"runId":"test"}`,
		"logs/task1.log":            "task1 output",
		"logs/task2.log":            "task2 output",
		"outputs/report.xml":        "<report/>",
		"outputs/coverage/data.txt": "coverage data",
	}

	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create dir %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", path, err)
		}
	}

	collected := collectFiles(tmpDir)

	// Should have all 7 files
	if len(collected) != 7 {
		t.Errorf("Expected 7 files, got %d", len(collected))
	}

	// Verify all expected files are present
	fileMap := make(map[string]bool)
	for _, f := range collected {
		fileMap[f.Path] = true
	}

	expectedPaths := []string{
		"pipeline.log",
		"config.toml",
		"run.json",
		"logs/task1.log",
		"logs/task2.log",
		"outputs/report.xml",
		"outputs/coverage/data.txt",
	}

	for _, expected := range expectedPaths {
		if !fileMap[expected] {
			t.Errorf("Expected to find %s in collected files", expected)
		}
	}
}
