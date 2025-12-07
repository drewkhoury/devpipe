package features

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

type commandsContext struct {
	*sharedContext
	configPaths []string
	sarifPath   string
	taskIDs     []string
}

func (c *commandsContext) aConfigWithMultipleTasks() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-commands-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	c.taskIDs = []string{"lint", "format", "test", "build"}

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[tasks.lint]
name = "Lint Code"
command = "echo linting"
type = "check-lint"

[tasks.format]
name = "Format Code"
command = "echo formatting"
type = "check-format"

[tasks.test]
name = "Run Tests"
command = "echo testing"
type = "test-unit"

[tasks.build]
name = "Build App"
command = "echo building"
type = "build"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *commandsContext) iRunDevpipeList() error {
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "list")
	cmd.Dir = c.tempDir
	output, err := cmd.CombinedOutput()
	c.output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitErr.ExitCode()
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *commandsContext) theOutputShouldContainAllTaskIDs() error {
	for _, taskID := range c.taskIDs {
		if !strings.Contains(c.output, taskID) {
			return fmt.Errorf("expected output to contain task ID %q, got: %s", taskID, c.output)
		}
	}
	return nil
}

func (c *commandsContext) iRunDevpipeListVerbose() error {
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "list", "--verbose")
	cmd.Dir = c.tempDir
	output, err := cmd.CombinedOutput()
	c.output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitErr.ExitCode()
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *commandsContext) theOutputShouldShowATableFormat() error {
	// Check for table formatting indicators (borders, headers, etc.)
	if !strings.Contains(c.output, "ID") && !strings.Contains(c.output, "Name") {
		return fmt.Errorf("expected output to show table format with headers, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) theOutputShouldContainTaskNames() error {
	taskNames := []string{"Lint Code", "Format Code", "Run Tests", "Build App"}
	for _, name := range taskNames {
		if !strings.Contains(c.output, name) {
			return fmt.Errorf("expected output to contain task name %q, got: %s", name, c.output)
		}
	}
	return nil
}

func (c *commandsContext) theOutputShouldContainTaskTypes() error {
	taskTypes := []string{"check-lint", "check-format", "test-unit", "build"}
	foundCount := 0
	for _, taskType := range taskTypes {
		if strings.Contains(c.output, taskType) {
			foundCount++
		}
	}
	if foundCount == 0 {
		return fmt.Errorf("expected output to contain task types, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) noConfigFileExists() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-commands-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "nonexistent-config.toml")
	return nil
}

func (c *commandsContext) theOutputShouldIndicateConfigFileNotFound() error {
	if !strings.Contains(strings.ToLower(c.output), "not found") && !strings.Contains(strings.ToLower(c.output), "error") {
		return fmt.Errorf("expected output to indicate config file not found, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) iRunDevpipeVersion() error {
	cmd := exec.Command(c.devpipeBinary, "version")
	output, err := cmd.CombinedOutput()
	c.output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitErr.ExitCode()
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *commandsContext) theOutputShouldContainVersionNumber() error {
	// Check for version patterns like "v1.0.0", "version", or semantic version
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "version") && !strings.Contains(c.output, "v") {
		return fmt.Errorf("expected output to contain version information, got: %s", c.output)
	}
	return nil
}

// Validate command steps
func (c *commandsContext) aValidConfigFileForValidation() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validate-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "valid-config.toml")
	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[tasks.test]
name = "Test Task"
command = "echo test"
type = "test-unit"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *commandsContext) anInvalidConfigFileWithMissingRequiredField() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validate-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "invalid-config.toml")
	// Missing required 'command' field
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.test]
name = "Test Task"
type = "test-unit"
`

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *commandsContext) multipleConfigFilesWithMixedValidity() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validate-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	// Valid config
	validPath := filepath.Join(c.tempDir, "valid.toml")
	validConfig := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[tasks.test]
name = "Test Task"
command = "echo test"
type = "test-unit"
`, c.tempDir)
	if err := os.WriteFile(validPath, []byte(validConfig), 0644); err != nil {
		return err
	}

	// Invalid config
	invalidPath := filepath.Join(c.tempDir, "invalid.toml")
	invalidConfig := `[tasks.test]
name = "Test Task"
`
	if err := os.WriteFile(invalidPath, []byte(invalidConfig), 0644); err != nil {
		return err
	}

	c.configPaths = []string{validPath, invalidPath}
	return nil
}

func (c *commandsContext) iRunDevpipeValidateCommand() error {
	cmd := exec.Command(c.devpipeBinary, "validate", c.configPath)
	cmd.Dir = c.tempDir
	output, err := cmd.CombinedOutput()
	c.output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitErr.ExitCode()
		} else {
			// Command failed to execute
			return fmt.Errorf("failed to execute command: %v", err)
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *commandsContext) iRunDevpipeValidateWithMultipleFiles() error {
	args := append([]string{"validate"}, c.configPaths...)
	cmd := exec.Command(c.devpipeBinary, args...)
	cmd.Dir = c.tempDir
	output, err := cmd.CombinedOutput()
	c.output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitErr.ExitCode()
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *commandsContext) noConfigFileExistsForValidation() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validate-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "nonexistent.toml")
	return nil
}

func (c *commandsContext) iRunDevpipeValidateWithNonexistentFile() error {
	cmd := exec.Command(c.devpipeBinary, "validate", c.configPath)
	cmd.Dir = c.tempDir
	output, err := cmd.CombinedOutput()
	c.output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitErr.ExitCode()
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *commandsContext) theOutputShouldIndicateValidationPassed() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "valid") || strings.Contains(lowerOutput, "error") {
		return fmt.Errorf("expected output to indicate validation passed, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) theOutputShouldShowValidationErrors() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "error") {
		return fmt.Errorf("expected output to show validation errors, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) theOutputShouldShowWhichFilesFailed() error {
	if !strings.Contains(c.output, "invalid.toml") {
		return fmt.Errorf("expected output to show which files failed, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) theOutputShouldIndicateFileNotFound() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "not found") && !strings.Contains(lowerOutput, "no such file") && !strings.Contains(lowerOutput, "error") {
		return fmt.Errorf("expected output to indicate file not found, got: %s", c.output)
	}
	return nil
}

// SARIF command steps
func (c *commandsContext) aSARIFFileWithSecurityFindings() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-sarif-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.sarifPath = filepath.Join(c.tempDir, "results.sarif")
	sarif := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "TestScanner",
          "version": "1.0.0",
          "rules": [
            {
              "id": "TEST001",
              "name": "TestVulnerability",
              "shortDescription": {
                "text": "Test security vulnerability"
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "TEST001",
          "level": "error",
          "message": {
            "text": "Potential security issue detected"
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

	return os.WriteFile(c.sarifPath, []byte(sarif), 0644)
}

func (c *commandsContext) aSARIFFileWithMultipleFindings() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-sarif-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.sarifPath = filepath.Join(c.tempDir, "results.sarif")
	sarif := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "TestScanner",
          "version": "1.0.0",
          "rules": [
            {
              "id": "TEST001",
              "name": "TestVulnerability1",
              "shortDescription": {
                "text": "Test security vulnerability 1"
              }
            },
            {
              "id": "TEST002",
              "name": "TestVulnerability2",
              "shortDescription": {
                "text": "Test security vulnerability 2"
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "TEST001",
          "level": "error",
          "message": {
            "text": "First security issue"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test1.go"
                },
                "region": {
                  "startLine": 10
                }
              }
            }
          ]
        },
        {
          "ruleId": "TEST001",
          "level": "error",
          "message": {
            "text": "Second security issue"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test2.go"
                },
                "region": {
                  "startLine": 20
                }
              }
            }
          ]
        },
        {
          "ruleId": "TEST002",
          "level": "warning",
          "message": {
            "text": "Third security issue"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test3.go"
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

	return os.WriteFile(c.sarifPath, []byte(sarif), 0644)
}

func (c *commandsContext) noSARIFFileExists() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-sarif-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.sarifPath = filepath.Join(c.tempDir, "nonexistent.sarif")
	return nil
}

func (c *commandsContext) iRunDevpipeSarifWithTheFile() error {
	cmd := exec.Command(c.devpipeBinary, "sarif", c.sarifPath)
	cmd.Dir = c.tempDir
	output, err := cmd.CombinedOutput()
	c.output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitErr.ExitCode()
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *commandsContext) iRunDevpipeSarifWithSummaryFlag() error {
	cmd := exec.Command(c.devpipeBinary, "sarif", "-s", c.sarifPath)
	cmd.Dir = c.tempDir
	output, err := cmd.CombinedOutput()
	c.output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitErr.ExitCode()
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *commandsContext) iRunDevpipeSarifWithVerboseFlag() error {
	cmd := exec.Command(c.devpipeBinary, "sarif", "-v", c.sarifPath)
	cmd.Dir = c.tempDir
	output, err := cmd.CombinedOutput()
	c.output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitErr.ExitCode()
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *commandsContext) iRunDevpipeSarifWithNonexistentFile() error {
	cmd := exec.Command(c.devpipeBinary, "sarif", c.sarifPath)
	cmd.Dir = c.tempDir
	output, err := cmd.CombinedOutput()
	c.output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitErr.ExitCode()
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *commandsContext) theOutputShouldDisplayTheFindings() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "test001") && !strings.Contains(lowerOutput, "security") && !strings.Contains(lowerOutput, "vulnerability") {
		return fmt.Errorf("expected output to display security findings, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) theOutputShouldShowGroupedSummary() error {
	lowerOutput := strings.ToLower(c.output)
	// Summary mode should show rule counts or grouped information
	if !strings.Contains(lowerOutput, "test001") || !strings.Contains(lowerOutput, "test002") {
		return fmt.Errorf("expected output to show grouped summary, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) theOutputShouldShowDetailedMetadata() error {
	// Verbose mode should show more details
	if len(c.output) < 100 {
		return fmt.Errorf("expected verbose output to show detailed metadata, got short output: %s", c.output)
	}
	return nil
}

func (c *commandsContext) theOutputShouldIndicateSARIFFileNotFound() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "not found") && !strings.Contains(lowerOutput, "no such file") && !strings.Contains(lowerOutput, "error") {
		return fmt.Errorf("expected output to indicate SARIF file not found, got: %s", c.output)
	}
	return nil
}

func InitializeCommandsScenario(ctx *godog.ScenarioContext, shared *sharedContext) {
	c := &commandsContext{sharedContext: shared}

	// List command steps
	ctx.Step(`^a config with multiple tasks$`, c.aConfigWithMultipleTasks)
	ctx.Step(`^I run devpipe list$`, c.iRunDevpipeList)
	ctx.Step(`^the execution should succeed$`, shared.theExecutionShouldSucceed)
	ctx.Step(`^the output should contain all task IDs$`, c.theOutputShouldContainAllTaskIDs)
	ctx.Step(`^I run devpipe list --verbose$`, c.iRunDevpipeListVerbose)
	ctx.Step(`^the output should show a table format$`, c.theOutputShouldShowATableFormat)
	ctx.Step(`^the output should contain task names$`, c.theOutputShouldContainTaskNames)
	ctx.Step(`^the output should contain task types$`, c.theOutputShouldContainTaskTypes)
	ctx.Step(`^no config file exists$`, c.noConfigFileExists)
	ctx.Step(`^the execution should fail$`, shared.theExecutionShouldFail)
	ctx.Step(`^the output should indicate config file not found$`, c.theOutputShouldIndicateConfigFileNotFound)
	ctx.Step(`^I run devpipe version$`, c.iRunDevpipeVersion)
	ctx.Step(`^the output should contain version number$`, c.theOutputShouldContainVersionNumber)
	ctx.Step(`^the output should contain "([^"]*)"$`, shared.theOutputShouldContain)

	// Validate command steps
	ctx.Step(`^a valid config file for validation$`, c.aValidConfigFileForValidation)
	ctx.Step(`^an invalid config file with missing required field$`, c.anInvalidConfigFileWithMissingRequiredField)
	ctx.Step(`^multiple config files with mixed validity$`, c.multipleConfigFilesWithMixedValidity)
	ctx.Step(`^no config file exists for validation$`, c.noConfigFileExistsForValidation)
	ctx.Step(`^I run devpipe validate command$`, c.iRunDevpipeValidateCommand)
	ctx.Step(`^I run devpipe validate with multiple files$`, c.iRunDevpipeValidateWithMultipleFiles)
	ctx.Step(`^I run devpipe validate with nonexistent file$`, c.iRunDevpipeValidateWithNonexistentFile)
	ctx.Step(`^the output should indicate validation passed$`, c.theOutputShouldIndicateValidationPassed)
	ctx.Step(`^the output should show validation errors$`, c.theOutputShouldShowValidationErrors)
	ctx.Step(`^the output should show which files failed$`, c.theOutputShouldShowWhichFilesFailed)
	ctx.Step(`^the output should indicate file not found$`, c.theOutputShouldIndicateFileNotFound)

	// SARIF command steps
	ctx.Step(`^a SARIF file with security findings$`, c.aSARIFFileWithSecurityFindings)
	ctx.Step(`^a SARIF file with multiple findings$`, c.aSARIFFileWithMultipleFindings)
	ctx.Step(`^no SARIF file exists$`, c.noSARIFFileExists)
	ctx.Step(`^I run devpipe sarif with the file$`, c.iRunDevpipeSarifWithTheFile)
	ctx.Step(`^I run devpipe sarif with summary flag$`, c.iRunDevpipeSarifWithSummaryFlag)
	ctx.Step(`^I run devpipe sarif with verbose flag$`, c.iRunDevpipeSarifWithVerboseFlag)
	ctx.Step(`^I run devpipe sarif with nonexistent file$`, c.iRunDevpipeSarifWithNonexistentFile)
	ctx.Step(`^the output should display the findings$`, c.theOutputShouldDisplayTheFindings)
	ctx.Step(`^the output should show grouped summary$`, c.theOutputShouldShowGroupedSummary)
	ctx.Step(`^the output should show detailed metadata$`, c.theOutputShouldShowDetailedMetadata)
	ctx.Step(`^the output should indicate SARIF file not found$`, c.theOutputShouldIndicateSARIFFileNotFound)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		shared.cleanup()
		return ctx, nil
	})
}
