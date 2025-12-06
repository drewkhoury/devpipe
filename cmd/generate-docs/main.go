// Copyright 2025 Andrew Khoury
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// generate-docs generates documentation from config structs using reflection
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/drew/devpipe/internal/config"
)

// FieldDoc represents documentation for a single field
type FieldDoc struct {
	Name        string
	Type        string
	Required    bool
	Default     string
	Description string
	ValidValues []string
}

// SectionDoc represents documentation for a config section
type SectionDoc struct {
	Name        string
	Description string
	Fields      []FieldDoc
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Usage: generate-docs")
		fmt.Println("Generates documentation from config structs:")
		fmt.Println("  - config.example.toml")
		fmt.Println("  - config.schema.json")
		fmt.Println("  - docs/configuration.md")
		fmt.Println("  - docs/cli-reference.md")
		fmt.Println("  - docs/config-validation.md")
		return
	}

	// Define documentation structure
	docs := buildDocumentation()

	// Generate outputs
	if err := generateExampleTOML(docs); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating config.example.toml: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated config.example.toml")

	if err := generateJSONSchema(docs); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating config.schema.json: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated config.schema.json")

	if err := generateMarkdownDocs(docs); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating docs/configuration.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated docs/configuration.md")

	if err := generateCLIDocs(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating docs/cli-reference.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated docs/cli-reference.md")

	if err := generateValidationDocs(); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating docs/config-validation.md: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Generated docs/config-validation.md")
}

// readSnippet reads a markdown snippet from the snippets directory
func readSnippet(name string) (string, error) {
	content, err := os.ReadFile(fmt.Sprintf("cmd/generate-docs/snippets/%s", name))
	if err != nil {
		return "", fmt.Errorf("failed to read snippet %s: %w", name, err)
	}
	return string(content), nil
}

func buildDocumentation() []SectionDoc {
	defaults := config.GetDefaults()

	return []SectionDoc{
		extractSection("defaults", "Global configuration options", defaults.Defaults, defaults.Defaults),
		extractSection("defaults.git", "Git integration settings", defaults.Defaults.Git, defaults.Defaults.Git),
		extractSection("task_defaults", "Default values that apply to all tasks unless overridden at the task level", defaults.TaskDefaults, defaults.TaskDefaults),
		extractSection("tasks.<task-id>", "Individual task configuration. Task ID must be unique.", config.TaskConfig{}, config.TaskConfig{}),
	}
}

// extractSection uses reflection to extract field documentation from struct tags
func extractSection(name, description string, value interface{}, defaultValue interface{}) SectionDoc {
	section := SectionDoc{
		Name:        name,
		Description: description,
		Fields:      []FieldDoc{},
	}

	t := reflect.TypeOf(value)
	v := reflect.ValueOf(defaultValue)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip internal fields (no doc tag)
		docTag := field.Tag.Get("doc")
		if docTag == "" {
			continue
		}

		tomlTag := field.Tag.Get("toml")
		if tomlTag == "" {
			continue
		}

		fieldDoc := FieldDoc{
			Name:        tomlTag,
			Type:        getFieldType(field.Type),
			Required:    field.Tag.Get("required") == "true",
			Description: docTag,
		}

		// Extract default value
		fieldValue := v.Field(i)
		fieldDoc.Default = getDefaultValue(fieldValue, field.Type)

		// Extract enum values
		enumTag := field.Tag.Get("enum")
		if enumTag != "" {
			fieldDoc.ValidValues = strings.Split(enumTag, ",")
		}

		section.Fields = append(section.Fields, fieldDoc)
	}

	return section
}

// getFieldType returns a string representation of the field type
func getFieldType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Bool:
		return "bool"
	case reflect.Ptr:
		return getFieldType(t.Elem())
	default:
		return t.String()
	}
}

// getDefaultValue returns a string representation of the default value
func getDefaultValue(v reflect.Value, t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Bool:
		if v.Bool() {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func generateExampleTOML(docs []SectionDoc) error {
	var sb strings.Builder

	sb.WriteString(`# =============================================================================
# devpipe Configuration Reference
# =============================================================================
# This is a comprehensive example showing ALL available configuration options.
# Copy sections you need to your own config.toml.
#
# Quick Start:
#   [tasks.my-task]
#   command = "npm test"
#
# Full documentation: https://github.com/drewkhoury/devpipe
# =============================================================================

`)

	for _, section := range docs {
		// Section header
		sb.WriteString("# -----------------------------------------------------------------------------\n")
		sb.WriteString(fmt.Sprintf("# [%s] - %s\n", section.Name, section.Description))
		sb.WriteString("# -----------------------------------------------------------------------------\n\n")

		// Write section
		if section.Name == "tasks.<task-id>" {
			sb.WriteString("# Example task with all options:\n")
			sb.WriteString("[tasks.example-task]\n")
		} else {
			sb.WriteString(fmt.Sprintf("[%s]\n", section.Name))
		}

		// Write fields
		for _, field := range section.Fields {
			// Write description as comment
			sb.WriteString(fmt.Sprintf("# %s\n", field.Description))
			if field.Required {
				sb.WriteString("# Required: yes\n")
			} else {
				sb.WriteString(fmt.Sprintf("# Default: %s\n", field.Default))
			}
			if len(field.ValidValues) > 0 {
				sb.WriteString(fmt.Sprintf("# Valid values: %s\n", strings.Join(field.ValidValues, ", ")))
			}

			// Write field with default value
			value := field.Default
			if field.Type == "string" && value != "" && value != "(inherited)" {
				value = fmt.Sprintf(`"%s"`, value)
			}
			if value == "" || value == "(inherited)" {
				sb.WriteString(fmt.Sprintf("# %s = \n", field.Name))
			} else {
				sb.WriteString(fmt.Sprintf("%s = %s\n", field.Name, value))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("\n")
	}

	// Add phase example
	sb.WriteString(`# -----------------------------------------------------------------------------
# Phase-Based Execution
# -----------------------------------------------------------------------------
# Use [tasks.phase-<name>] to create phase headers.
# Tasks under a phase header run in parallel. Phases execute sequentially.

[tasks.phase-quality]
name = "Quality Checks"
desc = "Linting and formatting"

[tasks.lint]
command = "npm run lint"
type = "check"

[tasks.format]
command = "npm run format:check"
type = "check"

[tasks.phase-build]
name = "Build"

[tasks.build]
command = "npm run build"
type = "build"

[tasks.phase-test]
name = "Tests"

[tasks.unit-tests]
command = "npm test"
type = "test"
metricsFormat = "junit"
metricsPath = "test-results/junit.xml"

[tasks.e2e-tests]
command = "npm run test:e2e"
type = "test"
`)

	return os.WriteFile("config.example.toml", []byte(sb.String()), 0644)
}

func generateJSONSchema(docs []SectionDoc) error {
	schema := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-07/schema#",
		"title":       "devpipe Configuration",
		"description": "Configuration schema for devpipe pipeline runner",
		"type":        "object",
		"properties":  make(map[string]interface{}),
	}

	properties := schema["properties"].(map[string]interface{})

	for _, section := range docs {
		sectionName := section.Name
		if strings.Contains(sectionName, ".") {
			// Handle nested sections like "defaults.git"
			continue // We'll handle these separately
		}

		sectionProps := map[string]interface{}{
			"type":        "object",
			"description": section.Description,
			"properties":  make(map[string]interface{}),
		}

		sectionFields := sectionProps["properties"].(map[string]interface{})

		for _, field := range section.Fields {
			fieldSchema := map[string]interface{}{
				"description": field.Description,
			}

			// Map type
			switch field.Type {
			case "string":
				fieldSchema["type"] = "string"
			case "int":
				fieldSchema["type"] = "integer"
			case "bool":
				fieldSchema["type"] = "boolean"
			}

			// Add default
			if field.Default != "" && field.Default != "(inherited)" {
				switch field.Type {
				case "string":
					fieldSchema["default"] = field.Default
				case "int":
					var intVal int
					_, _ = fmt.Sscanf(field.Default, "%d", &intVal) // Best effort parsing
					fieldSchema["default"] = intVal
				case "bool":
					fieldSchema["default"] = field.Default == "true"
				}
			}

			// Add enum for valid values
			if len(field.ValidValues) > 0 {
				fieldSchema["enum"] = field.ValidValues
			}

			sectionFields[field.Name] = fieldSchema
		}

		// Handle nested git section
		if sectionName == "defaults" {
			gitProps := map[string]interface{}{
				"type":        "object",
				"description": "Git integration settings",
				"properties":  make(map[string]interface{}),
			}

			// Find git section in docs
			for _, s := range docs {
				if s.Name == "defaults.git" {
					gitFields := gitProps["properties"].(map[string]interface{})
					for _, field := range s.Fields {
						fieldSchema := map[string]interface{}{
							"type":        "string",
							"description": field.Description,
						}
						if field.Default != "" {
							fieldSchema["default"] = field.Default
						}
						if len(field.ValidValues) > 0 {
							fieldSchema["enum"] = field.ValidValues
						}
						gitFields[field.Name] = fieldSchema
					}
				}
			}

			sectionFields["git"] = gitProps
		}

		properties[sectionName] = sectionProps
	}

	// Add tasks section
	tasksSection := docs[len(docs)-1] // tasks.<task-id> is last
	taskSchema := map[string]interface{}{
		"type":        "object",
		"description": tasksSection.Description,
		"properties":  make(map[string]interface{}),
	}

	taskProps := taskSchema["properties"].(map[string]interface{})
	for _, field := range tasksSection.Fields {
		fieldSchema := map[string]interface{}{
			"description": field.Description,
		}
		switch field.Type {
		case "string":
			fieldSchema["type"] = "string"
		case "bool":
			fieldSchema["type"] = "boolean"
		}
		if len(field.ValidValues) > 0 {
			fieldSchema["enum"] = field.ValidValues
		}
		taskProps[field.Name] = fieldSchema
	}

	properties["tasks"] = map[string]interface{}{
		"type": "object",
		"patternProperties": map[string]interface{}{
			"^[a-zA-Z0-9_-]+$": taskSchema,
		},
	}

	// Write to file
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("config.schema.json", data, 0644)
}

func generateMarkdownDocs(docs []SectionDoc) error {
	var sb strings.Builder

	// Add intro from snippet
	intro, err := readSnippet("config-intro.md")
	if err != nil {
		return err
	}
	sb.WriteString(intro)
	sb.WriteString("\n\n")

	// Generate tables from struct tags
	for _, section := range docs {
		sb.WriteString("### `[" + section.Name + "]`\n\n")
		sb.WriteString(section.Description + "\n\n")

		sb.WriteString("| Field | Type | Required | Default | Description |\n")
		sb.WriteString("|-------|------|----------|---------|-------------|\n")

		for _, field := range section.Fields {
			required := "No"
			if field.Required {
				required = "**Yes**"
			}
			defaultVal := field.Default
			if defaultVal == "" {
				defaultVal = "-"
			}
			desc := field.Description
			if len(field.ValidValues) > 0 {
				desc += fmt.Sprintf(" (valid: `%s`)", strings.Join(field.ValidValues, "`, `"))
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | `%s` | %s |\n",
				field.Name, field.Type, required, defaultVal, desc))
		}

		sb.WriteString("\n")
	}

	// Add phase examples from snippet
	phaseExamples, err := readSnippet("phase-examples.md")
	if err != nil {
		return err
	}
	sb.WriteString(phaseExamples)
	sb.WriteString("\n")

	// Create docs directory if it doesn't exist
	if err := os.MkdirAll("docs", 0755); err != nil {
		return err
	}

	return os.WriteFile("docs/configuration.md", []byte(sb.String()), 0644)
}

func generateCLIDocs() error {
	var sb strings.Builder

	// Add intro from snippet
	intro, err := readSnippet("cli-intro.md")
	if err != nil {
		return err
	}
	sb.WriteString(intro)
	sb.WriteString("\n\n")

	// TODO: Add auto-generated flag tables here when we parse CLI flags
	sb.WriteString("### Run Flags\n\n")
	sb.WriteString("| Flag | Description | Default |\n")
	sb.WriteString("|------|-------------|---------||\n")
	sb.WriteString("| `--config <path>` | Path to config file | `config.toml` |\n")
	sb.WriteString("| `--since <ref>` | Git ref to compare against (overrides config) | - |\n")
	sb.WriteString("| `--only <task-id>` | Run only a single task by id | - |\n")
	sb.WriteString("| `--skip <task-id>` | Skip a task by id (repeatable) | - |\n")
	sb.WriteString("| `--fix-type <type>` | Fix behavior: `auto`, `helper`, `none` (overrides config) | - |\n")
	sb.WriteString("| `--ui <mode>` | UI mode: `basic`, `full` | `basic` |\n")
	sb.WriteString("| `--dashboard` | Show dashboard with live progress | `false` |\n")
	sb.WriteString("| `--fail-fast` | Stop on first task failure | `false` |\n")
	sb.WriteString("| `--fast` | Skip long-running tasks (> fastThreshold) | `false` |\n")
	sb.WriteString("| `--dry-run` | Do not execute commands, simulate only | `false` |\n")
	sb.WriteString("| `--verbose` | Show verbose output (always logged to pipeline.log) | `false` |\n")
	sb.WriteString("| `--no-color` | Disable colored output | `false` |\n")
	sb.WriteString("\n")

	sb.WriteString("### Validate Flags\n\n")
	sb.WriteString("| Flag | Description | Default |\n")
	sb.WriteString("|------|-------------|---------||\n")
	sb.WriteString("| `--config <path>` | Path to config file to validate (supports multiple files) | `config.toml` |\n")
	sb.WriteString("\n")
	sb.WriteString("See [config-validation.md](config-validation.md) for more details.\n\n")

	// Add examples from snippet
	examples, err := readSnippet("cli-examples.md")
	if err != nil {
		return err
	}
	sb.WriteString(examples)
	sb.WriteString("\n")

	// Create docs directory if it doesn't exist
	if err := os.MkdirAll("docs", 0755); err != nil {
		return err
	}

	return os.WriteFile("docs/cli-reference.md", []byte(sb.String()), 0644)
}

func generateValidationDocs() error {
	// Simply copy the snippet to docs
	content, err := readSnippet("config-validation.md")
	if err != nil {
		return err
	}

	// Create docs directory if it doesn't exist
	if err := os.MkdirAll("docs", 0755); err != nil {
		return err
	}

	return os.WriteFile("docs/config-validation.md", []byte(content), 0644)
}
