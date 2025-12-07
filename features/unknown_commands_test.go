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

type unknownCommandsContext struct {
	*sharedContext
}

// Scenario: Unknown subcommand shows error
func (c *unknownCommandsContext) iRunDevpipeWithSubcommand(subcommand string) error {
	cmd := exec.Command(c.devpipeBinary, subcommand)
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

func (c *unknownCommandsContext) theOutputShouldShowAvailableCommands() error {
	lowerOutput := strings.ToLower(c.output)
	requiredCommands := []string{"list", "validate", "sarif", "generate-reports", "version", "help"}

	for _, cmd := range requiredCommands {
		if !strings.Contains(lowerOutput, cmd) {
			return fmt.Errorf("expected output to show command '%s', got: %s", cmd, c.output)
		}
	}
	return nil
}

func (c *unknownCommandsContext) theOutputShouldIndicateUnknownCommand() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "unknown") && !strings.Contains(lowerOutput, "invalid") {
		return fmt.Errorf("expected output to indicate unknown command, got: %s", c.output)
	}
	return nil
}

// Scenario: No subcommand runs default pipeline
func (c *unknownCommandsContext) iRunDevpipeWithoutAnySubcommand() error {
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath)
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

func (c *unknownCommandsContext) theDefaultPipelineShouldRun() error {
	// Check for pipeline execution indicators
	if !strings.Contains(c.output, "RUN") && !strings.Contains(c.output, "PASS") {
		return fmt.Errorf("expected pipeline to run, got output: %s", c.output)
	}
	return nil
}

func (c *unknownCommandsContext) theTaskShouldExecute() error {
	// Check for task execution
	if !strings.Contains(c.output, "simple-task") && !strings.Contains(c.output, "PASS") {
		return fmt.Errorf("expected task to execute, got output: %s", c.output)
	}
	return nil
}

// Scenario: Typo in subcommand suggests correction
func (c *unknownCommandsContext) theOutputShouldSuggestAsCorrection(suggestion string) error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, strings.ToLower(suggestion)) {
		return fmt.Errorf("expected output to suggest '%s', got: %s", suggestion, c.output)
	}
	// Check for "did you mean" or similar phrasing
	if !strings.Contains(lowerOutput, "did you mean") && !strings.Contains(lowerOutput, "suggestion") && !strings.Contains(lowerOutput, "try") {
		return fmt.Errorf("expected output to suggest correction with helpful phrasing, got: %s", c.output)
	}
	return nil
}

// Helper: Create a simple config for testing
func (c *unknownCommandsContext) aConfigWithASimpleTask() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-unknown-cmd-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.simple-task]
name = "Simple Task"
command = "echo hello"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *unknownCommandsContext) theExecutionShouldFail() error {
	if c.exitCode == 0 {
		return fmt.Errorf("expected execution to fail (non-zero exit), got exit code 0")
	}
	return nil
}

func (c *unknownCommandsContext) theExecutionShouldSucceed() error {
	if c.exitCode != 0 {
		return fmt.Errorf("expected execution to succeed (exit 0), got exit code %d\nOutput: %s", c.exitCode, c.output)
	}
	return nil
}

func InitializeUnknownCommandsScenario(ctx *godog.ScenarioContext, shared *sharedContext) {
	c := &unknownCommandsContext{sharedContext: shared}

	// Scenario: Unknown subcommand shows error
	ctx.Step(`^I run devpipe with subcommand "([^"]*)"$`, c.iRunDevpipeWithSubcommand)
	ctx.Step(`^the output should show available commands$`, c.theOutputShouldShowAvailableCommands)
	ctx.Step(`^the output should indicate unknown command$`, c.theOutputShouldIndicateUnknownCommand)

	// Scenario: No subcommand runs default pipeline
	ctx.Step(`^a config file with a simple echo task$`, c.aConfigWithASimpleTask)
	ctx.Step(`^I run devpipe without any subcommand$`, c.iRunDevpipeWithoutAnySubcommand)
	ctx.Step(`^the default pipeline should run$`, c.theDefaultPipelineShouldRun)
	ctx.Step(`^the task should execute$`, c.theTaskShouldExecute)

	// Scenario: Typo in subcommand suggests correction
	ctx.Step(`^the output should suggest "([^"]*)" as correction$`, c.theOutputShouldSuggestAsCorrection)

	// Shared steps
	ctx.Step(`^the execution should fail$`, c.theExecutionShouldFail)
	ctx.Step(`^the execution should succeed$`, c.theExecutionShouldSucceed)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
