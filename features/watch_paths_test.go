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

type watchPathsContext struct {
	*sharedContext
	tempDir    string
	configPath string
	output     string
	exitCode   int
	tasks      map[string]taskConfig
}

type taskConfig struct {
	id         string
	workdir    string
	watchPaths []string
}

// Helper: Initialize git repo in temp dir
func (c *watchPathsContext) initGitRepo() error {
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

	// Create initial commit
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

// Scenario: Task with watchPaths skipped when no files changed
func (c *watchPathsContext) aGitRepoWithNoChanges() error {
	c.tempDir = filepath.Join("/tmp/devpipe-testing", fmt.Sprintf("devpipe-watch-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	return c.initGitRepo()
}

func (c *watchPathsContext) aTaskWithWatchPathsFor(pattern string) error {
	if c.tasks == nil {
		c.tasks = make(map[string]taskConfig)
	}
	c.tasks["test-task"] = taskConfig{
		id:         "test-task",
		workdir:    ".",
		watchPaths: []string{pattern},
	}
	return nil
}

func (c *watchPathsContext) iRunDevpipe() error {
	// Create config
	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

`
	for _, task := range c.tasks {
		config += fmt.Sprintf(`[tasks.%s]
name = "%s"
command = "echo 'running %s'"
type = "test"
`, task.id, task.id, task.id)

		if task.workdir != "" && task.workdir != "." {
			config += fmt.Sprintf(`workdir = "%s"
`, task.workdir)
		}

		if len(task.watchPaths) > 0 {
			config += `watchPaths = [`
			for i, p := range task.watchPaths {
				if i > 0 {
					config += ", "
				}
				config += fmt.Sprintf(`"%s"`, p)
			}
			config += "]\n"
		}
		config += "\n"
	}

	if err := os.WriteFile(c.configPath, []byte(config), 0644); err != nil {
		return err
	}

	// Run devpipe
	cmd := exec.Command(c.devpipeBinary, "--config", c.configPath, "--verbose")
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

func (c *watchPathsContext) theTaskShouldBeSkipped() error {
	if strings.Contains(c.output, "RUN") && strings.Contains(c.output, "test-task") {
		return fmt.Errorf("expected task to be skipped, but it ran. Output: %s", c.output)
	}
	if !strings.Contains(c.output, "SKIP") {
		return fmt.Errorf("expected SKIP in output, got: %s", c.output)
	}
	return nil
}

func (c *watchPathsContext) theSkipReasonShouldMention(reason string) error {
	if !strings.Contains(c.output, reason) {
		return fmt.Errorf("expected skip reason to mention %q, got output: %s", reason, c.output)
	}
	return nil
}

// Scenario: Task with watchPaths runs when matching file changes
func (c *watchPathsContext) aGitRepoWithChangesTo(filename string) error {
	c.tempDir = filepath.Join("/tmp/devpipe-testing", fmt.Sprintf("devpipe-watch-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	if err := c.initGitRepo(); err != nil {
		return err
	}

	// Create the file in the appropriate directory
	dir := filepath.Dir(filename)
	if dir != "." {
		fullDir := filepath.Join(c.tempDir, dir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			return err
		}
	}

	fullPath := filepath.Join(c.tempDir, filename)
	if err := os.WriteFile(fullPath, []byte("changed"), 0644); err != nil {
		return err
	}

	// Stage the file so git detects it as changed
	cmd := exec.Command("git", "add", filename)
	cmd.Dir = c.tempDir
	return cmd.Run()
}

func (c *watchPathsContext) theTaskShouldRun() error {
	if !strings.Contains(c.output, "RUN") {
		return fmt.Errorf("expected task to run, got output: %s", c.output)
	}
	if !strings.Contains(c.output, "PASS") {
		return fmt.Errorf("expected task to pass, got output: %s", c.output)
	}
	return nil
}

// Scenario: Task without watchPaths always runs
func (c *watchPathsContext) aTaskWithoutWatchPaths() error {
	if c.tasks == nil {
		c.tasks = make(map[string]taskConfig)
	}
	c.tasks["test-task"] = taskConfig{
		id:         "test-task",
		workdir:    ".",
		watchPaths: nil,
	}
	return nil
}

// Scenario: Multiple watchPaths patterns
func (c *watchPathsContext) aTaskWithWatchPathsForAnd(pattern1, pattern2 string) error {
	if c.tasks == nil {
		c.tasks = make(map[string]taskConfig)
	}
	c.tasks["test-task"] = taskConfig{
		id:         "test-task",
		workdir:    ".",
		watchPaths: []string{pattern1, pattern2},
	}
	return nil
}

// Scenario: Ignore watch paths flag
func (c *watchPathsContext) iRunDevpipeWithIgnoreWatchPaths() error {
	// Create config
	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

`
	for _, task := range c.tasks {
		config += fmt.Sprintf(`[tasks.%s]
name = "%s"
command = "echo 'running %s'"
type = "test"
`, task.id, task.id, task.id)

		if len(task.watchPaths) > 0 {
			config += `watchPaths = [`
			for i, p := range task.watchPaths {
				if i > 0 {
					config += ", "
				}
				config += fmt.Sprintf(`"%s"`, p)
			}
			config += "]\n"
		}
		config += "\n"
	}

	if err := os.WriteFile(c.configPath, []byte(config), 0644); err != nil {
		return err
	}

	// Run devpipe with --ignore-watch-paths
	cmd := exec.Command(c.devpipeBinary, "--config", c.configPath, "--ignore-watch-paths")
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

// Scenario: WatchPaths relative to workdir
func (c *watchPathsContext) aTaskWithWorkdirAndWatchPathsFor(workdir, pattern string) error {
	if c.tasks == nil {
		c.tasks = make(map[string]taskConfig)
	}
	c.tasks["test-task"] = taskConfig{
		id:         "test-task",
		workdir:    workdir,
		watchPaths: []string{pattern},
	}
	return nil
}

// Scenario: Multiple tasks with different watchPaths
func (c *watchPathsContext) aTaskWithWatchPathsForTask(taskID, pattern string) error {
	if c.tasks == nil {
		c.tasks = make(map[string]taskConfig)
	}
	c.tasks[taskID] = taskConfig{
		id:         taskID,
		workdir:    ".",
		watchPaths: []string{pattern},
	}
	return nil
}

func (c *watchPathsContext) taskShouldRun(taskID string) error {
	if !strings.Contains(c.output, taskID) {
		return fmt.Errorf("expected %s in output, got: %s", taskID, c.output)
	}
	if !strings.Contains(c.output, "PASS") {
		return fmt.Errorf("expected %s to pass, got output: %s", taskID, c.output)
	}
	return nil
}

func (c *watchPathsContext) taskShouldBeSkipped(taskID string) error {
	// Check that task is mentioned with SKIP, not RUN
	lines := strings.Split(c.output, "\n")
	for _, line := range lines {
		if strings.Contains(line, taskID) {
			if strings.Contains(line, "RUN") && !strings.Contains(line, "SKIP") {
				return fmt.Errorf("expected %s to be skipped, but it ran. Output: %s", taskID, c.output)
			}
			if strings.Contains(line, "SKIP") {
				return nil // Found it skipped, good!
			}
		}
	}
	// If we get here, task wasn't mentioned at all, which is also fine for skipped tasks
	return nil
}

func (c *watchPathsContext) cleanup() {
	if c.tempDir != "" {
		_ = os.RemoveAll(c.tempDir)
	}
}

func InitializeWatchPathsScenario(ctx *godog.ScenarioContext, shared *sharedContext) {
	c := &watchPathsContext{sharedContext: shared}

	// Scenario: Task with watchPaths skipped when no files changed
	ctx.Step(`^a git repo with no changes$`, c.aGitRepoWithNoChanges)
	ctx.Step(`^a task with watchPaths for "([^"]*)"$`, c.aTaskWithWatchPathsFor)
	ctx.Step(`^I run devpipe$`, c.iRunDevpipe)
	ctx.Step(`^the task should be skipped$`, c.theTaskShouldBeSkipped)
	ctx.Step(`^the skip reason should mention "([^"]*)"$`, c.theSkipReasonShouldMention)

	// Scenario: Task with watchPaths runs when matching file changes
	ctx.Step(`^a git repo with changes to "([^"]*)"$`, c.aGitRepoWithChangesTo)
	ctx.Step(`^the task should run$`, c.theTaskShouldRun)

	// Scenario: Task without watchPaths always runs
	ctx.Step(`^a task without watchPaths$`, c.aTaskWithoutWatchPaths)

	// Scenario: Multiple watchPaths patterns
	ctx.Step(`^a task with watchPaths for "([^"]*)" and "([^"]*)"$`, c.aTaskWithWatchPathsForAnd)

	// Scenario: Ignore watch paths flag
	ctx.Step(`^I run devpipe with --ignore-watch-paths$`, c.iRunDevpipeWithIgnoreWatchPaths)

	// Scenario: WatchPaths relative to workdir
	ctx.Step(`^a task with workdir "([^"]*)" and watchPaths for "([^"]*)"$`, c.aTaskWithWorkdirAndWatchPathsFor)

	// Scenario: Multiple tasks with different watchPaths
	ctx.Step(`^a task "([^"]*)" with watchPaths for "([^"]*)"$`, c.aTaskWithWatchPathsForTask)
	ctx.Step(`^task "([^"]*)" should run$`, c.taskShouldRun)
	ctx.Step(`^task "([^"]*)" should be skipped$`, c.taskShouldBeSkipped)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
