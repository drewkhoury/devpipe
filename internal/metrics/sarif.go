package metrics

import (
	"fmt"

	"github.com/drew/devpipe/internal/model"
	"github.com/drew/devpipe/internal/sarif"
)

// ParseSARIF parses a SARIF file and returns metrics
// SARIF (Static Analysis Results Interchange Format) is used by security scanners like CodeQL, gosec, etc.
func ParseSARIF(path string) (*model.TaskMetrics, error) {
	// Parse SARIF file
	doc, err := sarif.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SARIF: %w", err)
	}

	// Extract findings
	findings := doc.GetFindings()

	// Count by severity level
	var errors, warnings, notes int
	for _, f := range findings {
		switch f.Level {
		case "error":
			errors++
		case "warning":
			warnings++
		case "note":
			notes++
		default:
			warnings++ // Default to warning
		}
	}

	// Group findings by rule ID for summary
	ruleCount := make(map[string]int)
	ruleSeverity := make(map[string]string)
	for _, f := range findings {
		ruleCount[f.RuleID]++
		if _, exists := ruleSeverity[f.RuleID]; !exists {
			ruleSeverity[f.RuleID] = f.SecuritySeverity
		}
	}

	// Convert findings to serializable format
	var findingsData []map[string]interface{}
	for _, f := range findings {
		finding := map[string]interface{}{
			"ruleId":   f.RuleID,
			"ruleName": f.RuleName,
			"file":     f.File,
			"line":     f.Line,
			"column":   f.Column,
			"message":  f.Message,
			"level":    f.Level,
		}

		// Add optional fields if present
		if f.ShortDesc != "" {
			finding["shortDesc"] = f.ShortDesc
		}
		if f.SecuritySeverity != "" {
			finding["severity"] = f.SecuritySeverity
		}
		if len(f.Tags) > 0 {
			finding["tags"] = f.Tags
		}
		if f.Precision != "" {
			finding["precision"] = f.Precision
		}
		if f.SourceLocation != "" {
			finding["source"] = f.SourceLocation
		}
		if len(f.DataFlowSteps) > 0 {
			var steps []map[string]interface{}
			for _, step := range f.DataFlowSteps {
				steps = append(steps, map[string]interface{}{
					"file":    step.File,
					"line":    step.Line,
					"column":  step.Column,
					"message": step.Message,
				})
			}
			finding["dataFlow"] = steps
		}

		findingsData = append(findingsData, finding)
	}

	// Build rule summary
	var rules []map[string]interface{}
	for ruleID, count := range ruleCount {
		rule := map[string]interface{}{
			"id":    ruleID,
			"count": count,
		}
		if severity, ok := ruleSeverity[ruleID]; ok && severity != "" {
			rule["severity"] = severity
		}
		rules = append(rules, rule)
	}

	return &model.TaskMetrics{
		Kind:          "security",
		SummaryFormat: "sarif",
		Data: map[string]interface{}{
			"total":    len(findings),
			"errors":   errors,
			"warnings": warnings,
			"notes":    notes,
			"findings": findingsData,
			"rules":    rules,
		},
	}, nil
}
