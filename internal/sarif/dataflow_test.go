package sarif

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFindings_WithDataFlow(t *testing.T) {
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
              "id": "SQL_INJECTION",
              "name": "SQL Injection",
              "shortDescription": {
                "text": "Potential SQL injection"
              },
              "fullDescription": {
                "text": "User input flows to SQL query without sanitization"
              },
              "help": {
                "text": "Use parameterized queries"
              },
              "properties": {
                "tags": ["security", "sql"],
                "precision": "high",
                "security-severity": "8.5"
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "SQL_INJECTION",
          "ruleIndex": 0,
          "level": "error",
          "message": {
            "text": "SQL injection vulnerability"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "app.go"
                },
                "region": {
                  "startLine": 50,
                  "startColumn": 10,
                  "endLine": 50,
                  "endColumn": 30
                }
              }
            }
          ],
          "codeFlows": [
            {
              "threadFlows": [
                {
                  "locations": [
                    {
                      "location": {
                        "physicalLocation": {
                          "artifactLocation": {
                            "uri": "app.go"
                          },
                          "region": {
                            "startLine": 10,
                            "startColumn": 5
                          }
                        },
                        "message": {
                          "text": "User input received"
                        }
                      }
                    },
                    {
                      "location": {
                        "physicalLocation": {
                          "artifactLocation": {
                            "uri": "app.go"
                          },
                          "region": {
                            "startLine": 30,
                            "startColumn": 8
                          }
                        },
                        "message": {
                          "text": "Input passed to query"
                        }
                      }
                    }
                  ]
                }
              ]
            }
          ],
          "relatedLocations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "app.go"
                },
                "region": {
                  "startLine": 10,
                  "startColumn": 5
                }
              },
              "message": {
                "text": "Source of tainted data"
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

	// Check basic fields
	if finding.RuleID != "SQL_INJECTION" {
		t.Errorf("RuleID = %s, want SQL_INJECTION", finding.RuleID)
	}

	if finding.RuleName != "SQL Injection" {
		t.Errorf("RuleName = %s, want SQL Injection", finding.RuleName)
	}

	if finding.Level != "error" {
		t.Errorf("Level = %s, want error", finding.Level)
	}

	// Check rule metadata
	if finding.ShortDesc != "Potential SQL injection" {
		t.Errorf("ShortDesc = %s, want 'Potential SQL injection'", finding.ShortDesc)
	}

	if finding.FullDesc != "User input flows to SQL query without sanitization" {
		t.Errorf("FullDesc incorrect: %s", finding.FullDesc)
	}

	if finding.HelpText != "Use parameterized queries" {
		t.Errorf("HelpText = %s, want 'Use parameterized queries'", finding.HelpText)
	}

	// Check tags
	if len(finding.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(finding.Tags))
	}

	// Check precision and severity
	if finding.Precision != "high" {
		t.Errorf("Precision = %s, want high", finding.Precision)
	}

	if finding.SecuritySeverity != "8.5" {
		t.Errorf("SecuritySeverity = %s, want 8.5", finding.SecuritySeverity)
	}

	// Check location details
	if finding.Line != 50 {
		t.Errorf("Line = %d, want 50", finding.Line)
	}

	if finding.Column != 10 {
		t.Errorf("Column = %d, want 10", finding.Column)
	}

	if finding.EndLine != 50 {
		t.Errorf("EndLine = %d, want 50", finding.EndLine)
	}

	if finding.EndColumn != 30 {
		t.Errorf("EndColumn = %d, want 30", finding.EndColumn)
	}

	// Check data flow steps
	if len(finding.DataFlowSteps) != 2 {
		t.Fatalf("Expected 2 data flow steps, got %d", len(finding.DataFlowSteps))
	}

	if finding.DataFlowSteps[0].Line != 10 {
		t.Errorf("First data flow step line = %d, want 10", finding.DataFlowSteps[0].Line)
	}

	if finding.DataFlowSteps[0].Message != "User input received" {
		t.Errorf("First data flow step message = %s", finding.DataFlowSteps[0].Message)
	}

	if finding.DataFlowSteps[1].Line != 30 {
		t.Errorf("Second data flow step line = %d, want 30", finding.DataFlowSteps[1].Line)
	}

	// Check source location
	if finding.SourceLocation == "" {
		t.Error("Expected non-empty source location")
	}
}

func TestGetFindings_DefaultLevel(t *testing.T) {
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
          "message": {
            "text": "Test finding without level"
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

	// Should default to "warning" when level is not specified
	if findings[0].Level != "warning" {
		t.Errorf("Level = %s, want warning (default)", findings[0].Level)
	}
}

func TestGetFindings_Sorting(t *testing.T) {
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
          "ruleId": "RULE1",
          "message": {"text": "Finding in file2"},
          "locations": [{
            "physicalLocation": {
              "artifactLocation": {"uri": "file2.go"},
              "region": {"startLine": 10}
            }
          }]
        },
        {
          "ruleId": "RULE2",
          "message": {"text": "Finding in file1"},
          "locations": [{
            "physicalLocation": {
              "artifactLocation": {"uri": "file1.go"},
              "region": {"startLine": 20}
            }
          }]
        },
        {
          "ruleId": "RULE3",
          "message": {"text": "Another finding in file1"},
          "locations": [{
            "physicalLocation": {
              "artifactLocation": {"uri": "file1.go"},
              "region": {"startLine": 5}
            }
          }]
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

	if len(findings) != 3 {
		t.Fatalf("Expected 3 findings, got %d", len(findings))
	}

	// Should be sorted by file, then line
	if findings[0].File != "file1.go" || findings[0].Line != 5 {
		t.Errorf("First finding should be file1.go:5, got %s:%d", findings[0].File, findings[0].Line)
	}

	if findings[1].File != "file1.go" || findings[1].Line != 20 {
		t.Errorf("Second finding should be file1.go:20, got %s:%d", findings[1].File, findings[1].Line)
	}

	if findings[2].File != "file2.go" || findings[2].Line != 10 {
		t.Errorf("Third finding should be file2.go:10, got %s:%d", findings[2].File, findings[2].Line)
	}
}
