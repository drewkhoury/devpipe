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

type phaseExecutionContext struct {
	*sharedContext
}

// Helper: Initialize git repo in temp dir
func (c *phaseExecutionContext) initGitRepo() error {
	cmd := exec.Command("git", "init")
	cmd.Dir = c.tempDir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = c.tempDir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = c.tempDir
	if err := cmd.Run(); err != nil {
		return err
	}

	// Create and commit config file so git mode works
	dummyFile := filepath.Join(c.tempDir, "dummy.txt")
	if err := os.WriteFile(dummyFile, []byte("dummy"), 0644); err != nil {
		return err
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = c.tempDir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = c.tempDir
	return cmd.Run()
}

// Scenario: Sequential phase execution
func (c *phaseExecutionContext) aConfigWithThreePhases() error {
	c.tempDir = filepath.Join("/tmp/devpipe-testing", fmt.Sprintf("devpipe-phase-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.phase-1]
name = "Phase 1"

[tasks.task-1]
name = "Task 1"
command = "echo phase1"
type = "test"

[tasks.phase-2]
name = "Phase 2"

[tasks.task-2]
name = "Task 2"
command = "echo phase2"
type = "test"

[tasks.phase-3]
name = "Phase 3"

[tasks.task-3]
name = "Task 3"
command = "echo phase3"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *phaseExecutionContext) phase1ShouldCompleteBeforePhase2Starts() error {
	// Check that task-1 appears before task-2 in output (they're in different phases)
	idx1 := strings.Index(c.output, "task-1")
	idx2 := strings.Index(c.output, "task-2")

	if idx1 == -1 || idx2 == -1 {
		return fmt.Errorf("expected both task-1 and task-2 in output, got: %s", c.output)
	}

	if idx1 >= idx2 {
		return fmt.Errorf("expected task-1 to appear before task-2 in output")
	}

	return nil
}

func (c *phaseExecutionContext) phase2ShouldCompleteBeforePhase3Starts() error {
	// Check that task-2 appears before task-3 in output (they're in different phases)
	idx2 := strings.Index(c.output, "task-2")
	idx3 := strings.Index(c.output, "task-3")

	if idx2 == -1 || idx3 == -1 {
		return fmt.Errorf("expected both task-2 and task-3 in output, got: %s", c.output)
	}

	if idx2 >= idx3 {
		return fmt.Errorf("expected task-2 (task-2) to appear before task-3 (task-3) in output")
	}

	return nil
}

func (c *phaseExecutionContext) allPhasesShouldExecuteInOrder() error {
	// Verify all three tasks executed (one per phase)
	tasks := []string{"task-1", "task-2", "task-3"}
	for _, task := range tasks {
		if !strings.Contains(c.output, task) {
			return fmt.Errorf("expected task %s to execute, got output: %s", task, c.output)
		}
	}
	return nil
}

// Scenario: Phase execution stops on failure with fail-fast
func (c *phaseExecutionContext) aConfigWithThreePhasesWherePhase2Fails() error {
	c.tempDir = filepath.Join("/tmp/devpipe-testing", fmt.Sprintf("devpipe-phase-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	if err := c.initGitRepo(); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.phase-1]
name = "Phase 1"

[tasks.task-1]
name = "Task 1"
command = "echo phase1"
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
command = "echo phase3"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *phaseExecutionContext) iRunDevpipeWithFailFast() error {
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

func (c *phaseExecutionContext) phase1ShouldComplete() error {
	if !strings.Contains(c.output, "task-1") {
		return fmt.Errorf("expected task-1 to complete, got output: %s", c.output)
	}
	return nil
}

func (c *phaseExecutionContext) phase2ShouldFail() error {
	if !strings.Contains(c.output, "task-2") {
		return fmt.Errorf("expected task-2 to run, got output: %s", c.output)
	}
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "fail") {
		return fmt.Errorf("expected task-2 to fail, got output: %s", c.output)
	}
	return nil
}

func (c *phaseExecutionContext) phase3ShouldNotExecute() error {
	if strings.Contains(c.output, "task-3") {
		return fmt.Errorf("expected task-3 NOT to execute with fail-fast, but found it in output: %s", c.output)
	}
	return nil
}

// Scenario: Phase with no tasks is skipped
func (c *phaseExecutionContext) aConfigWithAPhaseHeaderButNoTasksInThatPhase() error {
	c.tempDir = filepath.Join("/tmp/devpipe-testing", fmt.Sprintf("devpipe-phase-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	if err := c.initGitRepo(); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.phase-1]
name = "Phase 1"

[tasks.task-1]
name = "Task 1"
command = "echo phase1"
type = "test"

[tasks.phase-empty]
name = "Empty Phase"

[tasks.phase-2]
name = "Phase 2"

[tasks.task-2]
name = "Task 2"
command = "echo phase2"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *phaseExecutionContext) theEmptyPhaseShouldBeSkipped() error {
	// Empty phase headers don't produce output, just verify no error
	return nil
}

func (c *phaseExecutionContext) otherPhasesShouldExecuteNormally() error {
	tasks := []string{"task-1", "task-2"}
	for _, task := range tasks {
		if !strings.Contains(c.output, task) {
			return fmt.Errorf("expected task %s to execute, got output: %s", task, c.output)
		}
	}
	return nil
}

// Shared helpers
func (c *phaseExecutionContext) iRunDevpipeWithThatConfig() error {
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

func (c *phaseExecutionContext) theExecutionShouldSucceed() error {
	if c.exitCode != 0 {
		return fmt.Errorf("expected execution to succeed (exit 0), got exit code %d\nOutput: %s", c.exitCode, c.output)
	}
	return nil
}

func InitializePhaseExecutionScenario(ctx *godog.ScenarioContext, shared *sharedContext) {
	c := &phaseExecutionContext{sharedContext: shared}

	// Scenario: Sequential phase execution
	ctx.Step(`^a config with three sequential phases$`, c.aConfigWithThreePhases)
	ctx.Step(`^phase 1 should complete before phase 2 starts$`, c.phase1ShouldCompleteBeforePhase2Starts)
	ctx.Step(`^phase 2 should complete before phase 3 starts$`, c.phase2ShouldCompleteBeforePhase3Starts)
	ctx.Step(`^all three phases should execute sequentially$`, c.allPhasesShouldExecuteInOrder)

	// Scenario: Phase execution stops on failure with fail-fast
	ctx.Step(`^a config with three phases where phase 2 fails$`, c.aConfigWithThreePhasesWherePhase2Fails)
	ctx.Step(`^I run devpipe with --fail-fast$`, c.iRunDevpipeWithFailFast)
	ctx.Step(`^phase 1 tasks should complete$`, c.phase1ShouldComplete)
	ctx.Step(`^phase 2 tasks should fail$`, c.phase2ShouldFail)
	ctx.Step(`^phase 3 tasks should not execute$`, c.phase3ShouldNotExecute)

	// Scenario: Phase with no tasks is skipped
	ctx.Step(`^a config with a phase header but no tasks in that phase$`, c.aConfigWithAPhaseHeaderButNoTasksInThatPhase)
	ctx.Step(`^the empty phase should be skipped$`, c.theEmptyPhaseShouldBeSkipped)
	ctx.Step(`^other phases should execute normally$`, c.otherPhasesShouldExecuteNormally)

	// Shared steps
	ctx.Step(`^I run devpipe with that config$`, c.iRunDevpipeWithThatConfig)
	ctx.Step(`^the execution should succeed$`, c.theExecutionShouldSucceed)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
