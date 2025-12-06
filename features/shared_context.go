package features

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// JUnit XML structures for validation
type JUnitTestSuites struct {
	XMLName    xml.Name         `xml:"testsuites"`
	TestSuites []JUnitTestSuite `xml:"testsuite"`
}

type JUnitTestSuite struct {
	XMLName  xml.Name        `xml:"testsuite"`
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Errors   int             `xml:"errors,attr"`
	Skipped  int             `xml:"skipped,attr"`
	TestCase []JUnitTestCase `xml:"testcase"`
}

type JUnitTestCase struct {
	Name      string `xml:"name,attr"`
	Classname string `xml:"classname,attr"`
	Time      string `xml:"time,attr"`
}

// sharedContext holds ALL state for a scenario - used by all step definitions
type sharedContext struct {
	// Common fields
	output        string
	exitCode      int
	tempDir       string
	configPath    string
	devpipeBinary string

	// Advanced features fields
	outputRoot  string
	workdirA    string
	workdirB    string
	junitPath   string
	sarifPath   string
	fixableFile string

	// Error scenarios fields
	taskTimeout int
	taskSleep   int

	// Task execution fields
	taskName string
}

// initDevpipeBinary finds and sets the devpipe binary path
func (c *sharedContext) initDevpipeBinary() {
	wd, _ := os.Getwd()
	c.devpipeBinary = filepath.Join(wd, "..", "devpipe")
	if _, err := os.Stat(c.devpipeBinary); os.IsNotExist(err) {
		c.devpipeBinary = filepath.Join(wd, "devpipe")
	}
}

// runDevpipe executes devpipe with the given config and stores output
func (c *sharedContext) runDevpipe() error {
	if c.configPath == "" {
		return fmt.Errorf("configPath not set")
	}
	if c.devpipeBinary == "" {
		c.initDevpipeBinary()
	}

	cmd := exec.Command(c.devpipeBinary, "-config", c.configPath)
	if c.tempDir != "" {
		cmd.Dir = c.tempDir
	}
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

// theExecutionShouldSucceed checks that the command succeeded
func (c *sharedContext) theExecutionShouldSucceed() error {
	if c.exitCode != 0 {
		return fmt.Errorf("expected execution to succeed (exit 0), got exit code %d\nOutput: %s", c.exitCode, c.output)
	}
	return nil
}

// theExecutionShouldFail checks that the command failed
func (c *sharedContext) theExecutionShouldFail() error {
	if c.exitCode == 0 {
		return fmt.Errorf("expected execution to fail (non-zero exit), got exit code 0\nOutput: %s", c.output)
	}
	return nil
}

// theOutputShouldContain checks that output contains the expected string (case-insensitive)
func (c *sharedContext) theOutputShouldContain(expected string) error {
	expected = strings.Trim(expected, `"`)
	if !strings.Contains(strings.ToLower(c.output), strings.ToLower(expected)) {
		return fmt.Errorf("expected output to contain %q, got: %s", expected, c.output)
	}
	return nil
}

// cleanup removes temporary directories
func (c *sharedContext) cleanup() {
	if c.tempDir != "" {
		_ = os.RemoveAll(c.tempDir)
	}
}
