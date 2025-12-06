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

type errorScenariosContext struct {
	*sharedContext
}

func (c *errorScenariosContext) aConfigFileWithMalformedTOMLSyntax() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "malformed.toml")

	// Malformed TOML: unclosed string, invalid syntax
	malformedConfig := `[defaults]
outputRoot = ".devpipe

[tasks.test-task]
name = "Test Task
command = "echo test"
type = test"
`

	return os.WriteFile(c.configPath, []byte(malformedConfig), 0644)
}

func (c *errorScenariosContext) iRunDevpipeWithThatConfig() error {
	return c.runDevpipe()
}

func (c *errorScenariosContext) aConfigFileWithATaskMissingTheCommandField() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "missing-command.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.broken-task]
name = "Broken Task"
type = "test"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *errorScenariosContext) aConfigFileWithATaskMissingTheNameField() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "missing-name.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.broken-task]
command = "echo test"
type = "test"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *errorScenariosContext) aConfigWithATaskThatExitsWithCode(exitCode int) error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "exit-code.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.failing-task]
name = "Failing Task"
command = "exit %d"
type = "test"
`, c.tempDir, exitCode)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *errorScenariosContext) theOutputShouldIndicateTaskFailure() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "fail") && !strings.Contains(lowerOutput, "error") {
		return fmt.Errorf("expected output to indicate task failure, got: %s", c.output)
	}
	return nil
}

func (c *errorScenariosContext) theExitCodeShouldBeNonZero() error {
	if c.exitCode == 0 {
		return fmt.Errorf("expected non-zero exit code, got 0")
	}
	return nil
}

func (c *errorScenariosContext) aConfigWithATaskThatHasASecondTimeout(timeout int) error {
	c.taskTimeout = timeout
	return nil
}

func (c *errorScenariosContext) theTaskSleepsForSeconds(sleep int) error {
	c.taskSleep = sleep

	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "timeout.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.timeout-task]
name = "Timeout Task"
command = "sleep %d"
type = "test"
timeout = %d
`, c.tempDir, c.taskSleep, c.taskTimeout)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *errorScenariosContext) aConfigWithATaskUsingANonExistentBinary() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "missing-binary.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.missing-binary]
name = "Missing Binary"
command = "this-command-does-not-exist-12345"
type = "test"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *errorScenariosContext) theOutputShouldIndicateCommandNotFound() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "not found") && !strings.Contains(lowerOutput, "no such file") && !strings.Contains(lowerOutput, "executable") {
		return fmt.Errorf("expected output to indicate command not found, got: %s", c.output)
	}
	return nil
}

func (c *errorScenariosContext) aConfigWithATaskUsingAnInvalidCommandPath() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "invalid-path.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.invalid-path]
name = "Invalid Path"
command = "/nonexistent/path/to/command"
type = "test"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *errorScenariosContext) aConfigWithThreeTasksWhereTheSecondAndThirdFail() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "multiple-failures.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.first-task]
name = "First Task"
command = "echo 'First task success'"
type = "test"

[tasks.second-task]
name = "Second Task"
command = "exit 1"
type = "test"

[tasks.third-task]
name = "Third Task"
command = "exit 2"
type = "test"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *errorScenariosContext) theOutputShouldShowTheFirstTaskSucceeded() error {
	if !strings.Contains(c.output, "first-task") && !strings.Contains(c.output, "First Task") {
		return fmt.Errorf("expected output to show first task, got: %s", c.output)
	}
	return nil
}

func (c *errorScenariosContext) theOutputShouldShowTheSecondTaskFailed() error {
	if !strings.Contains(c.output, "second-task") && !strings.Contains(c.output, "Second Task") {
		return fmt.Errorf("expected output to show second task, got: %s", c.output)
	}
	return nil
}

func (c *errorScenariosContext) theOutputShouldShowTheThirdTaskFailed() error {
	if !strings.Contains(c.output, "third-task") && !strings.Contains(c.output, "Third Task") {
		return fmt.Errorf("expected output to show third task, got: %s", c.output)
	}
	return nil
}

func (c *errorScenariosContext) aConfigWithATaskThatHasAnInvalidWorkingDirectory() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "invalid-workdir.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.invalid-workdir]
name = "Invalid Workdir"
command = "echo test"
type = "test"
workingDir = "/this/directory/does/not/exist/12345"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *errorScenariosContext) theOutputShouldIndicateDirectoryError() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "directory") && !strings.Contains(lowerOutput, "no such file") && !strings.Contains(lowerOutput, "workdir") {
		return fmt.Errorf("expected output to indicate directory error, got: %s", c.output)
	}
	return nil
}

func (c *errorScenariosContext) aNonExistentConfigFilePath() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "does-not-exist.toml")
	return nil
}

func (c *errorScenariosContext) iRunDevpipeWithThatConfigPath() error {
	return c.iRunDevpipeWithThatConfig()
}

func (c *errorScenariosContext) theOutputShouldIndicateConfigFileNotFound() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "not found") && !strings.Contains(lowerOutput, "no such file") && !strings.Contains(lowerOutput, "does not exist") {
		return fmt.Errorf("expected output to indicate config file not found, got: %s", c.output)
	}
	return nil
}

func (c *errorScenariosContext) anEmptyConfigFile() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "empty.toml")

	// Create empty file
	return os.WriteFile(c.configPath, []byte(""), 0644)
}

func (c *errorScenariosContext) theOutputShouldIndicateConfigurationError() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "config") && !strings.Contains(lowerOutput, "error") && !strings.Contains(lowerOutput, "invalid") {
		return fmt.Errorf("expected output to indicate configuration error, got: %s", c.output)
	}
	return nil
}

func (c *errorScenariosContext) aConfigWithATaskThatHasAnEmptyCommand() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-error-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "empty-command.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.empty-command]
name = "Empty Command"
command = ""
type = "test"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *errorScenariosContext) theOutputShouldIndicateInvalidCommand() error {
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "command") && !strings.Contains(lowerOutput, "empty") && !strings.Contains(lowerOutput, "invalid") {
		return fmt.Errorf("expected output to indicate invalid command, got: %s", c.output)
	}
	return nil
}

func InitializeErrorScenariosScenario(ctx *godog.ScenarioContext, shared *sharedContext) {
	c := &errorScenariosContext{sharedContext: shared}

	ctx.Step(`^a config file with malformed TOML syntax$`, c.aConfigFileWithMalformedTOMLSyntax)
	ctx.Step(`^I run devpipe with that config$`, c.iRunDevpipeWithThatConfig)
	ctx.Step(`^the execution should fail$`, c.theExecutionShouldFail)
	ctx.Step(`^the output should contain "([^"]*)"$`, c.theOutputShouldContain)
	ctx.Step(`^a config file with a task missing the command field$`, c.aConfigFileWithATaskMissingTheCommandField)
	ctx.Step(`^a config file with a task missing the name field$`, c.aConfigFileWithATaskMissingTheNameField)
	ctx.Step(`^a config with a task that exits with code (\d+)$`, c.aConfigWithATaskThatExitsWithCode)
	ctx.Step(`^the output should indicate task failure$`, c.theOutputShouldIndicateTaskFailure)
	ctx.Step(`^the exit code should be non-zero$`, c.theExitCodeShouldBeNonZero)
	ctx.Step(`^a config with a task that has a (\d+) second timeout$`, c.aConfigWithATaskThatHasASecondTimeout)
	ctx.Step(`^the task sleeps for (\d+) seconds$`, c.theTaskSleepsForSeconds)
	ctx.Step(`^a config with a task using a non-existent binary$`, c.aConfigWithATaskUsingANonExistentBinary)
	ctx.Step(`^the output should indicate command not found$`, c.theOutputShouldIndicateCommandNotFound)
	ctx.Step(`^a config with a task using an invalid command path$`, c.aConfigWithATaskUsingAnInvalidCommandPath)
	ctx.Step(`^a config with three tasks where the second and third fail$`, c.aConfigWithThreeTasksWhereTheSecondAndThirdFail)
	ctx.Step(`^the output should show the first task succeeded$`, c.theOutputShouldShowTheFirstTaskSucceeded)
	ctx.Step(`^the output should show the second task failed$`, c.theOutputShouldShowTheSecondTaskFailed)
	ctx.Step(`^the output should show the third task failed$`, c.theOutputShouldShowTheThirdTaskFailed)
	ctx.Step(`^a config with a task that has an invalid working directory$`, c.aConfigWithATaskThatHasAnInvalidWorkingDirectory)
	ctx.Step(`^the output should indicate directory error$`, c.theOutputShouldIndicateDirectoryError)
	ctx.Step(`^a non-existent config file path$`, c.aNonExistentConfigFilePath)
	ctx.Step(`^I run devpipe with that config path$`, c.iRunDevpipeWithThatConfigPath)
	ctx.Step(`^the output should indicate config file not found$`, c.theOutputShouldIndicateConfigFileNotFound)
	ctx.Step(`^an empty config file$`, c.anEmptyConfigFile)
	ctx.Step(`^the output should indicate configuration error$`, c.theOutputShouldIndicateConfigurationError)
	ctx.Step(`^a config with a task that has an empty command$`, c.aConfigWithATaskThatHasAnEmptyCommand)
	ctx.Step(`^the output should indicate invalid command$`, c.theOutputShouldIndicateInvalidCommand)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
