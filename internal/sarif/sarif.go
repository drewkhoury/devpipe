package sarif

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SARIF represents the top-level SARIF document structure
type SARIF struct {
	Version string `json:"version"`
	Runs    []Run  `json:"runs"`
}

// Run represents a single analysis run
type Run struct {
	Tool    Tool     `json:"tool"`
	Results []Result `json:"results"`
}

// Tool represents the analysis tool information
type Tool struct {
	Driver Driver `json:"driver"`
}

// Driver represents the tool driver information
type Driver struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	InformationURI string `json:"informationUri"`
	Rules          []Rule `json:"rules"`
}

// Rule represents a rule definition
type Rule struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	ShortDescription MessageString   `json:"shortDescription"`
	FullDescription  MessageString   `json:"fullDescription"`
	Help             MessageString   `json:"help"`
	Properties       *RuleProperties `json:"properties,omitempty"`
}

// RuleProperties contains additional rule metadata
type RuleProperties struct {
	Tags             []string `json:"tags,omitempty"`
	Precision        string   `json:"precision,omitempty"`
	SecuritySeverity string   `json:"security-severity,omitempty"`
}

// Result represents a single finding
type Result struct {
	RuleID           string            `json:"ruleId"`
	RuleIndex        int               `json:"ruleIndex"`
	Message          Message           `json:"message"`
	Locations        []Location        `json:"locations"`
	Level            string            `json:"level,omitempty"`
	CodeFlows        []CodeFlow        `json:"codeFlows,omitempty"`
	RelatedLocations []RelatedLocation `json:"relatedLocations,omitempty"`
}

// Message represents a result message
type Message struct {
	Text string `json:"text"`
}

// MessageString represents a message with text
type MessageString struct {
	Text string `json:"text"`
}

// Location represents a location in source code
type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
	Message          *Message         `json:"message,omitempty"`
}

// RelatedLocation represents a related location
type RelatedLocation struct {
	ID               int              `json:"id,omitempty"`
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
	Message          Message          `json:"message"`
}

// CodeFlow represents a data flow path
type CodeFlow struct {
	ThreadFlows []ThreadFlow `json:"threadFlows"`
}

// ThreadFlow represents a thread of execution
type ThreadFlow struct {
	Locations []ThreadFlowLocation `json:"locations"`
}

// ThreadFlowLocation represents a location in a code flow
type ThreadFlowLocation struct {
	Location Location `json:"location"`
}

// PhysicalLocation represents a physical location in a file
type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           Region           `json:"region"`
}

// ArtifactLocation represents a file location
type ArtifactLocation struct {
	URI   string `json:"uri"`
	Index int    `json:"index,omitempty"`
}

// Region represents a region in a file
type Region struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn,omitempty"`
	EndLine     int `json:"endLine,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

// Finding represents a simplified view of a SARIF result
type Finding struct {
	RuleID           string
	RuleName         string
	File             string
	Line             int
	Column           int
	EndLine          int
	EndColumn        int
	Message          string
	Level            string
	ShortDesc        string
	FullDesc         string
	HelpText         string
	Tags             []string
	Precision        string
	SecuritySeverity string
	DataFlowSteps    []DataFlowStep
	SourceLocation   string
}

// DataFlowStep represents a step in the data flow
type DataFlowStep struct {
	File    string
	Line    int
	Column  int
	Message string
}

// Parse reads and parses a SARIF file
func Parse(filename string) (*SARIF, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var sarif SARIF
	if err := json.Unmarshal(data, &sarif); err != nil {
		return nil, fmt.Errorf("failed to parse SARIF: %w", err)
	}

	return &sarif, nil
}

// GetFindings extracts all findings from a SARIF document
func (s *SARIF) GetFindings() []Finding {
	var findings []Finding

	for _, run := range s.Runs {
		// Build a map of rule IDs to rule metadata
		ruleMap := make(map[string]*Rule)
		for i := range run.Tool.Driver.Rules {
			rule := &run.Tool.Driver.Rules[i]
			ruleMap[rule.ID] = rule
		}

		for _, result := range run.Results {
			if len(result.Locations) == 0 {
				continue
			}

			loc := result.Locations[0].PhysicalLocation

			// Get rule metadata
			rule := ruleMap[result.RuleID]
			ruleName := result.RuleID
			var shortDesc, fullDesc, helpText string
			var tags []string
			var precision, securitySeverity string

			if rule != nil {
				if rule.Name != "" {
					ruleName = rule.Name
				}
				shortDesc = rule.ShortDescription.Text
				fullDesc = rule.FullDescription.Text
				helpText = rule.Help.Text

				if rule.Properties != nil {
					tags = rule.Properties.Tags
					precision = rule.Properties.Precision
					securitySeverity = rule.Properties.SecuritySeverity
				}
			}

			level := result.Level
			if level == "" {
				level = "warning"
			}

			// Extract data flow steps
			var dataFlowSteps []DataFlowStep
			if len(result.CodeFlows) > 0 && len(result.CodeFlows[0].ThreadFlows) > 0 {
				for _, tfLoc := range result.CodeFlows[0].ThreadFlows[0].Locations {
					msg := ""
					if tfLoc.Location.Message != nil {
						msg = tfLoc.Location.Message.Text
					}
					dataFlowSteps = append(dataFlowSteps, DataFlowStep{
						File:    tfLoc.Location.PhysicalLocation.ArtifactLocation.URI,
						Line:    tfLoc.Location.PhysicalLocation.Region.StartLine,
						Column:  tfLoc.Location.PhysicalLocation.Region.StartColumn,
						Message: msg,
					})
				}
			}

			// Extract source location from related locations
			var sourceLocation string
			if len(result.RelatedLocations) > 0 {
				relLoc := result.RelatedLocations[0]
				sourceLocation = fmt.Sprintf("%s:%d:%d - %s",
					relLoc.PhysicalLocation.ArtifactLocation.URI,
					relLoc.PhysicalLocation.Region.StartLine,
					relLoc.PhysicalLocation.Region.StartColumn,
					relLoc.Message.Text)
			}

			findings = append(findings, Finding{
				RuleID:           result.RuleID,
				RuleName:         ruleName,
				File:             loc.ArtifactLocation.URI,
				Line:             loc.Region.StartLine,
				Column:           loc.Region.StartColumn,
				EndLine:          loc.Region.EndLine,
				EndColumn:        loc.Region.EndColumn,
				Message:          result.Message.Text,
				Level:            level,
				ShortDesc:        shortDesc,
				FullDesc:         fullDesc,
				HelpText:         helpText,
				Tags:             tags,
				Precision:        precision,
				SecuritySeverity: securitySeverity,
				DataFlowSteps:    dataFlowSteps,
				SourceLocation:   sourceLocation,
			})
		}
	}

	// Sort findings by file, then line
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		return findings[i].Line < findings[j].Line
	})

	return findings
}

// PrintFindings prints findings in a human-readable format
func PrintFindings(findings []Finding, verbose bool) {
	if len(findings) == 0 {
		fmt.Println("✅ No security issues found!")
		return
	}

	fmt.Printf("⚠️  Found %d security issue(s):\n\n", len(findings))

	for i, f := range findings {
		if i > 0 {
			fmt.Println()
		}

		// Color code by level
		levelIcon := "⚠️ "
		switch strings.ToLower(f.Level) {
		case "error":
			levelIcon = "❌"
		case "warning":
			levelIcon = "⚠️ "
		case "note":
			levelIcon = "ℹ️ "
		}

		// Basic info (always shown)
		fmt.Printf("%s Rule:    %s\n", levelIcon, f.RuleID)

		// Location with column info if available
		if f.Column > 0 {
			if f.EndLine > 0 && f.EndColumn > 0 {
				fmt.Printf("   File:    %s:%d:%d-%d:%d\n", f.File, f.Line, f.Column, f.EndLine, f.EndColumn)
			} else {
				fmt.Printf("   File:    %s:%d:%d\n", f.File, f.Line, f.Column)
			}
		} else {
			fmt.Printf("   File:    %s:%d\n", f.File, f.Line)
		}

		fmt.Printf("   Message: %s\n", f.Message)

		// Verbose mode - show additional details
		if verbose {
			if f.RuleName != "" && f.RuleName != f.RuleID {
				fmt.Printf("   Name:    %s\n", f.RuleName)
			}

			if f.Level != "" {
				fmt.Printf("   Level:   %s\n", f.Level)
			}

			if f.ShortDesc != "" {
				fmt.Printf("   Info:    %s\n", f.ShortDesc)
			}

			if f.Precision != "" {
				fmt.Printf("   Precision: %s\n", f.Precision)
			}

			if f.SecuritySeverity != "" {
				fmt.Printf("   Severity:  %s\n", f.SecuritySeverity)
			}

			if len(f.Tags) > 0 {
				fmt.Printf("   Tags:    %s\n", strings.Join(f.Tags, ", "))
			}

			if f.FullDesc != "" && f.FullDesc != f.ShortDesc {
				fmt.Printf("   Details: %s\n", f.FullDesc)
			}

			if f.HelpText != "" {
				// Truncate help text if too long
				help := f.HelpText
				if len(help) > 200 {
					help = help[:197] + "..."
				}
				fmt.Printf("   Help:    %s\n", help)
			}

			// Show source location (where tainted data originates)
			if f.SourceLocation != "" {
				fmt.Printf("   Source:  %s\n", f.SourceLocation)
			}

			// Show data flow path
			if len(f.DataFlowSteps) > 0 {
				fmt.Printf("   Data Flow Path (%d steps):\n", len(f.DataFlowSteps))
				for i, step := range f.DataFlowSteps {
					stepNum := i + 1
					if step.Column > 0 {
						fmt.Printf("     %d. %s:%d:%d", stepNum, step.File, step.Line, step.Column)
					} else {
						fmt.Printf("     %d. %s:%d", stepNum, step.File, step.Line)
					}
					if step.Message != "" {
						fmt.Printf(" - %s", step.Message)
					}
					fmt.Println()
				}
			}
		}
	}
}

// PrintSummary prints a summary of findings grouped by rule
func PrintSummary(findings []Finding) {
	if len(findings) == 0 {
		fmt.Println("✅ No security issues found!")
		return
	}

	// Group by rule ID
	ruleCount := make(map[string]int)
	for _, f := range findings {
		ruleCount[f.RuleID]++
	}

	// Sort rules by count (descending)
	type ruleStats struct {
		id    string
		count int
	}
	var stats []ruleStats
	for id, count := range ruleCount {
		stats = append(stats, ruleStats{id, count})
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].count != stats[j].count {
			return stats[i].count > stats[j].count
		}
		return stats[i].id < stats[j].id
	})

	fmt.Printf("📊 Security Issues Summary (%d total):\n\n", len(findings))
	for _, s := range stats {
		fmt.Printf("  %3d  %s\n", s.count, s.id)
	}
}

// FindSARIFFiles finds all SARIF files in a directory
func FindSARIFFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".sarif") {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}
