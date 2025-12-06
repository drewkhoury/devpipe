package metrics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseJUnitXML(t *testing.T) {
	tests := []struct {
		name        string
		xmlContent  string
		wantErr     bool
		wantTests   int
		wantPassed  int
		wantFailed  int
	}{
		{
			name: "valid junit xml with passing tests",
			xmlContent: `<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="MyTests" tests="2" failures="0" errors="0" skipped="0">
  <testcase name="test1" classname="MyClass" time="0.001"/>
  <testcase name="test2" classname="MyClass" time="0.002"/>
</testsuite>`,
			wantErr:    false,
			wantTests:  2,
			wantPassed: 2,
			wantFailed: 0,
		},
		{
			name: "valid junit xml with failures",
			xmlContent: `<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="MyTests" tests="3" failures="1" errors="0" skipped="0">
  <testcase name="test1" classname="MyClass" time="0.001"/>
  <testcase name="test2" classname="MyClass" time="0.002">
    <failure message="assertion failed">Expected true but got false</failure>
  </testcase>
  <testcase name="test3" classname="MyClass" time="0.001"/>
</testsuite>`,
			wantErr:    false,
			wantTests:  3,
			wantPassed: 2,
			wantFailed: 1,
		},
		{
			name: "valid junit xml with skipped tests",
			xmlContent: `<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="MyTests" tests="3" failures="0" errors="0" skipped="1">
  <testcase name="test1" classname="MyClass" time="0.001"/>
  <testcase name="test2" classname="MyClass" time="0.000">
    <skipped/>
  </testcase>
  <testcase name="test3" classname="MyClass" time="0.001"/>
</testsuite>`,
			wantErr:    false,
			wantTests:  3,
			wantPassed: 2,
			wantFailed: 0,
		},
		{
			name:       "invalid xml",
			xmlContent: `not valid xml`,
			wantErr:    true,
		},
		{
			name:       "empty file",
			xmlContent: ``,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			xmlPath := filepath.Join(tmpDir, "junit.xml")
			
			if err := os.WriteFile(xmlPath, []byte(tt.xmlContent), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Parse JUnit XML
			metrics, err := ParseJUnitXML(xmlPath)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJUnitXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify metrics
			if metrics == nil {
				t.Fatal("Expected non-nil metrics")
			}

			// Verify summary format
			if metrics.SummaryFormat != "junit" {
				t.Errorf("SummaryFormat = %s, want junit", metrics.SummaryFormat)
			}

			// Verify data map
			if metrics.Data == nil {
				t.Fatal("Expected non-nil Data map")
			}

			tests, ok := metrics.Data["tests"].(int)
			if !ok {
				t.Fatal("Expected tests to be int")
			}
			if tests != tt.wantTests {
				t.Errorf("tests = %d, want %d", tests, tt.wantTests)
			}

			failures, ok := metrics.Data["failures"].(int)
			if !ok {
				t.Fatal("Expected failures to be int")
			}
			
			errors, ok := metrics.Data["errors"].(int)
			if !ok {
				t.Fatal("Expected errors to be int")
			}
			
			skipped, ok := metrics.Data["skipped"].(int)
			if !ok {
				t.Fatal("Expected skipped to be int")
			}
			
			// Passed = total - failures - errors - skipped
			passed := tests - failures - errors - skipped
			if passed != tt.wantPassed {
				t.Errorf("passed = %d, want %d", passed, tt.wantPassed)
			}

			if failures != tt.wantFailed {
				t.Errorf("failures = %d, want %d", failures, tt.wantFailed)
			}
		})
	}
}

func TestParseJUnitXML_NonExistentFile(t *testing.T) {
	_, err := ParseJUnitXML("/nonexistent/file.xml")
	
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseJUnitXML_MultipleTestSuites(t *testing.T) {
	xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="Suite1" tests="2" failures="0" errors="0" skipped="0">
    <testcase name="test1" classname="Class1" time="0.001"/>
    <testcase name="test2" classname="Class1" time="0.002"/>
  </testsuite>
  <testsuite name="Suite2" tests="1" failures="0" errors="0" skipped="0">
    <testcase name="test3" classname="Class2" time="0.001"/>
  </testsuite>
</testsuites>`

	tmpDir := t.TempDir()
	xmlPath := filepath.Join(tmpDir, "junit.xml")
	
	if err := os.WriteFile(xmlPath, []byte(xmlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	metrics, err := ParseJUnitXML(xmlPath)
	
	if err != nil {
		t.Fatalf("ParseJUnitXML() error = %v", err)
	}

	// Should aggregate across all suites
	tests, ok := metrics.Data["tests"].(int)
	if !ok {
		t.Fatal("Expected tests to be int")
	}
	if tests != 3 {
		t.Errorf("tests = %d, want 3", tests)
	}

	failures, ok := metrics.Data["failures"].(int)
	if !ok {
		t.Fatal("Expected failures to be int")
	}
	
	errors, ok := metrics.Data["errors"].(int)
	if !ok {
		t.Fatal("Expected errors to be int")
	}
	
	skipped, ok := metrics.Data["skipped"].(int)
	if !ok {
		t.Fatal("Expected skipped to be int")
	}
	
	passed := tests - failures - errors - skipped
	if passed != 3 {
		t.Errorf("passed = %d, want 3", passed)
	}
}
