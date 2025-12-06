package sarif

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrintFindings(t *testing.T) {
	findings := []Finding{
		{
			RuleID:  "RULE1",
			File:    "test.go",
			Line:    10,
			Column:  5,
			Message: "Test finding",
			Level:   "error",
		},
		{
			RuleID:  "RULE2",
			File:    "app.go",
			Line:    20,
			Message: "Another finding",
			Level:   "warning",
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintFindings(findings, false)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should contain finding information
	if !strings.Contains(output, "RULE1") {
		t.Error("Expected output to contain RULE1")
	}

	if !strings.Contains(output, "test.go") {
		t.Error("Expected output to contain test.go")
	}

	if !strings.Contains(output, "Test finding") {
		t.Error("Expected output to contain 'Test finding'")
	}
}

func TestPrintFindings_Verbose(t *testing.T) {
	findings := []Finding{
		{
			RuleID:           "SQL_INJECTION",
			RuleName:         "SQL Injection",
			File:             "app.go",
			Line:             50,
			Message:          "SQL injection vulnerability",
			Level:            "error",
			ShortDesc:        "Potential SQL injection",
			FullDesc:         "User input flows to SQL query",
			HelpText:         "Use parameterized queries",
			SecuritySeverity: "8.5",
			Tags:             []string{"security", "sql"},
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintFindings(findings, true)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Verbose mode should include more details
	if !strings.Contains(output, "SQL_INJECTION") {
		t.Error("Expected output to contain SQL_INJECTION")
	}

	if !strings.Contains(output, "SQL injection vulnerability") {
		t.Error("Expected output to contain message")
	}
}

func TestPrintSummary(t *testing.T) {
	findings := []Finding{
		{RuleID: "RULE1", Level: "error"},
		{RuleID: "RULE1", Level: "error"},
		{RuleID: "RULE2", Level: "warning"},
		{RuleID: "RULE3", Level: "note"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintSummary(findings)

	_ = w.Close() // Test cleanup
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r) // Test output capture
	output := buf.String()

	// Should contain summary information - just verify non-empty output
	if len(output) == 0 {
		t.Error("Expected non-empty summary output")
	}

	// Should contain some count information
	t.Logf("Summary output: %s", output)
}

func TestFindSARIFFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some SARIF files
	file1 := filepath.Join(tmpDir, "results.sarif")
	file2 := filepath.Join(tmpDir, "subdir", "more.sarif")
	file3 := filepath.Join(tmpDir, "not-sarif.txt")

	_ = os.WriteFile(file1, []byte("{}"), 0644)        // Test setup
	_ = os.MkdirAll(filepath.Dir(file2), 0755)         // Test setup
	_ = os.WriteFile(file2, []byte("{}"), 0644)        // Test setup
	_ = os.WriteFile(file3, []byte("not sarif"), 0644) // Test setup

	files, err := FindSARIFFiles(tmpDir)
	if err != nil {
		t.Fatalf("FindSARIFFiles() error = %v", err)
	}

	// Should find 2 SARIF files
	if len(files) != 2 {
		t.Errorf("Expected 2 SARIF files, got %d", len(files))
	}

	// Verify files are .sarif files
	for _, f := range files {
		if !strings.HasSuffix(f, ".sarif") {
			t.Errorf("File %s does not have .sarif extension", f)
		}
	}
}

func TestFindSARIFFiles_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	files, err := FindSARIFFiles(tmpDir)
	if err != nil {
		t.Fatalf("FindSARIFFiles() error = %v", err)
	}

	// Should return empty list
	if len(files) != 0 {
		t.Errorf("Expected 0 SARIF files, got %d", len(files))
	}
}

func TestFindSARIFFiles_NonexistentDir(t *testing.T) {
	_, err := FindSARIFFiles("/nonexistent/directory")
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}
}
