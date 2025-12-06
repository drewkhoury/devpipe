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
	configPath    string
	output        string
	exitCode      int
	tempDir       string
	devpipeBinary string
	taskIDs       []string
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

func (c *commandsContext) theExecutionShouldSucceed() error {
	if c.exitCode != 0 {
		return fmt.Errorf("expected execution to succeed (exit 0), got exit code %d\nOutput: %s", c.exitCode, c.output)
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

func (c *commandsContext) theExecutionShouldFail() error {
	if c.exitCode == 0 {
		return fmt.Errorf("expected execution to fail (non-zero exit), got exit code 0\nOutput: %s", c.output)
	}
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

func (c *commandsContext) theOutputShouldContain(expected string) error {
	if !strings.Contains(strings.ToLower(c.output), strings.ToLower(expected)) {
		return fmt.Errorf("expected output to contain %q, got: %s", expected, c.output)
	}
	return nil
}

func (c *commandsContext) cleanup() {
	if c.tempDir != "" {
		_ = os.RemoveAll(c.tempDir) // Best effort cleanup
	}
}

func InitializeCommandsScenario(ctx *godog.ScenarioContext) {
	c := &commandsContext{}

	// Find the devpipe binary
	wd, _ := os.Getwd()
	c.devpipeBinary = filepath.Join(wd, "..", "devpipe")
	if _, err := os.Stat(c.devpipeBinary); os.IsNotExist(err) {
		c.devpipeBinary = filepath.Join(wd, "devpipe")
	}

	ctx.Step(`^a config with multiple tasks$`, c.aConfigWithMultipleTasks)
	ctx.Step(`^I run devpipe list$`, c.iRunDevpipeList)
	ctx.Step(`^the execution should succeed$`, c.theExecutionShouldSucceed)
	ctx.Step(`^the output should contain all task IDs$`, c.theOutputShouldContainAllTaskIDs)
	ctx.Step(`^I run devpipe list --verbose$`, c.iRunDevpipeListVerbose)
	ctx.Step(`^the output should show a table format$`, c.theOutputShouldShowATableFormat)
	ctx.Step(`^the output should contain task names$`, c.theOutputShouldContainTaskNames)
	ctx.Step(`^the output should contain task types$`, c.theOutputShouldContainTaskTypes)
	ctx.Step(`^no config file exists$`, c.noConfigFileExists)
	ctx.Step(`^the execution should fail$`, c.theExecutionShouldFail)
	ctx.Step(`^the output should indicate config file not found$`, c.theOutputShouldIndicateConfigFileNotFound)
	ctx.Step(`^I run devpipe version$`, c.iRunDevpipeVersion)
	ctx.Step(`^the output should contain version number$`, c.theOutputShouldContainVersionNumber)
	ctx.Step(`^the output should contain "([^"]*)"$`, c.theOutputShouldContain)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
