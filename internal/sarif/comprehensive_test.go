package sarif

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComprehensiveSARIF(t *testing.T) {
	// Test with a comprehensive SARIF document covering many features
	sarifContent := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "ComprehensiveTool",
          "version": "2.0.0",
          "informationUri": "https://example.com/tool",
          "rules": [
            {
              "id": "RULE1",
              "name": "Rule One",
              "shortDescription": {"text": "Short desc 1"},
              "fullDescription": {"text": "Full desc 1"},
              "help": {"text": "Help text 1"},
              "properties": {
                "tags": ["tag1", "tag2"],
                "precision": "very-high",
                "security-severity": "9.0"
              }
            },
            {
              "id": "RULE2",
              "name": "Rule Two",
              "shortDescription": {"text": "Short desc 2"},
              "fullDescription": {"text": "Full desc 2"},
              "help": {"text": "Help text 2"}
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "RULE1",
          "ruleIndex": 0,
          "level": "error",
          "message": {"text": "Error message"},
          "locations": [{
            "physicalLocation": {
              "artifactLocation": {"uri": "file1.go"},
              "region": {
                "startLine": 10,
                "startColumn": 5,
                "endLine": 10,
                "endColumn": 20
              }
            }
          }]
        },
        {
          "ruleId": "RULE2",
          "ruleIndex": 1,
          "level": "warning",
          "message": {"text": "Warning message"},
          "locations": [{
            "physicalLocation": {
              "artifactLocation": {"uri": "file2.go"},
              "region": {
                "startLine": 20,
                "startColumn": 10
              }
            }
          }]
        },
        {
          "ruleId": "RULE3",
          "level": "note",
          "message": {"text": "Note without rule definition"},
          "locations": [{
            "physicalLocation": {
              "artifactLocation": {"uri": "file3.go"},
              "region": {"startLine": 30}
            }
          }]
        }
      ]
    }
  ]
}`

	tmpDir := t.TempDir()
	sarifPath := filepath.Join(tmpDir, "comprehensive.sarif")
	
	if err := os.WriteFile(sarifPath, []byte(sarifContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test Parse
	doc, err := Parse(sarifPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify document structure
	if doc.Version != "2.1.0" {
		t.Errorf("Version = %s, want 2.1.0", doc.Version)
	}

	if len(doc.Runs) != 1 {
		t.Fatalf("Expected 1 run, got %d", len(doc.Runs))
	}

	run := doc.Runs[0]

	// Verify tool info
	if run.Tool.Driver.Name != "ComprehensiveTool" {
		t.Errorf("Tool name = %s, want ComprehensiveTool", run.Tool.Driver.Name)
	}

	if run.Tool.Driver.Version != "2.0.0" {
		t.Errorf("Tool version = %s, want 2.0.0", run.Tool.Driver.Version)
	}

	if run.Tool.Driver.InformationURI != "https://example.com/tool" {
		t.Errorf("Tool URI = %s", run.Tool.Driver.InformationURI)
	}

	// Verify rules
	if len(run.Tool.Driver.Rules) != 2 {
		t.Fatalf("Expected 2 rules, got %d", len(run.Tool.Driver.Rules))
	}

	rule1 := run.Tool.Driver.Rules[0]
	if rule1.ID != "RULE1" {
		t.Errorf("Rule1 ID = %s, want RULE1", rule1.ID)
	}

	if rule1.Name != "Rule One" {
		t.Errorf("Rule1 Name = %s, want 'Rule One'", rule1.Name)
	}

	if rule1.Properties == nil {
		t.Fatal("Expected rule1 to have properties")
	}

	if len(rule1.Properties.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(rule1.Properties.Tags))
	}

	if rule1.Properties.Precision != "very-high" {
		t.Errorf("Precision = %s, want very-high", rule1.Properties.Precision)
	}

	// Verify results
	if len(run.Results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(run.Results))
	}

	// Test GetFindings
	findings := doc.GetFindings()

	if len(findings) != 3 {
		t.Fatalf("Expected 3 findings, got %d", len(findings))
	}

	// Check first finding (with full rule metadata)
	f1 := findings[0]
	if f1.RuleID != "RULE1" {
		t.Errorf("Finding1 RuleID = %s, want RULE1", f1.RuleID)
	}

	if f1.RuleName != "Rule One" {
		t.Errorf("Finding1 RuleName = %s, want 'Rule One'", f1.RuleName)
	}

	if f1.ShortDesc != "Short desc 1" {
		t.Errorf("Finding1 ShortDesc = %s", f1.ShortDesc)
	}

	if f1.FullDesc != "Full desc 1" {
		t.Errorf("Finding1 FullDesc = %s", f1.FullDesc)
	}

	if f1.HelpText != "Help text 1" {
		t.Errorf("Finding1 HelpText = %s", f1.HelpText)
	}

	if len(f1.Tags) != 2 {
		t.Errorf("Finding1 expected 2 tags, got %d", len(f1.Tags))
	}

	if f1.Precision != "very-high" {
		t.Errorf("Finding1 Precision = %s", f1.Precision)
	}

	if f1.SecuritySeverity != "9.0" {
		t.Errorf("Finding1 SecuritySeverity = %s", f1.SecuritySeverity)
	}

	if f1.File != "file1.go" {
		t.Errorf("Finding1 File = %s, want file1.go", f1.File)
	}

	if f1.Line != 10 {
		t.Errorf("Finding1 Line = %d, want 10", f1.Line)
	}

	if f1.Column != 5 {
		t.Errorf("Finding1 Column = %d, want 5", f1.Column)
	}

	if f1.EndLine != 10 {
		t.Errorf("Finding1 EndLine = %d, want 10", f1.EndLine)
	}

	if f1.EndColumn != 20 {
		t.Errorf("Finding1 EndColumn = %d, want 20", f1.EndColumn)
	}

	// Check second finding (with rule but no properties)
	f2 := findings[1]
	if f2.RuleID != "RULE2" {
		t.Errorf("Finding2 RuleID = %s, want RULE2", f2.RuleID)
	}

	if f2.RuleName != "Rule Two" {
		t.Errorf("Finding2 RuleName = %s, want 'Rule Two'", f2.RuleName)
	}

	if f2.Level != "warning" {
		t.Errorf("Finding2 Level = %s, want warning", f2.Level)
	}

	// Properties should be empty for RULE2
	if f2.Precision != "" {
		t.Errorf("Finding2 Precision should be empty, got %s", f2.Precision)
	}

	if len(f2.Tags) != 0 {
		t.Errorf("Finding2 should have no tags, got %d", len(f2.Tags))
	}

	// Check third finding (no rule definition)
	f3 := findings[2]
	if f3.RuleID != "RULE3" {
		t.Errorf("Finding3 RuleID = %s, want RULE3", f3.RuleID)
	}

	// RuleName should default to RuleID when no rule definition exists
	if f3.RuleName != "RULE3" {
		t.Errorf("Finding3 RuleName = %s, want RULE3 (default)", f3.RuleName)
	}

	if f3.Level != "note" {
		t.Errorf("Finding3 Level = %s, want note", f3.Level)
	}

	// Metadata should be empty for undefined rule
	if f3.ShortDesc != "" {
		t.Errorf("Finding3 ShortDesc should be empty, got %s", f3.ShortDesc)
	}
}

func TestSARIFStructs(t *testing.T) {
	// Test that all struct types can be created
	_ = SARIF{
		Version: "2.1.0",
		Runs:    []Run{},
	}

	_ = Run{
		Tool:    Tool{},
		Results: []Result{},
	}

	_ = Tool{
		Driver: Driver{},
	}

	_ = Driver{
		Name:    "Test",
		Version: "1.0",
		Rules:   []Rule{},
	}

	_ = Rule{
		ID:               "TEST",
		Name:             "Test Rule",
		ShortDescription: MessageString{Text: "Short"},
		FullDescription:  MessageString{Text: "Full"},
		Help:             MessageString{Text: "Help"},
		Properties:       &RuleProperties{},
	}

	_ = RuleProperties{
		Tags:             []string{"tag"},
		Precision:        "high",
		SecuritySeverity: "7.0",
	}

	_ = Result{
		RuleID:           "TEST",
		Message:          Message{Text: "Message"},
		Locations:        []Location{},
		CodeFlows:        []CodeFlow{},
		RelatedLocations: []RelatedLocation{},
	}

	_ = Finding{
		RuleID:           "TEST",
		RuleName:         "Test",
		File:             "test.go",
		Line:             10,
		Message:          "Test message",
		Level:            "warning",
		DataFlowSteps:    []DataFlowStep{},
		SecuritySeverity: "7.0",
	}

	_ = DataFlowStep{
		File:    "test.go",
		Line:    10,
		Column:  5,
		Message: "Step",
	}

	// Test passes if all structs compile
	t.Log("All SARIF structs can be created")
}
