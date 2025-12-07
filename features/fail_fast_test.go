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

type failFastContext struct {
	*sharedContext
}

func (c *failFastContext) aConfigWithThreeTasksWhereTheSecondFails() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-failfast-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	// Use different phases to ensure sequential execution
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.phase-1]
name = "Phase 1"

[tasks.task-1]
name = "Task 1"
command = "echo task1"
type = "test"

[tasks.phase-2]
name = "Phase 2"

[tasks.task-2]
name = "Task 2"
command = "exit 1"
type = "test"

[tasks.phase-3]
name = "Phase 3"

[tasks.task-3]
name = "Task 3"
command = "echo task3"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *failFastContext) aConfigWithThreeTasksWhereSecondAndThirdFail() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-failfast-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	// Use different phases to ensure sequential execution
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.phase-1]
name = "Phase 1"

[tasks.task-1]
name = "Task 1"
command = "echo task1"
type = "test"

[tasks.phase-2]
name = "Phase 2"

[tasks.task-2]
name = "Task 2"
command = "exit 1"
type = "test"

[tasks.phase-3]
name = "Phase 3"

[tasks.task-3]
name = "Task 3"
command = "exit 1"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *failFastContext) aConfigWithThreePassingTasks() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-failfast-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.task-1]
name = "Task 1"
command = "echo task1"
type = "test"

[tasks.task-2]
name = "Task 2"
command = "echo task2"
type = "test"

[tasks.task-3]
name = "Task 3"
command = "echo task3"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *failFastContext) iRunDevpipeWithFailFast() error {
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "-fail-fast")
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

func (c *failFastContext) iRunDevpipeWithoutFailFast() error {
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

func (c *failFastContext) theExecutionShouldFail() error {
	if c.exitCode == 0 {
		return fmt.Errorf("expected execution to fail (non-zero exit), got exit code 0\nOutput: %s", c.output)
	}
	return nil
}

func (c *failFastContext) theExecutionShouldSucceed() error {
	if c.exitCode != 0 {
		return fmt.Errorf("expected execution to succeed (exit 0), got exit code %d\nOutput: %s", c.exitCode, c.output)
	}
	return nil
}

func (c *failFastContext) theFirstTaskShouldHaveRun() error {
	if !strings.Contains(c.output, "task-1") {
		return fmt.Errorf("expected task-1 to have run, got output: %s", c.output)
	}
	return nil
}

func (c *failFastContext) theSecondTaskShouldHaveFailed() error {
	if !strings.Contains(c.output, "task-2") {
		return fmt.Errorf("expected task-2 to have run, got output: %s", c.output)
	}
	// Check for failure indicators
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "fail") && !strings.Contains(lowerOutput, "error") {
		return fmt.Errorf("expected task-2 to show failure, got output: %s", c.output)
	}
	return nil
}

func (c *failFastContext) theThirdTaskShouldNotHaveRun() error {
	if strings.Contains(c.output, "task-3") {
		return fmt.Errorf("expected task-3 NOT to have run, but found it in output: %s", c.output)
	}
	return nil
}

func (c *failFastContext) onlyTheFirstTwoTasksShouldHaveRun() error {
	if !strings.Contains(c.output, "task-1") {
		return fmt.Errorf("expected task-1 to have run, got output: %s", c.output)
	}
	if !strings.Contains(c.output, "task-2") {
		return fmt.Errorf("expected task-2 to have run, got output: %s", c.output)
	}
	return nil
}

func (c *failFastContext) allThreeTasksShouldHaveRun() error {
	tasks := []string{"task-1", "task-2", "task-3"}
	for _, task := range tasks {
		if !strings.Contains(c.output, task) {
			return fmt.Errorf("expected %s to have run, got output: %s", task, c.output)
		}
	}
	return nil
}

func InitializeFailFastScenario(ctx *godog.ScenarioContext, shared *sharedContext) {
	c := &failFastContext{sharedContext: shared}

	ctx.Step(`^a config with three tasks where the second fails$`, c.aConfigWithThreeTasksWhereTheSecondFails)
	ctx.Step(`^a config with three tasks where second and third fail$`, c.aConfigWithThreeTasksWhereSecondAndThirdFail)
	ctx.Step(`^a config with three passing tasks$`, c.aConfigWithThreePassingTasks)

	ctx.Step(`^I run devpipe with --fail-fast$`, c.iRunDevpipeWithFailFast)
	ctx.Step(`^I run devpipe without fail-fast$`, c.iRunDevpipeWithoutFailFast)

	ctx.Step(`^the execution should fail$`, c.theExecutionShouldFail)
	ctx.Step(`^the execution should succeed$`, c.theExecutionShouldSucceed)

	ctx.Step(`^the first task should have run$`, c.theFirstTaskShouldHaveRun)
	ctx.Step(`^the second task should have failed$`, c.theSecondTaskShouldHaveFailed)
	ctx.Step(`^the third task should not have run$`, c.theThirdTaskShouldNotHaveRun)
	ctx.Step(`^only the first two tasks should have run$`, c.onlyTheFirstTwoTasksShouldHaveRun)
	ctx.Step(`^all three tasks should have run$`, c.allThreeTasksShouldHaveRun)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
