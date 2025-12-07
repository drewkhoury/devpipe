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
	cmd := exec.Command(c.devpipeBinary, "list", "--config", c.configPath)
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
	cmd := exec.Command(c.devpipeBinary, "list", "--verbose", "--config", c.configPath)
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
	upperOutput := strings.ToUpper(c.output)
	if !strings.Contains(upperOutput, "NAME") && !strings.Contains(upperOutput, "TYPE") {
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

func (c *commandsContext) aCustomConfigFileWithSpecificTasks() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-commands-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "custom.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.custom-task-1]
name = "Custom Task One"
command = "echo custom1"
type = "test-unit"

[tasks.custom-task-2]
name = "Custom Task Two"
command = "echo custom2"
type = "build"
`
	c.taskIDs = []string{"custom-task-1", "custom-task-2"}
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *commandsContext) iRunDevpipeListWithConfigFlag() error {
	cmd := exec.Command(c.devpipeBinary, "list", "--config", c.configPath)
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

func (c *commandsContext) theOutputShouldContainTasksFromCustomConfig() error {
	customTasks := []string{"custom-task-1", "custom-task-2"}
	for _, task := range customTasks {
		if !strings.Contains(c.output, task) {
			return fmt.Errorf("expected output to contain custom task %q, got: %s", task, c.output)
		}
	}
	return nil
}

func (c *commandsContext) iRunDevpipeListWithNonexistentConfig() error {
	nonexistentPath := filepath.Join(c.tempDir, "nonexistent.toml")
	cmd := exec.Command(c.devpipeBinary, "list", "--config", nonexistentPath)
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

func (c *commandsContext) aValidDefaultConfigFile() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validate-default-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	// Create config.toml in temp directory
	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[tasks.test]
name = "Test Task"
command = "echo test"
type = "test-unit"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *commandsContext) iRunDevpipeValidateWithoutArguments() error {
	cmd := exec.Command(c.devpipeBinary, "validate")
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

// Generate-reports command steps
func (c *commandsContext) aConfigWithExistingRunReports() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-generate-reports-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	// Create config.toml
	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.test]
name = "Test Task"
command = "echo test"
type = "test-unit"
`
	if err := os.WriteFile(c.configPath, []byte(config), 0644); err != nil {
		return err
	}

	// Create .devpipe/runs directory with mock run directories
	runsDir := filepath.Join(c.tempDir, ".devpipe", "runs")
	if err := os.MkdirAll(runsDir, 0755); err != nil {
		return err
	}

	// Create 2 mock run directories
	for i := 1; i <= 2; i++ {
		runDir := filepath.Join(runsDir, fmt.Sprintf("2025-12-07T00-00-%02d_test", i))
		if err := os.MkdirAll(runDir, 0755); err != nil {
			return err
		}
		// Create a minimal log file
		logFile := filepath.Join(runDir, "logs", "test.log")
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(logFile, []byte("test log"), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (c *commandsContext) aValidConfigButNoRunsDirectory() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-generate-reports-noruns-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	// Create config.toml
	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.test]
name = "Test Task"
command = "echo test"
type = "test-unit"
`
	if err := os.WriteFile(c.configPath, []byte(config), 0644); err != nil {
		return err
	}

	// Create .devpipe directory but NOT the runs subdirectory
	devpipeDir := filepath.Join(c.tempDir, ".devpipe")
	if err := os.MkdirAll(devpipeDir, 0755); err != nil {
		return err
	}

	return nil
}

func (c *commandsContext) iRunDevpipeGenerateReports() error {
	cmd := exec.Command(c.devpipeBinary, "generate-reports")
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

func (c *commandsContext) theOutputShouldIndicateReportsWereRegenerated() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "regenerat") {
		return fmt.Errorf("expected output to indicate reports were regenerated, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) theOutputShouldShowTheNumberOfRunsProcessed() error {
	// Should contain a number followed by "reports" or "runs"
	if !strings.Contains(c.output, "2 reports") && !strings.Contains(c.output, "Regenerated 2") {
		return fmt.Errorf("expected output to show number of runs processed, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) theOutputShouldIndicateConfigError() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "config") && !strings.Contains(lowerOutput, "error") {
		return fmt.Errorf("expected output to indicate config error, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) theOutputShouldIndicateRunsDirectoryError() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "runs") || !strings.Contains(lowerOutput, "error") {
		return fmt.Errorf("expected output to indicate runs directory error, got: %s", c.output)
	}
	return nil
}

// SARIF directory scan steps
func (c *commandsContext) aDirectoryWithMultipleSARIFFiles() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-sarif-dir-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	// Create 2 SARIF files with different findings
	sarif1 := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [{
    "tool": {"driver": {"name": "Scanner1", "version": "1.0.0"}},
    "results": [{
      "ruleId": "RULE001",
      "level": "error",
      "message": {"text": "Finding from file 1"},
      "locations": [{"physicalLocation": {"artifactLocation": {"uri": "file1.go"}}}]
    }]
  }]
}`
	sarif2 := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [{
    "tool": {"driver": {"name": "Scanner2", "version": "1.0.0"}},
    "results": [{
      "ruleId": "RULE002",
      "level": "warning",
      "message": {"text": "Finding from file 2"},
      "locations": [{"physicalLocation": {"artifactLocation": {"uri": "file2.go"}}}]
    }]
  }]
}`

	if err := os.WriteFile(filepath.Join(c.tempDir, "results1.sarif"), []byte(sarif1), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(c.tempDir, "results2.sarif"), []byte(sarif2), 0644); err != nil {
		return err
	}

	return nil
}

func (c *commandsContext) anEmptyDirectory() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-sarif-empty-%d", os.Getpid()))
	return os.MkdirAll(c.tempDir, 0755)
}

func (c *commandsContext) aDirectoryWithNonSARIFFiles() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-sarif-nofiles-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	// Create some non-SARIF files
	if err := os.WriteFile(filepath.Join(c.tempDir, "test.txt"), []byte("not sarif"), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(c.tempDir, "data.json"), []byte("{}"), 0644); err != nil {
		return err
	}

	return nil
}

func (c *commandsContext) iRunDevpipeSarifWithDirectoryFlag() error {
	cmd := exec.Command(c.devpipeBinary, "sarif", "-d", c.tempDir)
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

func (c *commandsContext) theOutputShouldDisplayFindingsFromAllFiles() error {
	// Should contain findings from both files
	if !strings.Contains(c.output, "RULE001") || !strings.Contains(c.output, "RULE002") {
		return fmt.Errorf("expected output to display findings from all files, got: %s", c.output)
	}
	return nil
}

func (c *commandsContext) theOutputShouldIndicateNoSARIFFilesFound() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "no sarif files") && !strings.Contains(lowerOutput, "no") {
		return fmt.Errorf("expected output to indicate no SARIF files found, got: %s", c.output)
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
	ctx.Step(`^I run devpipe list with verbose flag$`, c.iRunDevpipeListVerbose)
	ctx.Step(`^I run devpipe list --verbose$`, c.iRunDevpipeListVerbose)
	ctx.Step(`^the output should show a table format$`, c.theOutputShouldShowATableFormat)
	ctx.Step(`^the output should contain task names$`, c.theOutputShouldContainTaskNames)
	ctx.Step(`^the output should contain task types$`, c.theOutputShouldContainTaskTypes)
	ctx.Step(`^a custom config file with specific tasks$`, c.aCustomConfigFileWithSpecificTasks)
	ctx.Step(`^I run devpipe list with config flag$`, c.iRunDevpipeListWithConfigFlag)
	ctx.Step(`^the output should contain tasks from custom config$`, c.theOutputShouldContainTasksFromCustomConfig)
	ctx.Step(`^I run devpipe list with nonexistent config$`, c.iRunDevpipeListWithNonexistentConfig)
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
	ctx.Step(`^a valid default config file$`, c.aValidDefaultConfigFile)
	ctx.Step(`^I run devpipe validate without arguments$`, c.iRunDevpipeValidateWithoutArguments)

	// Generate-reports command steps
	ctx.Step(`^a config with existing run reports$`, c.aConfigWithExistingRunReports)
	ctx.Step(`^a valid config but no runs directory$`, c.aValidConfigButNoRunsDirectory)
	ctx.Step(`^I run devpipe generate-reports$`, c.iRunDevpipeGenerateReports)
	ctx.Step(`^the output should indicate reports were regenerated$`, c.theOutputShouldIndicateReportsWereRegenerated)
	ctx.Step(`^the output should show the number of runs processed$`, c.theOutputShouldShowTheNumberOfRunsProcessed)
	ctx.Step(`^the output should indicate config error$`, c.theOutputShouldIndicateConfigError)
	ctx.Step(`^the output should indicate runs directory error$`, c.theOutputShouldIndicateRunsDirectoryError)

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

	// SARIF directory scan steps
	ctx.Step(`^a directory with multiple SARIF files$`, c.aDirectoryWithMultipleSARIFFiles)
	ctx.Step(`^an empty directory$`, c.anEmptyDirectory)
	ctx.Step(`^a directory with non-SARIF files$`, c.aDirectoryWithNonSARIFFiles)
	ctx.Step(`^I run devpipe sarif with directory flag$`, c.iRunDevpipeSarifWithDirectoryFlag)
	ctx.Step(`^the output should display findings from all files$`, c.theOutputShouldDisplayFindingsFromAllFiles)
	ctx.Step(`^the output should indicate no SARIF files found$`, c.theOutputShouldIndicateNoSARIFFilesFound)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		shared.cleanup()
		return ctx, nil
	})
}
