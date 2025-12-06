package features

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
)

type runFlagsContext struct {
	configPath       string
	output           string
	exitCode         int
	tempDir          string
	devpipeBinary    string
	testFilePath     string
	customConfigPath string
}

func (c *runFlagsContext) aConfigWithTasks(taskList string) error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-flags-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")

	// Initialize git repo and create a file
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	// Parse task names from the list (e.g., "task-a", "task-b", and "task-c")
	tasks := strings.Split(taskList, ",")
	for i := range tasks {
		tasks[i] = strings.Trim(tasks[i], `" and`)
	}

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

`, c.tempDir)

	for _, task := range tasks {
		task = strings.TrimSpace(task)
		if task == "" {
			continue
		}
		config += fmt.Sprintf(`
[tasks.%s]
name = "%s"
command = "echo 'Running %s'"
type = "test"

`, task, task, task)
	}

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *runFlagsContext) iRunDevpipeWithOnly(taskName string) error {
	taskName = strings.Trim(taskName, `"`)
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "-only", taskName)
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

func (c *runFlagsContext) theExecutionShouldSucceed() error {
	if c.exitCode != 0 {
		return fmt.Errorf("expected execution to succeed (exit 0), got exit code %d\nOutput: %s", c.exitCode, c.output)
	}
	return nil
}

func (c *runFlagsContext) onlyShouldRun(taskName string) error {
	taskName = strings.Trim(taskName, `"`)
	if !strings.Contains(c.output, taskName) {
		return fmt.Errorf("expected %q to run, but not found in output: %s", taskName, c.output)
	}
	return nil
}

func (c *runFlagsContext) andShouldNotRun(taskList string) error {
	// Parse task names from the list (comma-separated)
	taskList = strings.Trim(taskList, `"`)
	tasks := strings.Split(taskList, ",")

	for _, task := range tasks {
		task = strings.TrimSpace(task)
		if strings.Contains(c.output, task) {
			return fmt.Errorf("expected %q not to run, but found in output: %s", task, c.output)
		}
	}
	return nil
}

func (c *runFlagsContext) iRunDevpipeWithSkip(taskName string) error {
	taskName = strings.Trim(taskName, `"`)
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "-skip", taskName)
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

func (c *runFlagsContext) andShouldRun(taskList string) error {
	// Parse task names from the list (comma-separated)
	taskList = strings.Trim(taskList, `"`)
	tasks := strings.Split(taskList, ",")

	for _, task := range tasks {
		task = strings.TrimSpace(task)
		if !strings.Contains(c.output, task) {
			return fmt.Errorf("expected %q to run, but not found in output: %s", task, c.output)
		}
	}
	return nil
}

func (c *runFlagsContext) shouldNotRun(taskName string) error {
	taskName = strings.Trim(taskName, `"`)
	if strings.Contains(c.output, taskName) {
		return fmt.Errorf("expected %q not to run, but found in output: %s", taskName, c.output)
	}
	return nil
}

func (c *runFlagsContext) aConfigWithATaskThatCreatesAFile() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-flags-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	c.testFilePath = filepath.Join(c.tempDir, "test-output.txt")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.create-file]
name = "Create File"
command = "echo 'test' > %s"
type = "test"
`, c.tempDir, c.testFilePath)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *runFlagsContext) iRunDevpipeWithDryRun() error {
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "-dry-run")
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

func (c *runFlagsContext) theOutputShouldContain(expected string) error {
	expected = strings.Trim(expected, `"`)
	if !strings.Contains(strings.ToLower(c.output), strings.ToLower(expected)) {
		return fmt.Errorf("expected output to contain %q, got: %s", expected, c.output)
	}
	return nil
}

func (c *runFlagsContext) theFileShouldNotBeCreated() error {
	if _, err := os.Stat(c.testFilePath); !os.IsNotExist(err) {
		return fmt.Errorf("expected file %s not to be created, but it exists", c.testFilePath)
	}
	return nil
}

func (c *runFlagsContext) aConfigWithAPassingTaskAndAFailingTask() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-flags-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.passing-task]
name = "Passing Task"
command = "echo 'pass'"
type = "test"

[tasks.failing-task]
name = "Failing Task"
command = "exit 1"
type = "test"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *runFlagsContext) iRunDevpipeWithFailFast() error {
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

func (c *runFlagsContext) theExecutionShouldFail() error {
	if c.exitCode == 0 {
		return fmt.Errorf("expected execution to fail (non-zero exit), got exit code 0\nOutput: %s", c.output)
	}
	return nil
}

func (c *runFlagsContext) aConfigWithASimpleTask() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-flags-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.simple-task]
name = "Simple Task"
command = "echo 'hello'"
type = "test"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *runFlagsContext) iRunDevpipeWithVerbose() error {
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "-verbose")
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

func (c *runFlagsContext) iRunDevpipeWithUI(uiMode string) error {
	uiMode = strings.Trim(uiMode, `"`)
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "-ui", uiMode)
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

func (c *runFlagsContext) theOutputShouldNotContainAnimationCharacters() error {
	// Check for common animation characters like spinners, progress bars
	animationChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	for _, char := range animationChars {
		if strings.Contains(c.output, char) {
			return fmt.Errorf("expected no animation characters, but found %q in output", char)
		}
	}
	return nil
}

func (c *runFlagsContext) iRunDevpipeWithNoColor() error {
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "-no-color")
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

func (c *runFlagsContext) theOutputShouldNotContainANSIColorCodes() error {
	// Check for ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	if ansiRegex.MatchString(c.output) {
		return fmt.Errorf("expected no ANSI color codes, but found them in output")
	}
	return nil
}

func (c *runFlagsContext) aConfigFileAtACustomPath() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-flags-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	// Create config at a non-standard path
	c.customConfigPath = filepath.Join(c.tempDir, "custom", "my-config.toml")
	if err := os.MkdirAll(filepath.Dir(c.customConfigPath), 0755); err != nil {
		return err
	}

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.custom-config-task]
name = "Custom Config Task"
command = "echo 'from custom config'"
type = "test"
`, c.tempDir)

	return os.WriteFile(c.customConfigPath, []byte(config), 0644)
}

func (c *runFlagsContext) iRunDevpipeWithConfigPointingToThatPath() error {
	cmd := exec.Command(c.devpipeBinary, "-config", c.customConfigPath)
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

func (c *runFlagsContext) tasksFromTheCustomConfigShouldRun() error {
	if !strings.Contains(c.output, "custom-config-task") {
		return fmt.Errorf("expected custom config task to run, but not found in output: %s", c.output)
	}
	return nil
}

func (c *runFlagsContext) aConfigWithATaskThatWritesAFile() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-flags-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	c.testFilePath = filepath.Join(c.tempDir, "output.txt")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.write-file]
name = "Write File"
command = "echo test > %s"
type = "test"
`, c.tempDir, c.testFilePath)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *runFlagsContext) theFileShouldNotExist() error {
	if _, err := os.Stat(c.testFilePath); !os.IsNotExist(err) {
		return fmt.Errorf("expected file %s not to exist, but it does", c.testFilePath)
	}
	return nil
}

func (c *runFlagsContext) aConfigWithSequentialTasksWhereSecondFails() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-flags-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")

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
command = "echo first"
type = "test"

[tasks.failing-task]
name = "Failing Task"
command = "exit 1"
type = "test"

[tasks.third-task]
name = "Third Task"
command = "echo third"
type = "test"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *runFlagsContext) laterTasksShouldNotRun() error {
	if strings.Contains(c.output, "third-task") || strings.Contains(c.output, "Third Task") {
		return fmt.Errorf("expected later tasks not to run, but found them in output")
	}
	return nil
}

func (c *runFlagsContext) theOutputShouldShowVerboseDetails() error {
	// Verbose mode typically shows more detailed logging
	if len(c.output) < 100 {
		return fmt.Errorf("expected verbose output to be detailed, but got short output: %s", c.output)
	}
	return nil
}

func (c *runFlagsContext) aConfigWithFastAndSlowTasks() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-flags-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"
fastThreshold = 5

[defaults.git]
mode = ""

[tasks.fast-task]
name = "Fast Task"
command = "echo fast"
type = "test"

[tasks.slow-task]
name = "Slow Task"
command = "sleep 10"
type = "test"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *runFlagsContext) iRunDevpipeWithFast() error {
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "-fast")
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

func (c *runFlagsContext) slowTasksShouldBeSkipped() error {
	if !strings.Contains(strings.ToLower(c.output), "skip") {
		return fmt.Errorf("expected slow tasks to be skipped, but no skip message in output")
	}
	return nil
}

func (c *runFlagsContext) fastTasksShouldRun() error {
	if !strings.Contains(c.output, "fast-task") && !strings.Contains(c.output, "Fast Task") {
		return fmt.Errorf("expected fast task to run, but not found in output")
	}
	return nil
}

func (c *runFlagsContext) aConfigWithAFixableTask() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-flags-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")

	// Initialize git repo
	cmd := exec.Command("git", "init", c.tempDir)
	_ = cmd.Run()
	_ = os.WriteFile(filepath.Join(c.tempDir, "dummy.txt"), []byte("test"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s/.devpipe"

[defaults.git]
mode = ""

[tasks.fixable-task]
name = "Fixable Task"
command = "echo needs fix"
type = "check-format"
fixCommand = "echo fixed"
`, c.tempDir)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *runFlagsContext) iRunDevpipeWithFixType(fixType string) error {
	fixType = strings.Trim(fixType, `"`)
	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath, "-fix-type", fixType)
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

func (c *runFlagsContext) iRunDevpipeWithMultipleSkipFlagsFor(taskList string) error {
	taskList = strings.Trim(taskList, `"`)
	tasks := strings.Split(taskList, ",")

	args := []string{"-config", c.configPath}
	for _, task := range tasks {
		task = strings.TrimSpace(task)
		args = append(args, "-skip", task)
	}

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

func (c *runFlagsContext) cleanup() {
	if c.tempDir != "" {
		_ = os.RemoveAll(c.tempDir) // Best effort cleanup
	}
}

func InitializeRunFlagsScenario(ctx *godog.ScenarioContext) {
	c := &runFlagsContext{}

	// Find the devpipe binary
	wd, _ := os.Getwd()
	c.devpipeBinary = filepath.Join(wd, "..", "devpipe")
	if _, err := os.Stat(c.devpipeBinary); os.IsNotExist(err) {
		c.devpipeBinary = filepath.Join(wd, "devpipe")
	}

	ctx.Step(`^a config with tasks "([^"]*)"$`, c.aConfigWithTasks)
	ctx.Step(`^I run devpipe with --only "([^"]*)"$`, c.iRunDevpipeWithOnly)
	ctx.Step(`^the execution should succeed$`, c.theExecutionShouldSucceed)
	ctx.Step(`^only "([^"]*)" should run$`, c.onlyShouldRun)
	ctx.Step(`^"([^"]*)" should not run$`, c.andShouldNotRun)
	ctx.Step(`^I run devpipe with --skip "([^"]*)"$`, c.iRunDevpipeWithSkip)
	ctx.Step(`^"([^"]*)" should run$`, c.andShouldRun)
	ctx.Step(`^a config with a task that writes a file$`, c.aConfigWithATaskThatWritesAFile)
	ctx.Step(`^I run devpipe with --dry-run$`, c.iRunDevpipeWithDryRun)
	ctx.Step(`^the file should not exist$`, c.theFileShouldNotExist)
	ctx.Step(`^a config with sequential tasks where second fails$`, c.aConfigWithSequentialTasksWhereSecondFails)
	ctx.Step(`^I run devpipe with --fail-fast$`, c.iRunDevpipeWithFailFast)
	ctx.Step(`^the execution should fail$`, c.theExecutionShouldFail)
	ctx.Step(`^later tasks should not run$`, c.laterTasksShouldNotRun)
	ctx.Step(`^a config with a simple task$`, c.aConfigWithASimpleTask)
	ctx.Step(`^I run devpipe with --verbose$`, c.iRunDevpipeWithVerbose)
	ctx.Step(`^the output should show verbose details$`, c.theOutputShouldShowVerboseDetails)
	ctx.Step(`^a config with fast and slow tasks$`, c.aConfigWithFastAndSlowTasks)
	ctx.Step(`^I run devpipe with --fast$`, c.iRunDevpipeWithFast)
	ctx.Step(`^slow tasks should be skipped$`, c.slowTasksShouldBeSkipped)
	ctx.Step(`^fast tasks should run$`, c.fastTasksShouldRun)
	ctx.Step(`^I run devpipe with --ui "([^"]*)"$`, c.iRunDevpipeWithUI)
	ctx.Step(`^the output should not contain animation characters$`, c.theOutputShouldNotContainAnimationCharacters)
	ctx.Step(`^I run devpipe with --no-color$`, c.iRunDevpipeWithNoColor)
	ctx.Step(`^the output should not contain ANSI color codes$`, c.theOutputShouldNotContainANSIColorCodes)
	ctx.Step(`^a config file at a custom path$`, c.aConfigFileAtACustomPath)
	ctx.Step(`^I run devpipe with --config pointing to that path$`, c.iRunDevpipeWithConfigPointingToThatPath)
	ctx.Step(`^tasks from the custom config should run$`, c.tasksFromTheCustomConfigShouldRun)
	ctx.Step(`^a config with a fixable task$`, c.aConfigWithAFixableTask)
	ctx.Step(`^I run devpipe with --fix-type "([^"]*)"$`, c.iRunDevpipeWithFixType)
	ctx.Step(`^I run devpipe with multiple --skip flags for "([^"]*)"$`, c.iRunDevpipeWithMultipleSkipFlagsFor)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
