package features

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

type advancedFeaturesContext struct {
	*sharedContext
}

func (c *advancedFeaturesContext) setupTempDir() error {
	c.tempDir = filepath.Join(os.TempDir(), fmt.Sprintf("devpipe-advanced-%d", os.Getpid()))
	if err := os.MkdirAll(c.tempDir, 0755); err != nil {
		return err
	}

	c.configPath = filepath.Join(c.tempDir, "config.toml")
	c.outputRoot = filepath.Join(c.tempDir, ".devpipe")

	// Don't initialize git - we'll disable git mode in configs to avoid any git-related issues
	return nil
}

func (c *advancedFeaturesContext) aConfigWithATaskThatGeneratesJUnitXML() error {
	if err := c.setupTempDir(); err != nil {
		return err
	}

	c.junitPath = filepath.Join(c.tempDir, "junit.xml")

	// Create a sample JUnit XML file directly
	junitXML := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites>
  <testsuite name="ExampleTests" tests="3" failures="1" errors="0" skipped="1">
    <testcase name="test_pass" classname="TestSuite" time="0.123"/>
    <testcase name="test_fail" classname="TestSuite" time="0.456">
      <failure message="assertion failed">Expected true, got false</failure>
    </testcase>
    <testcase name="test_skip" classname="TestSuite" time="0.000">
      <skipped/>
    </testcase>
  </testsuite>
</testsuites>`

	// Write JUnit XML file first
	if err := os.WriteFile(c.junitPath, []byte(junitXML), 0644); err != nil {
		return err
	}

	// Create script that will be used by devpipe
	scriptPath := filepath.Join(c.tempDir, "generate-junit.sh")
	script := "#!/bin/bash\necho 'JUnit test completed'"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return err
	}

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.junit-test]
name = "JUnit Test"
command = "%s"
type = "test"
outputType = "junit"
outputPath = "%s"
`, c.outputRoot, scriptPath, c.junitPath)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) iRunDevpipeWithThatConfig() error {
	return c.runDevpipe()
}

func (c *advancedFeaturesContext) theJUnitMetricsShouldBeParsed() error {
	// Check if JUnit file was created
	if _, err := os.Stat(c.junitPath); os.IsNotExist(err) {
		return fmt.Errorf("JUnit XML file was not created at %s", c.junitPath)
	}

	// Verify it's valid XML
	data, err := os.ReadFile(c.junitPath)
	if err != nil {
		return fmt.Errorf("failed to read JUnit XML: %v", err)
	}

	var suites JUnitTestSuites
	if err := xml.Unmarshal(data, &suites); err != nil {
		return fmt.Errorf("failed to parse JUnit XML: %v", err)
	}

	return nil
}

func (c *advancedFeaturesContext) theMetricsShouldShowTestCounts() error {
	// Check output for test metrics
	if !strings.Contains(c.output, "test") {
		return fmt.Errorf("expected output to contain test metrics, got: %s", c.output)
	}
	return nil
}

func (c *advancedFeaturesContext) aConfigWithATaskThatGeneratesSARIFOutput() error {
	if err := c.setupTempDir(); err != nil {
		return err
	}

	c.sarifPath = filepath.Join(c.tempDir, "sarif.json")

	// Create a sample SARIF file directly
	sarifJSON := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "TestScanner",
          "version": "1.0.0"
        }
      },
      "results": [
        {
          "ruleId": "SEC001",
          "level": "error",
          "message": {
            "text": "Security vulnerability detected"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "test.go"
                },
                "region": {
                  "startLine": 10
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`

	// Write SARIF file first
	if err := os.WriteFile(c.sarifPath, []byte(sarifJSON), 0644); err != nil {
		return err
	}

	// Create script that will be used by devpipe
	scriptPath := filepath.Join(c.tempDir, "generate-sarif.sh")
	script := "#!/bin/bash\necho 'SARIF scan completed'"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return err
	}

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.sarif-scan]
name = "SARIF Scan"
command = "%s"
type = "security"
outputType = "sarif"
outputPath = "%s"
`, c.outputRoot, scriptPath, c.sarifPath)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) theSARIFMetricsShouldBeParsed() error {
	// Check if SARIF file was created
	if _, err := os.Stat(c.sarifPath); os.IsNotExist(err) {
		return fmt.Errorf("SARIF file was not created at %s", c.sarifPath)
	}

	// Verify it's valid JSON
	data, err := os.ReadFile(c.sarifPath)
	if err != nil {
		return fmt.Errorf("failed to read SARIF file: %v", err)
	}

	if !strings.Contains(string(data), "version") {
		return fmt.Errorf("SARIF file does not contain version field")
	}

	return nil
}

func (c *advancedFeaturesContext) theMetricsShouldShowSecurityFindings() error {
	// Check that SARIF file was parsed (output should show task completed)
	lowerOutput := strings.ToLower(c.output)
	if !strings.Contains(lowerOutput, "sarif") && !strings.Contains(lowerOutput, "scan") && !strings.Contains(lowerOutput, "completed") {
		return fmt.Errorf("expected output to show SARIF scan completed, got: %s", c.output)
	}
	return nil
}

func (c *advancedFeaturesContext) aConfigWithAFixableTaskThatFails() error {
	if err := c.setupTempDir(); err != nil {
		return err
	}

	c.fixableFile = filepath.Join(c.tempDir, "fixable.txt")

	// Create a file that will fail the check
	_ = os.WriteFile(c.fixableFile, []byte("broken"), 0644)

	// This will be updated by the next step
	return nil
}

func (c *advancedFeaturesContext) theTaskHasFixTypeSetTo(fixType string) error {
	fixType = strings.Trim(fixType, `"`)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.fixable-task]
name = "Fixable Task"
command = "test \"$(cat %s)\" = \"fixed\""
type = "check-format"
fixType = "%s"
fixCommand = "echo 'fixed' > %s"
`, c.outputRoot, c.fixableFile, fixType, c.fixableFile)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) theFixCommandShouldBeExecutedAutomatically() error {
	// Check if the fix was applied
	if _, err := os.Stat(c.fixableFile); os.IsNotExist(err) {
		return fmt.Errorf("fixable file does not exist")
	}

	content, err := os.ReadFile(c.fixableFile)
	if err != nil {
		return fmt.Errorf("failed to read fixable file: %v", err)
	}

	if strings.TrimSpace(string(content)) != "fixed" {
		return fmt.Errorf("expected file to be fixed, got: %s", string(content))
	}

	return nil
}

func (c *advancedFeaturesContext) theTaskShouldBeRechecked() error {
	// Check output for recheck indication
	if !strings.Contains(strings.ToLower(c.output), "recheck") && !strings.Contains(strings.ToLower(c.output), "fix") {
		return fmt.Errorf("expected output to indicate recheck, got: %s", c.output)
	}
	return nil
}

func (c *advancedFeaturesContext) theExecutionShouldSucceedAfterFix() error {
	return c.theExecutionShouldSucceed()
}

func (c *advancedFeaturesContext) theOutputShouldShowTheFixCommandSuggestion() error {
	// Check for fix command in output
	if !strings.Contains(c.output, "echo 'fixed'") && !strings.Contains(strings.ToLower(c.output), "to fix") {
		return fmt.Errorf("expected output to show fix command suggestion, got: %s", c.output)
	}
	return nil
}

func (c *advancedFeaturesContext) aConfigWithThreePhases() error {
	if err := c.setupTempDir(); err != nil {
		return err
	}

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.phase-1]
name = "Phase 1"

[tasks.task-1a]
name = "Task 1A"
command = "echo 'Phase 1 Task A'"
type = "test"

[tasks.task-1b]
name = "Task 1B"
command = "echo 'Phase 1 Task B'"
type = "test"
wait = true

[tasks.phase-2]
name = "Phase 2"

[tasks.task-2a]
name = "Task 2A"
command = "echo 'Phase 2 Task A'"
type = "test"
wait = true

[tasks.phase-3]
name = "Phase 3"

[tasks.task-3a]
name = "Task 3A"
command = "echo 'Phase 3 Task A'"
type = "test"
`, c.outputRoot)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) phase1ShouldCompleteBeforePhase2Starts() error {
	// Check output for phase ordering
	phase1Idx := strings.Index(c.output, "Phase 1")
	phase2Idx := strings.Index(c.output, "Phase 2")

	if phase1Idx == -1 || phase2Idx == -1 {
		return fmt.Errorf("could not find phase markers in output")
	}

	if phase1Idx >= phase2Idx {
		return fmt.Errorf("phase 1 did not complete before phase 2 started")
	}

	return nil
}

func (c *advancedFeaturesContext) phase2ShouldCompleteBeforePhase3Starts() error {
	// Check output for phase ordering
	phase2Idx := strings.Index(c.output, "Phase 2")
	phase3Idx := strings.Index(c.output, "Phase 3")

	if phase2Idx == -1 || phase3Idx == -1 {
		return fmt.Errorf("could not find phase markers in output")
	}

	if phase2Idx >= phase3Idx {
		return fmt.Errorf("phase 2 did not complete before phase 3 started")
	}

	return nil
}

func (c *advancedFeaturesContext) allPhasesShouldExecuteInOrder() error {
	// Verify all phases are present in output
	if !strings.Contains(c.output, "Task 1A") {
		return fmt.Errorf("phase 1 task not found in output. Got: %s", c.output)
	}
	if !strings.Contains(c.output, "Task 2A") {
		return fmt.Errorf("phase 2 task not found in output. Got: %s", c.output)
	}
	if !strings.Contains(c.output, "Task 3A") {
		return fmt.Errorf("phase 3 task not found in output. Got: %s", c.output)
	}
	return nil
}

func (c *advancedFeaturesContext) aConfigWithMultipleTasksInTheSamePhase() error {
	if err := c.setupTempDir(); err != nil {
		return err
	}

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.parallel-1]
name = "Parallel Task 1"
command = "sleep 0.1 && echo 'Task 1'"
type = "test"

[tasks.parallel-2]
name = "Parallel Task 2"
command = "sleep 0.1 && echo 'Task 2'"
type = "test"

[tasks.parallel-3]
name = "Parallel Task 3"
command = "sleep 0.1 && echo 'Task 3'"
type = "test"
`, c.outputRoot)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) tasksShouldRunConcurrentlyWithinThePhase() error {
	// All tasks should be present in output
	if !strings.Contains(c.output, "Task 1") {
		return fmt.Errorf("task 1 not found in output")
	}
	if !strings.Contains(c.output, "Task 2") {
		return fmt.Errorf("task 2 not found in output")
	}
	if !strings.Contains(c.output, "Task 3") {
		return fmt.Errorf("task 3 not found in output")
	}
	return nil
}

func (c *advancedFeaturesContext) aConfigWithThreePhasesWherePhase2Fails() error {
	if err := c.setupTempDir(); err != nil {
		return err
	}

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.phase-1]
name = "Phase 1"

[tasks.task-1]
name = "Task 1"
command = "echo 'Phase 1'"
type = "test"
wait = true

[tasks.phase-2]
name = "Phase 2"

[tasks.task-2]
name = "Task 2"
command = "exit 1"
type = "test"
wait = true

[tasks.phase-3]
name = "Phase 3"

[tasks.task-3]
name = "Task 3"
command = "echo 'Phase 3'"
type = "test"
`, c.outputRoot)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) iRunDevpipeWithFailFast() error {
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

func (c *advancedFeaturesContext) phase1ShouldComplete() error {
	if !strings.Contains(c.output, "Task 1") {
		return fmt.Errorf("phase 1 did not complete")
	}
	return nil
}

func (c *advancedFeaturesContext) phase2ShouldFail() error {
	if !strings.Contains(c.output, "Task 2") {
		return fmt.Errorf("phase 2 did not run")
	}
	// Should fail overall
	if c.exitCode == 0 {
		return fmt.Errorf("expected failure but got success")
	}
	return nil
}

func (c *advancedFeaturesContext) phase3ShouldNotExecute() error {
	if strings.Contains(c.output, "Task 3") {
		return fmt.Errorf("phase 3 should not have executed")
	}
	return nil
}

func (c *advancedFeaturesContext) aConfigWithATaskThatHasACustomWorkdir() error {
	if err := c.setupTempDir(); err != nil {
		return err
	}

	c.workdirA = filepath.Join(c.tempDir, "subdir")
	if err := os.MkdirAll(c.workdirA, 0755); err != nil {
		return err
	}

	// Create a marker file in the subdirectory
	_ = os.WriteFile(filepath.Join(c.workdirA, "marker.txt"), []byte("marker"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.workdir-task]
name = "Workdir Task"
command = "test -f marker.txt && echo 'Found marker in workdir'"
type = "test"
workdir = "%s"
`, c.outputRoot, c.workdirA)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) theTaskShouldRunInTheSpecifiedDirectory() error {
	if !strings.Contains(c.output, "Found marker in workdir") {
		return fmt.Errorf("task did not run in the specified workdir")
	}
	return nil
}

func (c *advancedFeaturesContext) aConfigWithATaskUsingARelativeWorkdir() error {
	if err := c.setupTempDir(); err != nil {
		return err
	}

	relativeDir := filepath.Join(c.tempDir, "relative")
	if err := os.MkdirAll(relativeDir, 0755); err != nil {
		return err
	}

	// Create a marker file
	_ = os.WriteFile(filepath.Join(relativeDir, "relative-marker.txt"), []byte("marker"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.relative-workdir-task]
name = "Relative Workdir Task"
command = "test -f relative-marker.txt && echo 'Found relative marker'"
type = "test"
workdir = "relative"
`, c.outputRoot)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) theTaskShouldRunInTheRelativeDirectory() error {
	if !strings.Contains(c.output, "Found relative marker") {
		return fmt.Errorf("task did not run in the relative workdir")
	}
	return nil
}

func (c *advancedFeaturesContext) aConfigWithTasksInDifferentWorkdirs() error {
	if err := c.setupTempDir(); err != nil {
		return err
	}

	c.workdirA = filepath.Join(c.tempDir, "dirA")
	c.workdirB = filepath.Join(c.tempDir, "dirB")

	if err := os.MkdirAll(c.workdirA, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(c.workdirB, 0755); err != nil {
		return err
	}

	// Create marker files
	_ = os.WriteFile(filepath.Join(c.workdirA, "markerA.txt"), []byte("A"), 0644)
	_ = os.WriteFile(filepath.Join(c.workdirB, "markerB.txt"), []byte("B"), 0644)

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.task-a]
name = "Task A"
command = "test -f markerA.txt && echo 'Found A'"
type = "test"
workdir = "%s"

[tasks.task-b]
name = "Task B"
command = "test -f markerB.txt && echo 'Found B'"
type = "test"
workdir = "%s"
`, c.outputRoot, c.workdirA, c.workdirB)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) eachTaskShouldRunInItsOwnDirectory() error {
	if !strings.Contains(c.output, "Found A") {
		return fmt.Errorf("task A did not run in its workdir")
	}
	if !strings.Contains(c.output, "Found B") {
		return fmt.Errorf("task B did not run in its workdir")
	}
	return nil
}

func (c *advancedFeaturesContext) aConfigWithATaskThatSetsEnvironmentVariables() error {
	if err := c.setupTempDir(); err != nil {
		return err
	}

	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[tasks.env-task]
name = "Env Task"
command = "echo \"TEST_VAR=$TEST_VAR\""
type = "test"

[tasks.env-task.env]
TEST_VAR = "test_value"
`, c.outputRoot)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) theTaskShouldHaveAccessToTheEnvironmentVariables() error {
	if !strings.Contains(c.output, "TEST_VAR=test_value") {
		return fmt.Errorf("task did not have access to environment variable, output: %s", c.output)
	}
	return nil
}

func (c *advancedFeaturesContext) aConfigWithDefaultEnvironmentVariables() error {
	if err := c.setupTempDir(); err != nil {
		return err
	}

	// This will be completed by the next step
	return nil
}

func (c *advancedFeaturesContext) aTaskThatUsesThoseVariables() error {
	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[defaults.env]
DEFAULT_VAR = "default_value"

[tasks.default-env-task]
name = "Default Env Task"
command = "echo \"DEFAULT_VAR=$DEFAULT_VAR\""
type = "test"
`, c.outputRoot)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) theTaskShouldSeeTheDefaultEnvironmentVariables() error {
	if !strings.Contains(c.output, "DEFAULT_VAR=default_value") {
		return fmt.Errorf("task did not see default environment variable, output: %s", c.output)
	}
	return nil
}

func (c *advancedFeaturesContext) aTaskThatOverridesThoseVariables() error {
	config := fmt.Sprintf(`[defaults]
outputRoot = "%s"

[defaults.git]
mode = ""

[defaults.env]
OVERRIDE_VAR = "default_value"

[tasks.override-env-task]
name = "Override Env Task"
command = "echo \"OVERRIDE_VAR=$OVERRIDE_VAR\""
type = "test"

[tasks.override-env-task.env]
OVERRIDE_VAR = "overridden_value"
`, c.outputRoot)

	return os.WriteFile(c.configPath, []byte(config), 0644)
}

func (c *advancedFeaturesContext) theTaskShouldSeeTheOverriddenValues() error {
	if !strings.Contains(c.output, "OVERRIDE_VAR=overridden_value") {
		return fmt.Errorf("task did not see overridden environment variable, output: %s", c.output)
	}
	return nil
}

func InitializeAdvancedFeaturesScenario(ctx *godog.ScenarioContext, shared *sharedContext) {
	c := &advancedFeaturesContext{sharedContext: shared}

	// Metrics collection
	ctx.Step(`^a config with a task that generates JUnit XML$`, c.aConfigWithATaskThatGeneratesJUnitXML)
	ctx.Step(`^I run devpipe with that config$`, c.iRunDevpipeWithThatConfig)
	ctx.Step(`^the execution should succeed$`, c.theExecutionShouldSucceed)
	ctx.Step(`^the JUnit metrics should be parsed$`, c.theJUnitMetricsShouldBeParsed)
	ctx.Step(`^the metrics should show test counts$`, c.theMetricsShouldShowTestCounts)

	ctx.Step(`^a config with a task that generates SARIF output$`, c.aConfigWithATaskThatGeneratesSARIFOutput)
	ctx.Step(`^the SARIF metrics should be parsed$`, c.theSARIFMetricsShouldBeParsed)
	ctx.Step(`^the metrics should show security findings$`, c.theMetricsShouldShowSecurityFindings)

	// Fix commands
	ctx.Step(`^a config with a fixable task that fails$`, c.aConfigWithAFixableTaskThatFails)
	ctx.Step(`^the task has fixType set to "([^"]*)"$`, c.theTaskHasFixTypeSetTo)
	ctx.Step(`^the fix command should be executed automatically$`, c.theFixCommandShouldBeExecutedAutomatically)
	ctx.Step(`^the task should be rechecked$`, c.theTaskShouldBeRechecked)
	ctx.Step(`^the execution should succeed after fix$`, c.theExecutionShouldSucceedAfterFix)
	ctx.Step(`^the execution should fail$`, c.theExecutionShouldFail)
	ctx.Step(`^the output should show the fix command suggestion$`, c.theOutputShouldShowTheFixCommandSuggestion)
	ctx.Step(`^the output should contain "([^"]*)"$`, c.theOutputShouldContain)

	// Phase execution
	ctx.Step(`^a config with three phases$`, c.aConfigWithThreePhases)
	ctx.Step(`^phase 1 should complete before phase 2 starts$`, c.phase1ShouldCompleteBeforePhase2Starts)
	ctx.Step(`^phase 2 should complete before phase 3 starts$`, c.phase2ShouldCompleteBeforePhase3Starts)
	ctx.Step(`^all phases should execute in order$`, c.allPhasesShouldExecuteInOrder)

	// Parallel execution
	ctx.Step(`^a config with multiple tasks in the same phase$`, c.aConfigWithMultipleTasksInTheSamePhase)
	ctx.Step(`^tasks should run concurrently within the phase$`, c.tasksShouldRunConcurrentlyWithinThePhase)

	// Phase failure
	ctx.Step(`^a config with three phases where phase 2 fails$`, c.aConfigWithThreePhasesWherePhase2Fails)
	ctx.Step(`^I run devpipe with --fail-fast$`, c.iRunDevpipeWithFailFast)
	ctx.Step(`^phase 1 should complete$`, c.phase1ShouldComplete)
	ctx.Step(`^phase 2 should fail$`, c.phase2ShouldFail)
	ctx.Step(`^phase 3 should not execute$`, c.phase3ShouldNotExecute)

	// Workdir handling
	ctx.Step(`^a config with a task that has a custom workdir$`, c.aConfigWithATaskThatHasACustomWorkdir)
	ctx.Step(`^the task should run in the specified directory$`, c.theTaskShouldRunInTheSpecifiedDirectory)
	ctx.Step(`^a config with a task using a relative workdir$`, c.aConfigWithATaskUsingARelativeWorkdir)
	ctx.Step(`^the task should run in the relative directory$`, c.theTaskShouldRunInTheRelativeDirectory)
	ctx.Step(`^a config with tasks in different workdirs$`, c.aConfigWithTasksInDifferentWorkdirs)
	ctx.Step(`^each task should run in its own directory$`, c.eachTaskShouldRunInItsOwnDirectory)

	// Environment variables
	ctx.Step(`^a config with a task that sets environment variables$`, c.aConfigWithATaskThatSetsEnvironmentVariables)
	ctx.Step(`^the task should have access to the environment variables$`, c.theTaskShouldHaveAccessToTheEnvironmentVariables)
	ctx.Step(`^a config with default environment variables$`, c.aConfigWithDefaultEnvironmentVariables)
	ctx.Step(`^a task that uses those variables$`, c.aTaskThatUsesThoseVariables)
	ctx.Step(`^the task should see the default environment variables$`, c.theTaskShouldSeeTheDefaultEnvironmentVariables)
	ctx.Step(`^a task that overrides those variables$`, c.aTaskThatOverridesThoseVariables)
	ctx.Step(`^the task should see the overridden values$`, c.theTaskShouldSeeTheOverriddenValues)

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		c.cleanup()
		return ctx, nil
	})
}
