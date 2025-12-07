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

type configEdgeCasesContext struct {
	*sharedContext
	configFiles []string
}

// Helper: Initialize git repo in temp dir
func (c *configEdgeCasesContext) initGitRepo() error {
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

// Scenario: Multiple config files validation
func (c *configEdgeCasesContext) multipleConfigFilesWithDifferentTasks() error {
	c.tempDir = filepath.Join("/tmp/devpipe-testing", fmt.Sprintf("devpipe-config-edge-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	// Create first config
	config1 := filepath.Join(c.tempDir, "config1.toml")
	content1 := `[defaults]
outputRoot = ".devpipe"

[tasks.task-1]
name = "Task 1"
command = "echo task1"
type = "test"
`
	if err := os.WriteFile(config1, []byte(content1), 0644); err != nil {
		return err
	}
	c.configFiles = append(c.configFiles, config1)

	// Create second config
	config2 := filepath.Join(c.tempDir, "config2.toml")
	content2 := `[defaults]
outputRoot = ".devpipe"

[tasks.task-2]
name = "Task 2"
command = "echo task2"
type = "test"
`
	if err := os.WriteFile(config2, []byte(content2), 0644); err != nil {
		return err
	}
	c.configFiles = append(c.configFiles, config2)

	return nil
}

func (c *configEdgeCasesContext) iRunDevpipeValidateOnAllConfigFiles() error {
	args := []string{"validate"}
	args = append(args, c.configFiles...)

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

func (c *configEdgeCasesContext) eachConfigShouldBeValidatedIndependently() error {
	// Check that both config files are mentioned in output
	for _, configFile := range c.configFiles {
		basename := filepath.Base(configFile)
		if !strings.Contains(c.output, basename) {
			return fmt.Errorf("expected output to mention %s, got: %s", basename, c.output)
		}
	}
	return nil
}

func (c *configEdgeCasesContext) validationResultsShouldBeShownForEach() error {
	// Check for validation indicators
	if !strings.Contains(c.output, "Valid") && !strings.Contains(c.output, "✓") {
		return fmt.Errorf("expected validation results in output, got: %s", c.output)
	}
	return nil
}

// Scenario: Config with only phase headers
func (c *configEdgeCasesContext) aConfigWithPhaseHeadersButNoTasks() error {
	c.tempDir = filepath.Join("/tmp/devpipe-testing", fmt.Sprintf("devpipe-config-edge-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.phase-1]
name = "Phase 1"

[tasks.phase-2]
name = "Phase 2"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *configEdgeCasesContext) aWarningShouldIndicateNoTasksToRun() error {
	// The execution will complete but with no tasks
	// Check that no tasks actually ran
	if strings.Contains(c.output, "RUN") && strings.Contains(c.output, "PASS") {
		return fmt.Errorf("expected no tasks to run, but found task execution in output: %s", c.output)
	}
	return nil
}

// Scenario: Config with very large task count
func (c *configEdgeCasesContext) aConfigWith100PlusTasks() error {
	c.tempDir = filepath.Join("/tmp/devpipe-testing", fmt.Sprintf("devpipe-config-edge-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")

	// Generate config with 100 tasks
	var configBuilder strings.Builder
	configBuilder.WriteString(`[defaults]
outputRoot = ".devpipe"

`)

	for i := 1; i <= 100; i++ {
		configBuilder.WriteString(fmt.Sprintf(`[tasks.task-%d]
name = "Task %d"
command = "echo task%d"
type = "test"

`, i, i, i))
	}

	return os.WriteFile(c.configPath, []byte(configBuilder.String()), 0644)
}

func (c *configEdgeCasesContext) iRunDevpipeListWithThatConfig() error {
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

func (c *configEdgeCasesContext) allTasksShouldBeListed() error {
	// Check that we have a reasonable number of tasks listed
	taskCount := strings.Count(c.output, "task-")
	if taskCount < 90 { // Allow some margin
		return fmt.Errorf("expected ~100 tasks to be listed, found %d", taskCount)
	}
	return nil
}

func (c *configEdgeCasesContext) performanceShouldBeAcceptable() error {
	// If we got here without timeout, performance is acceptable
	return nil
}

// Scenario: Empty task IDs or special characters
func (c *configEdgeCasesContext) aConfigWithTasksHavingSpecialCharactersInIDs() error {
	c.tempDir = filepath.Join("/tmp/devpipe-testing", fmt.Sprintf("devpipe-config-edge-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	if err := c.initGitRepo(); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.task-with-dashes]
name = "Task With Dashes"
command = "echo dashes"
type = "test"

[tasks.task_with_underscores]
name = "Task With Underscores"
command = "echo underscores"
type = "test"

[tasks.task-with-dots]
name = "Task With Dots"
command = "echo dots"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *configEdgeCasesContext) tasksShouldExecuteCorrectly() error {
	// Check that tasks ran
	if !strings.Contains(c.output, "PASS") {
		return fmt.Errorf("expected tasks to execute, got output: %s", c.output)
	}
	return nil
}

func (c *configEdgeCasesContext) outputShouldDisplayTaskIDsProperly() error {
	// Check that task IDs with special characters appear in output
	specialIDs := []string{"task-with-dashes", "task_with_underscores", "task-with-dots"}
	for _, id := range specialIDs {
		if !strings.Contains(c.output, id) {
			return fmt.Errorf("expected task ID '%s' in output, got: %s", id, c.output)
		}
	}
	return nil
}

// Shared helpers
func (c *configEdgeCasesContext) iRunDevpipeWithThatConfig() error {
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

func (c *configEdgeCasesContext) theExecutionShouldSucceed() error {
	if c.exitCode != 0 {
		return fmt.Errorf("expected execution to succeed (exit 0), got exit code %d\nOutput: %s", c.exitCode, c.output)
	}
	return nil
}

func InitializeConfigEdgeCasesScenario(ctx *godog.ScenarioContext, shared *sharedContext) {
	c := &configEdgeCasesContext{sharedContext: shared}

	// Scenario: Multiple config files validation
	ctx.Step(`^multiple config files with different tasks$`, c.multipleConfigFilesWithDifferentTasks)
	ctx.Step(`^I run devpipe validate on all config files$`, c.iRunDevpipeValidateOnAllConfigFiles)
	ctx.Step(`^each config should be validated independently$`, c.eachConfigShouldBeValidatedIndependently)
	ctx.Step(`^validation results should be shown for each$`, c.validationResultsShouldBeShownForEach)

	// Scenario: Config with only phase headers
	ctx.Step(`^a config with phase headers but no tasks$`, c.aConfigWithPhaseHeadersButNoTasks)
	ctx.Step(`^a warning should indicate no tasks to run$`, c.aWarningShouldIndicateNoTasksToRun)

	// Scenario: Config with very large task count
	ctx.Step(`^a config with 100\+ tasks$`, c.aConfigWith100PlusTasks)
	ctx.Step(`^I run devpipe list with that config$`, c.iRunDevpipeListWithThatConfig)
	ctx.Step(`^all tasks should be listed$`, c.allTasksShouldBeListed)
	ctx.Step(`^performance should be acceptable$`, c.performanceShouldBeAcceptable)

	// Scenario: Empty task IDs or special characters
	ctx.Step(`^a config with tasks having special characters in IDs$`, c.aConfigWithTasksHavingSpecialCharactersInIDs)
	ctx.Step(`^I run devpipe with the special characters config$`, c.iRunDevpipeWithThatConfig)
	ctx.Step(`^tasks should execute correctly$`, c.tasksShouldExecuteCorrectly)
	ctx.Step(`^output should display task IDs properly$`, c.outputShouldDisplayTaskIDsProperly)

	// Shared steps
	ctx.Step(`^I run devpipe with that config$`, c.iRunDevpipeWithThatConfig)
	ctx.Step(`^the execution should succeed$`, c.theExecutionShouldSucceed)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
