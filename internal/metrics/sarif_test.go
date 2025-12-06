package metrics

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSARIF(t *testing.T) {
	tests := []struct {
		name           string
		sarifContent   string
		wantErr        bool
		wantFindings   int
		wantErrors     int
		wantWarnings   int
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
