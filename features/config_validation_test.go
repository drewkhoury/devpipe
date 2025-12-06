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
	// Use the actual config.toml from the repo root (one level up from features/)
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

func InitializeConfigValidationScenario(ctx *godog.ScenarioContext, shared *sharedContext) {
	c := &configValidationContext{sharedContext: shared}

	ctx.Step(`^a valid config file "([^"]*)"$`, c.aValidConfigFile)
	ctx.Step(`^I run devpipe validate$`, c.iRunDevpipeValidate)
	ctx.Step(`^the validation should succeed$`, c.theValidationShouldSucceed)
	ctx.Step(`^the output should contain "([^"]*)"$`, c.theOutputShouldContain)
	ctx.Step(`^a config file with missing task command$`, c.aConfigFileWithMissingTaskCommand)
	ctx.Step(`^I run devpipe validate on it$`, c.iRunDevpipeValidateOnIt)
	ctx.Step(`^the validation should fail$`, c.theValidationShouldFail)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
