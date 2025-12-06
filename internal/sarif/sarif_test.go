package sarif

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		sarifContent string
		wantErr      bool
		wantRuns     int
		wantResults  int
	}{
		{
			name: "valid sarif with one result",
			sarifContent: `{
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "TestTool",
          "version": "1.0.0"
        }
      },
      "results": [
        {
          "ruleId": "TEST001",
          "level": "warning",
          "message": {
            "text": "Test warning"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.go"
                },
                "region": {
                  "startLine": 10
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`,
			wantErr:     false,
			wantRuns:    1,
			wantResults: 1,
		},
		{
			name: "valid sarif with no results",
			sarifContent: `{
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "TestTool",
          "version": "1.0.0"
        }
      },
      "results": []
    }
  ]
}`,
			wantErr:     false,
			wantRuns:    1,
			wantResults: 0,
		},
		{
			name:         "invalid json",
			sarifContent: `not valid json`,
			wantErr:      true,
		},
		{
			name:         "empty file",
			sarifContent: ``,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			sarifPath := filepath.Join(tmpDir, "results.sarif")
			
			if err := os.WriteFile(sarifPath, []byte(tt.sarifContent), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Parse SARIF
			doc, err := Parse(sarifPath)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify structure
			if doc == nil {
				t.Fatal("Expected non-nil document")
			}

			if len(doc.Runs) != tt.wantRuns {
				t.Errorf("Runs count = %d, want %d", len(doc.Runs), tt.wantRuns)
			}

			if tt.wantRuns > 0 && len(doc.Runs[0].Results) != tt.wantResults {
				t.Errorf("Results count = %d, want %d", len(doc.Runs[0].Results), tt.wantResults)
			}
		})
	}
}

func TestParse_NonExistentFile(t *testing.T) {
	_, err := Parse("/nonexistent/file.sarif")
	
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestGetFindings(t *testing.T) {
	sarifContent := `{
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "TestTool",
          "version": "1.0.0",
          "rules": [
            {
              "id": "TEST001",
              "name": "TestRule",
              "shortDescription": {
                "text": "Test rule description"
              },
              "properties": {
                "security-severity": "7.5"
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "TEST001",
          "ruleIndex": 0,
          "level": "warning",
          "message": {
            "text": "Test warning message"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.go"
                },
                "region": {
                  "startLine": 10,
                  "startColumn": 5
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`

	tmpDir := t.TempDir()
	sarifPath := filepath.Join(tmpDir, "results.sarif")
	
	if err := os.WriteFile(sarifPath, []byte(sarifContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	doc, err := Parse(sarifPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	findings := doc.GetFindings()

	if len(findings) != 1 {
		t.Fatalf("Expected 1 finding, got %d", len(findings))
	}

	finding := findings[0]

	if finding.RuleID != "TEST001" {
		t.Errorf("RuleID = %s, want TEST001", finding.RuleID)
	}

	if finding.Level != "warning" {
		t.Errorf("Level = %s, want warning", finding.Level)
	}

	if finding.Message != "Test warning message" {
		t.Errorf("Message = %s, want 'Test warning message'", finding.Message)
	}

	if finding.File != "test.go" {
		t.Errorf("File = %s, want test.go", finding.File)
	}

	if finding.Line != 10 {
		t.Errorf("Line = %d, want 10", finding.Line)
	}

	if finding.SecuritySeverity != "7.5" {
		t.Errorf("SecuritySeverity = %s, want 7.5", finding.SecuritySeverity)
	}
}

func TestFinding_Struct(t *testing.T) {
	// Test Finding struct creation
	finding := Finding{
		RuleID:           "TEST001",
		RuleName:         "TestRule",
		Message:          "Test message",
		File:             "test.go",
		Line:             10,
		Column:           5,
		Level:            "warning",
		SecuritySeverity: "7.5",
	}

	if finding.RuleID != "TEST001" {
		t.Errorf("RuleID = %s, want TEST001", finding.RuleID)
	}

	if finding.Line != 10 {
		t.Errorf("Line = %d, want 10", finding.Line)
	}
}

func TestGetFindings_MultipleRuns(t *testing.T) {
	sarifContent := `{
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Tool1",
          "version": "1.0.0"
        }
      },
      "results": [
        {
          "ruleId": "RULE1",
          "level": "error",
          "message": {
            "text": "Error 1"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "file1.go"
                },
                "region": {
                  "startLine": 5
                }
              }
            }
          ]
        }
      ]
    },
    {
      "tool": {
        "driver": {
          "name": "Tool2",
          "version": "2.0.0"
        }
      },
      "results": [
        {
          "ruleId": "RULE2",
          "level": "warning",
          "message": {
            "text": "Warning 1"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "file2.go"
                },
                "region": {
                  "startLine": 10
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`

	tmpDir := t.TempDir()
	sarifPath := filepath.Join(tmpDir, "results.sarif")
	
	if err := os.WriteFile(sarifPath, []byte(sarifContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	doc, err := Parse(sarifPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	findings := doc.GetFindings()

	// Should aggregate findings from both runs
	if len(findings) != 2 {
		t.Fatalf("Expected 2 findings, got %d", len(findings))
	}

	// Check first finding
	if findings[0].RuleID != "RULE1" {
		t.Errorf("First finding RuleID = %s, want RULE1", findings[0].RuleID)
	}

	if findings[0].Level != "error" {
		t.Errorf("First finding Level = %s, want error", findings[0].Level)
	}

	// Check second finding
	if findings[1].RuleID != "RULE2" {
		t.Errorf("Second finding RuleID = %s, want RULE2", findings[1].RuleID)
	}

	if findings[1].Level != "warning" {
		t.Errorf("Second finding Level = %s, want warning", findings[1].Level)
	}
}

func TestGetFindings_NoLocations(t *testing.T) {
	sarifContent := `{
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "TestTool",
          "version": "1.0.0"
        }
      },
      "results": [
        {
          "ruleId": "TEST001",
          "level": "note",
          "message": {
            "text": "Test note without location"
          },
          "locations": []
        }
      ]
    }
  ]
}`

	tmpDir := t.TempDir()
	sarifPath := filepath.Join(tmpDir, "results.sarif")
	
	if err := os.WriteFile(sarifPath, []byte(sarifContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	doc, err := Parse(sarifPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	findings := doc.GetFindings()

	// Findings without locations may be filtered out by GetFindings
	// This is expected behavior - just verify no panic
	if len(findings) == 0 {
		// This is acceptable - findings without locations are skipped
		return
	}

	// If a finding is returned, it should have empty location info
	if findings[0].File != "" {
		t.Errorf("Expected empty file for finding without location, got %s", findings[0].File)
	}

	if findings[0].Line != 0 {
		t.Errorf("Expected line 0 for finding without location, got %d", findings[0].Line)
	}
}
