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

type taskExecutionContext struct {
	*sharedContext
}

func (c *taskExecutionContext) aConfigFileWithASimpleEchoTask() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-test-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "test-config.toml")
	c.outputRoot = filepath.Join(c.tempDir, ".devpipe")

	// Initialize git repo so devpipe doesn't error, and create a file so tasks run
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()                                                                 // Test setup
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644) // Test setup

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.echo-test]
name = "Echo Test"
command = "echo 'Hello from devpipe'"
type = "test"
`, c.outputRoot)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *taskExecutionContext) iRunDevpipeWithThatConfig() error {
	return c.runDevpipe()
}

func (c *taskExecutionContext) aConfigFileWithAFailingTask() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-test-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "test-config.toml")
	c.outputRoot = filepath.Join(c.tempDir, ".devpipe")

	// Initialize git repo and create a file
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()                                                                 // Test setup
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644) // Test setup

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.failing-test]
name = "Failing Test"
command = "exit 1"
type = "test"
`, c.outputRoot)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *taskExecutionContext) aConfigFileWithMultipleTasks() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-test-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "test-config.toml")
	c.outputRoot = filepath.Join(c.tempDir, ".devpipe")
	c.taskName = "task-one"

	// Initialize git repo and create a file
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()                                                                 // Test setup
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644) // Test setup

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.task-one]
name = "Task One"
command = "echo 'Task one running'"
type = "test"

[tasks.task-two]
name = "Task Two"
command = "echo 'Task two running'"
type = "test"

[tasks.task-three]
name = "Task Three"
command = "echo 'Task three running'"
type = "test"
`, c.outputRoot)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *taskExecutionContext) iRunDevpipeWithOnlyFlagForOneTask() error {
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "-only", c.taskName)
	cmd.Dir = c.tempDir
	output, err := cmd.CombinedOutput()
	c.output = string(output)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			c.exitCode = exitErr.ExitCode()
		} else {
			return err
		}
	} else {
		c.exitCode = 0
	}

	return nil
}

func (c *taskExecutionContext) onlyTheSpecifiedTaskShouldRun() error {
	if !strings.Contains(c.output, "task-one") {
		return fmt.Errorf("expected output to contain the specified task 'task-one', got: %s", c.output)
	}
	return nil
}

func (c *taskExecutionContext) otherTasksShouldNotAppearInOutput() error {
	if strings.Contains(c.output, "task-two") || strings.Contains(c.output, "task-three") {
		return fmt.Errorf("expected other tasks not to run, but found them in output: %s", c.output)
	}
	return nil
}

func (c *taskExecutionContext) aConfigFileWithASimpleTask() error {
	return c.aConfigFileWithASimpleEchoTask()
}

func (c *taskExecutionContext) logFilesShouldBeCreatedInTheOutputDirectory() error {
	// Check that the output directory exists
	if _, err := os.Stat(c.outputRoot); os.IsNotExist(err) {
		return fmt.Errorf("output directory %s does not exist", c.outputRoot)
	}

	// Check for runs directory
	runsDir := filepath.Join(c.outputRoot, "runs")
	if _, err := os.Stat(runsDir); os.IsNotExist(err) {
		return fmt.Errorf("runs directory %s does not exist", runsDir)
	}

	// Check that at least one run directory exists
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		return fmt.Errorf("failed to read runs directory: %v", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no run directories found in %s", runsDir)
	}

	// Check for logs directory in the most recent run
	runDir := filepath.Join(runsDir, entries[0].Name())
	logsDir := filepath.Join(runDir, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return fmt.Errorf("logs directory %s does not exist", logsDir)
	}

	return nil
}

func InitializeTaskExecutionScenario(ctx *godog.ScenarioContext, shared *sharedContext) {
	c := &taskExecutionContext{sharedContext: shared}

	ctx.Step(`^a config file with a simple echo task$`, c.aConfigFileWithASimpleEchoTask)
	ctx.Step(`^I run devpipe with that config$`, c.iRunDevpipeWithThatConfig)
	ctx.Step(`^the execution should succeed$`, c.theExecutionShouldSucceed)
	ctx.Step(`^the output should contain "([^"]*)"$`, c.theOutputShouldContain)
	ctx.Step(`^a config file with a failing task$`, c.aConfigFileWithAFailingTask)
	ctx.Step(`^the execution should fail$`, c.theExecutionShouldFail)
	ctx.Step(`^a config file with multiple tasks$`, c.aConfigFileWithMultipleTasks)
	ctx.Step(`^I run devpipe with --only flag for one task$`, c.iRunDevpipeWithOnlyFlagForOneTask)
	ctx.Step(`^only the specified task should run$`, c.onlyTheSpecifiedTaskShouldRun)
	ctx.Step(`^other tasks should not appear in output$`, c.otherTasksShouldNotAppearInOutput)
	ctx.Step(`^a config file with a simple task$`, c.aConfigFileWithASimpleTask)
	ctx.Step(`^log files should be created in the output directory$`, c.logFilesShouldBeCreatedInTheOutputDirectory)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
