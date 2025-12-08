package features

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cucumber/godog"
)

type configValidationContext struct {
	*sharedContext
}

func (c *configValidationContext) aValidConfigFile(filename string) error {
	// Use the actual config.toml from the project root (one level up from features/)
	c.configPath = filepath.Join("..", filename)
	return nil
}

func (c *configValidationContext) iRunDevpipeValidate() error {
	cmd := exec.Command(c.devpipeBinary, "validate", c.configPath)
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

func (c *configValidationContext) theValidationShouldSucceed() error {
	if c.exitCode != 0 {
		return fmt.Errorf("expected validation to succeed (exit 0), got exit code %d\nOutput: %s", c.exitCode, c.output)
	}
	return nil
}

func (c *configValidationContext) aConfigFileWithMissingTaskCommand() error {
	// Create a temporary invalid config with malformed TOML
	c.tempDir = os.TempDir()
	c.configPath = filepath.Join(c.tempDir, "invalid-config.toml")

	// Invalid TOML: unclosed string and missing command field
	invalidConfig := `[defaults]
outputRoot = ".devpipe"

[tasks.broken-task]
name = "Broken Task
command = 
type = "test"
`

	return os.WriteFile(c.configPath, []byte(invalidConfig), 0644)
}

func (c *configValidationContext) iRunDevpipeValidateOnIt() error {
	return c.iRunDevpipeValidate()
}

func (c *configValidationContext) theValidationShouldFail() error {
	if c.exitCode == 0 {
		return fmt.Errorf("expected validation to fail (non-zero exit), got exit code 0\nOutput: %s", c.output)
	}
	return nil
}

// New validation error scenarios
func (c *configValidationContext) aConfigWithInvalidUIMode() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validation-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"
uiMode = "invalid-mode"

[tasks.test]
name = "Test"
command = "echo test"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *configValidationContext) aConfigWithInvalidAnimatedGroupBy() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validation-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"
animatedGroupBy = "invalid-groupby"

[tasks.test]
name = "Test"
command = "echo test"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *configValidationContext) aConfigWithNegativeFastThreshold() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validation-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"
fastThreshold = -100

[tasks.test]
name = "Test"
command = "echo test"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *configValidationContext) aConfigWithNegativeAnimationRefreshRate() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validation-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"
animationRefreshMs = -50

[tasks.test]
name = "Test"
command = "echo test"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *configValidationContext) aConfigWithInvalidGitMode() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validation-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[defaults.git]
mode = "invalid-git-mode"

[tasks.test]
name = "Test"
command = "echo test"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *configValidationContext) aConfigWithInvalidTaskLevelFixType() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validation-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.test]
name = "Test"
command = "echo test"
type = "test"
fixType = "invalid-fix-type"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *configValidationContext) aConfigWithFixTypeButMissingFixCommand() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validation-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.test]
name = "Test"
command = "echo test"
type = "test"
fixType = "auto"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

// Warning scenarios
func (c *configValidationContext) aConfigWithGitRefModeButNoRefSpecified() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validation-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[defaults.git]
mode = "ref"

[tasks.test]
name = "Test"
command = "echo test"
type = "test"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *configValidationContext) aConfigWithMetricsFormatButNoMetricsPath() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validation-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.test]
name = "Test"
command = "echo test"
type = "test"
outputType = "junit"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *configValidationContext) aConfigWithMetricsPathButNoMetricsFormat() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-validation-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	config := `[defaults]
outputRoot = ".devpipe"

[tasks.test]
name = "Test"
command = "echo test"
type = "test"
outputPath = "results.xml"
`
	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *configValidationContext) iRunDevpipeValidateWithThatConfig() error {
	cmd := exec.Command(c.devpipeBinary, "validate", c.configPath)
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

// Assertion helpers
func (c *configValidationContext) theOutputShouldIndicateInvalidUIMode() error {
	return c.theOutputShouldContain("Invalid UI mode")
}

func (c *configValidationContext) theOutputShouldIndicateInvalidGroupBy() error {
	return c.theOutputShouldContain("Invalid groupBy")
}

func (c *configValidationContext) theOutputShouldIndicateNegativeThreshold() error {
	return c.theOutputShouldContain("must be non-negative")
}

func (c *configValidationContext) theOutputShouldIndicateNegativeRefreshRate() error {
	return c.theOutputShouldContain("must be non-negative")
}

func (c *configValidationContext) theOutputShouldIndicateInvalidGitMode() error {
	return c.theOutputShouldContain("Invalid git mode")
}

func (c *configValidationContext) theOutputShouldIndicateInvalidFixType() error {
	return c.theOutputShouldContain("Invalid fix type")
}

func (c *configValidationContext) theOutputShouldIndicateMissingFixCommand() error {
	return c.theOutputShouldContain("fixCommand is not specified")
}

func (c *configValidationContext) theOutputShouldShowValidationWarningAboutMissingRef() error {
	return c.theOutputShouldContain("no ref is specified")
}

func (c *configValidationContext) theOutputShouldShowValidationWarningAboutMissingMetricsPath() error {
	return c.theOutputShouldContain("metricsPath is not specified")
}

func (c *configValidationContext) theOutputShouldShowValidationWarningAboutMissingMetricsFormat() error {
	return c.theOutputShouldContain("metricsFormat is not specified")
}

func (c *configValidationContext) theExecutionShouldSucceed() error {
	if c.exitCode != 0 {
		return fmt.Errorf("expected execution to succeed (exit 0), got exit code %d\nOutput: %s", c.exitCode, c.output)
	}
	return nil
}

func InitializeConfigValidationScenario(ctx *godog.ScenarioContext, shared *sharedContext) {
	c := &configValidationContext{sharedContext: shared}

	ctx.Step(`^a valid config file "([^"]*)"$`, c.aValidConfigFile)
	ctx.Step(`^I run devpipe validate$`, c.iRunDevpipeValidate)
	ctx.Step(`^the validation should succeed$`, c.theValidationShouldSucceed)
	ctx.Step(`^the output should contain "([^"]*)"$`, c.theOutputShouldContain)
	ctx.Step(`^a config file with missing task command$`, c.aConfigFileWithMissingTaskCommand)
	ctx.Step(`^I run devpipe validate on it$`, c.iRunDevpipeValidateOnIt)
	ctx.Step(`^the validation should fail$`, c.theValidationShouldFail)

	// New validation error scenarios
	ctx.Step(`^a config with invalid UI mode$`, c.aConfigWithInvalidUIMode)
	ctx.Step(`^a config with invalid animated group-by$`, c.aConfigWithInvalidAnimatedGroupBy)
	ctx.Step(`^a config with negative fast threshold$`, c.aConfigWithNegativeFastThreshold)
	ctx.Step(`^a config with negative animation refresh rate$`, c.aConfigWithNegativeAnimationRefreshRate)
	ctx.Step(`^a config with invalid git mode$`, c.aConfigWithInvalidGitMode)
	ctx.Step(`^a config with invalid task-level fix type$`, c.aConfigWithInvalidTaskLevelFixType)
	ctx.Step(`^a config with fix type but missing fix command$`, c.aConfigWithFixTypeButMissingFixCommand)
	ctx.Step(`^a config with git ref mode but no ref specified$`, c.aConfigWithGitRefModeButNoRefSpecified)
	ctx.Step(`^a config with metrics format but no metrics path$`, c.aConfigWithMetricsFormatButNoMetricsPath)
	ctx.Step(`^a config with metrics path but no metrics format$`, c.aConfigWithMetricsPathButNoMetricsFormat)

	ctx.Step(`^I run devpipe validate with that config$`, c.iRunDevpipeValidateWithThatConfig)
	ctx.Step(`^the execution should fail$`, c.theValidationShouldFail)
	ctx.Step(`^the execution should succeed$`, c.theExecutionShouldSucceed)

	ctx.Step(`^the output should indicate invalid UI mode$`, c.theOutputShouldIndicateInvalidUIMode)
	ctx.Step(`^the output should indicate invalid group-by$`, c.theOutputShouldIndicateInvalidGroupBy)
	ctx.Step(`^the output should indicate negative threshold$`, c.theOutputShouldIndicateNegativeThreshold)
	ctx.Step(`^the output should indicate negative refresh rate$`, c.theOutputShouldIndicateNegativeRefreshRate)
	ctx.Step(`^the output should indicate invalid git mode$`, c.theOutputShouldIndicateInvalidGitMode)
	ctx.Step(`^the output should indicate invalid fix type$`, c.theOutputShouldIndicateInvalidFixType)
	ctx.Step(`^the output should indicate missing fix command$`, c.theOutputShouldIndicateMissingFixCommand)
	ctx.Step(`^the output should show validation warning about missing ref$`, c.theOutputShouldShowValidationWarningAboutMissingRef)
	ctx.Step(`^the output should show validation warning about missing metrics path$`, c.theOutputShouldShowValidationWarningAboutMissingMetricsPath)
	ctx.Step(`^the output should show validation warning about missing metrics format$`, c.theOutputShouldShowValidationWarningAboutMissingMetricsFormat)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
