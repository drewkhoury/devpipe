package metrics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSARIF(t *testing.T) {
	tests := []struct {
		name         string
		sarifContent string
		wantErr      bool
		wantFindings int
		wantErrors   int
		wantWarnings int
	}{
		{
			name: "valid sarif with warnings",
			sarifContent: `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
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
			wantErr:      false,
			wantFindings: 1,
			wantErrors:   0,
			wantWarnings: 1,
		},
		{
			name: "valid sarif with errors",
			sarifContent: `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
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
          "level": "error",
          "message": {
            "text": "Test error"
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
        },
        {
          "ruleId": "TEST002",
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
                  "startLine": 20
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`,
			wantErr:      false,
			wantFindings: 2,
			wantErrors:   1,
			wantWarnings: 1,
		},
		{
			name: "empty sarif",
			sarifContent: `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": []
}`,
			wantErr:      false,
			wantFindings: 0,
			wantErrors:   0,
			wantWarnings: 0,
		},
		{
			name:         "invalid json",
			sarifContent: `not valid json`,
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
			metrics, err := ParseSARIF(sarifPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSARIF() error = %v, wantErr %v", err, tt.wantErr)
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
			if metrics.SummaryFormat != "sarif" {
				t.Errorf("SummaryFormat = %s, want sarif", metrics.SummaryFormat)
			}

			// Verify data map
			if metrics.Data == nil {
				t.Fatal("Expected non-nil Data map")
			}

			total, ok := metrics.Data["total"].(int)
			if !ok {
				t.Fatal("Expected total to be int")
			}
			if total != tt.wantFindings {
				t.Errorf("total = %d, want %d", total, tt.wantFindings)
			}

			errors, ok := metrics.Data["errors"].(int)
			if !ok {
				t.Fatal("Expected errors to be int")
			}
			if errors != tt.wantErrors {
				t.Errorf("errors = %d, want %d", errors, tt.wantErrors)
			}

			warnings, ok := metrics.Data["warnings"].(int)
			if !ok {
				t.Fatal("Expected warnings to be int")
			}
			if warnings != tt.wantWarnings {
				t.Errorf("warnings = %d, want %d", warnings, tt.wantWarnings)
			}
		})
	}
}

func TestParseSARIF_NonExistentFile(t *testing.T) {
	_, err := ParseSARIF("/nonexistent/file.sarif")

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseSARIF_WithNotes(t *testing.T) {
	sarifContent := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
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
            "text": "Informational note"
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

	metrics, err := ParseSARIF(sarifPath)

	if err != nil {
		t.Fatalf("ParseSARIF() error = %v", err)
	}

	// Verify notes count
	notes, ok := metrics.Data["notes"].(int)
	if !ok {
		t.Fatal("Expected notes to be int")
	}
	if notes != 1 {
		t.Errorf("notes = %d, want 1", notes)
	}

	// Verify Kind field
	if metrics.Kind != "security" {
		t.Errorf("Expected Kind 'security', got %s", metrics.Kind)
	}
}

func TestParseSARIF_UnknownLevel(t *testing.T) {
	sarifContent := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
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
          "level": "unknown",
          "message": {
            "text": "Unknown level finding"
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

	metrics, err := ParseSARIF(sarifPath)

	if err != nil {
		t.Fatalf("ParseSARIF() error = %v", err)
	}

	// Unknown level should default to warning
	warnings, ok := metrics.Data["warnings"].(int)
	if !ok {
		t.Fatal("Expected warnings to be int")
	}
	if warnings != 1 {
		t.Errorf("warnings = %d, want 1 (unknown level should default to warning)", warnings)
	}
}

func TestParseSARIF_WithOptionalFields(t *testing.T) {
	sarifContent := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "CodeQL",
          "version": "2.0.0",
          "rules": [
            {
              "id": "SQL001",
              "name": "SQL Injection",
              "shortDescription": {
                "text": "Potential SQL injection"
              },
              "properties": {
                "tags": ["security", "external/cwe/cwe-89"],
                "precision": "high",
                "security-severity": "8.5"
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "SQL001",
          "ruleIndex": 0,
          "level": "error",
          "message": {
            "text": "Unsanitized user input flows to SQL query"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "database.go",
                  "uriBaseId": "%SRCROOT%"
                },
                "region": {
                  "startLine": 42,
                  "startColumn": 15
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
                            "uri": "handler.go"
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
                            "uri": "database.go"
                          },
                          "region": {
                            "startLine": 42,
                            "startColumn": 15
                          }
                        },
                        "message": {
                          "text": "Flows to SQL query"
                        }
                      }
                    }
                  ]
                }
              ]
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

	metrics, err := ParseSARIF(sarifPath)

	if err != nil {
		t.Fatalf("ParseSARIF() error = %v", err)
	}

	// Verify findings
	findings, ok := metrics.Data["findings"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected findings to be []map[string]interface{}")
	}

	if len(findings) != 1 {
		t.Fatalf("Expected 1 finding, got %d", len(findings))
	}

	finding := findings[0]

	// Verify basic fields
	if finding["ruleId"] != "SQL001" {
		t.Errorf("Expected ruleId 'SQL001', got %v", finding["ruleId"])
	}
	if finding["file"] != "database.go" {
		t.Errorf("Expected file 'database.go', got %v", finding["file"])
	}
	if finding["line"] != 42 {
		t.Errorf("Expected line 42, got %v", finding["line"])
	}
	if finding["column"] != 15 {
		t.Errorf("Expected column 15, got %v", finding["column"])
	}

	// Verify optional fields are present
	if _, hasShortDesc := finding["shortDesc"]; !hasShortDesc {
		t.Error("Expected shortDesc field")
	}
	if _, hasSeverity := finding["severity"]; !hasSeverity {
		t.Error("Expected severity field")
	}
	if _, hasTags := finding["tags"]; !hasTags {
		t.Error("Expected tags field")
	}
	if _, hasPrecision := finding["precision"]; !hasPrecision {
		t.Error("Expected precision field")
	}
	// Note: source field only present when SARIF has relatedLocations
	// This test doesn't include relatedLocations, so we don't check for it
	if _, hasDataFlow := finding["dataFlow"]; !hasDataFlow {
		t.Error("Expected dataFlow field")
	}

	// Verify data flow steps
	dataFlow, ok := finding["dataFlow"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected dataFlow to be []map[string]interface{}")
	}
	if len(dataFlow) != 2 {
		t.Errorf("Expected 2 data flow steps, got %d", len(dataFlow))
	}

	// Verify first data flow step
	if dataFlow[0]["file"] != "handler.go" {
		t.Errorf("Expected first step file 'handler.go', got %v", dataFlow[0]["file"])
	}
	if dataFlow[0]["line"] != 10 {
		t.Errorf("Expected first step line 10, got %v", dataFlow[0]["line"])
	}
	if dataFlow[0]["message"] != "User input received" {
		t.Errorf("Expected first step message 'User input received', got %v", dataFlow[0]["message"])
	}

	// Verify rules summary
	rules, ok := metrics.Data["rules"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected rules to be []map[string]interface{}")
	}
	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	rule := rules[0]
	if rule["id"] != "SQL001" {
		t.Errorf("Expected rule id 'SQL001', got %v", rule["id"])
	}
	if rule["count"] != 1 {
		t.Errorf("Expected rule count 1, got %v", rule["count"])
	}
	if _, hasSeverity := rule["severity"]; !hasSeverity {
		t.Error("Expected severity in rule summary")
	}
}

func TestParseSARIF_MultipleRules(t *testing.T) {
	sarifContent := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "TestTool",
          "version": "1.0.0",
          "rules": [
            {
              "id": "RULE001",
              "properties": {
                "security-severity": "7.5"
              }
            },
            {
              "id": "RULE002",
              "properties": {
                "security-severity": "5.0"
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "RULE001",
          "ruleIndex": 0,
          "level": "warning",
          "message": {
            "text": "First finding"
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
        },
        {
          "ruleId": "RULE001",
          "ruleIndex": 0,
          "level": "warning",
          "message": {
            "text": "Second finding"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.go"
                },
                "region": {
                  "startLine": 20
                }
              }
            }
          ]
        },
        {
          "ruleId": "RULE002",
          "ruleIndex": 1,
          "level": "warning",
          "message": {
            "text": "Third finding"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.go"
                },
                "region": {
                  "startLine": 30
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

	metrics, err := ParseSARIF(sarifPath)

	if err != nil {
		t.Fatalf("ParseSARIF() error = %v", err)
	}

	// Verify total findings
	total, ok := metrics.Data["total"].(int)
	if !ok {
		t.Fatal("Expected total to be int")
	}
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}

	// Verify rules summary
	rules, ok := metrics.Data["rules"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected rules to be []map[string]interface{}")
	}
	if len(rules) != 2 {
		t.Fatalf("Expected 2 rules, got %d", len(rules))
	}

	// Verify rule counts
	ruleCounts := make(map[string]int)
	for _, rule := range rules {
		ruleID, ok := rule["id"].(string)
		if !ok {
			t.Fatal("Expected rule id to be string")
		}
		count, ok := rule["count"].(int)
		if !ok {
			t.Fatal("Expected rule count to be int")
		}
		ruleCounts[ruleID] = count
	}

	if ruleCounts["RULE001"] != 2 {
		t.Errorf("Expected RULE001 count 2, got %d", ruleCounts["RULE001"])
	}
	if ruleCounts["RULE002"] != 1 {
		t.Errorf("Expected RULE002 count 1, got %d", ruleCounts["RULE002"])
	}
}

func TestParseSARIF_WithRelatedLocations(t *testing.T) {
	sarifContent := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
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
            "text": "Tainted data flows to sink"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "sink.go"
                },
                "region": {
                  "startLine": 50,
                  "startColumn": 10
                }
              }
            }
          ],
          "relatedLocations": [
            {
              "id": 1,
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "source.go"
                },
                "region": {
                  "startLine": 10,
                  "startColumn": 5
                }
              },
              "message": {
                "text": "Tainted data originates here"
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

	metrics, err := ParseSARIF(sarifPath)

	if err != nil {
		t.Fatalf("ParseSARIF() error = %v", err)
	}

	// Verify findings
	findings, ok := metrics.Data["findings"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected findings to be []map[string]interface{}")
	}

	if len(findings) != 1 {
		t.Fatalf("Expected 1 finding, got %d", len(findings))
	}

	finding := findings[0]

	// Verify source field is present when relatedLocations exist
	source, hasSource := finding["source"]
	if !hasSource {
		t.Error("Expected source field when relatedLocations are present")
	}

	// Verify source contains the expected information
	sourceStr, ok := source.(string)
	if !ok {
		t.Fatal("Expected source to be string")
	}
	if !strings.Contains(sourceStr, "source.go") {
		t.Errorf("Expected source to contain 'source.go', got %s", sourceStr)
	}
	if !strings.Contains(sourceStr, "10:5") {
		t.Errorf("Expected source to contain '10:5', got %s", sourceStr)
	}
	if !strings.Contains(sourceStr, "Tainted data originates here") {
		t.Errorf("Expected source to contain message, got %s", sourceStr)
	}
}

func TestParseSARIF_EmptySeverity(t *testing.T) {
	sarifContent := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "TestTool",
          "version": "1.0.0",
          "rules": [
            {
              "id": "RULE001"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "RULE001",
          "ruleIndex": 0,
          "level": "warning",
          "message": {
            "text": "Test finding"
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

	metrics, err := ParseSARIF(sarifPath)

	if err != nil {
		t.Fatalf("ParseSARIF() error = %v", err)
	}

	// Verify rules summary
	rules, ok := metrics.Data["rules"].([]map[string]interface{})
	if !ok {
		t.Fatal("Expected rules to be []map[string]interface{}")
	}
	if len(rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(rules))
	}

	rule := rules[0]
	// When severity is empty, it should not be included in the rule summary
	if _, hasSeverity := rule["severity"]; hasSeverity {
		t.Error("Expected no severity field when security-severity is empty")
	}
}
