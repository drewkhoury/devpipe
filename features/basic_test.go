package features

import (
	"fmt"
	"testing"

	"github.com/cucumber/godog"
)

// Test context to hold state between steps
type testContext struct {
	number int
	str    string
	result interface{}
}

func (tc *testContext) iHaveTheNumber(num int) error {
	tc.number = num
	return nil
}

func (tc *testContext) iAddToIt(num int) error {
	tc.result = tc.number + num
	return nil
}

func (tc *testContext) theResultShouldBe(expected int) error {
	if tc.result != expected {
		return fmt.Errorf("expected %d, got %v", expected, tc.result)
	}
	return nil
}

func (tc *testContext) iHaveTheString(str string) error {
	tc.str = str
	return nil
}

func (tc *testContext) iAppend(suffix string) error {
	tc.result = tc.str + suffix
	return nil
}

func (tc *testContext) theResultShouldBeString(expected string) error {
	if tc.result != expected {
		return fmt.Errorf("expected %q, got %v", expected, tc.result)
	}
	return nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	tc := &testContext{}

	sc.Step(`^I have the number (\d+)$`, tc.iHaveTheNumber)
	sc.Step(`^I add (\d+) to it$`, tc.iAddToIt)
	sc.Step(`^the result should be (\d+)$`, tc.theResultShouldBe)
	sc.Step(`^I have the string "([^"]*)"$`, tc.iHaveTheString)
	sc.Step(`^I append "([^"]*)"$`, tc.iAppend)
	sc.Step(`^the result should be "([^"]*)"$`, tc.theResultShouldBeString)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			// Create ONE shared context instance per scenario
			shared := &sharedContext{}
			shared.initDevpipeBinary()

			// Initialize all step definitions with the same shared context
			InitializeScenario(sc)
			InitializeConfigValidationScenario(sc, shared)
			InitializeErrorScenariosScenario(sc, shared)
			InitializeAdvancedFeaturesScenario(sc, shared)
			InitializeTaskExecutionScenario(sc, shared)
			InitializeRunFlagsScenario(sc)
			InitializeCommandsScenario(sc, shared)
			InitializeFailFastScenario(sc, shared)
			InitializeUnknownCommandsScenario(sc, shared)
			InitializePhaseExecutionScenario(sc, shared)
			InitializeSinceFlagScenario(sc, shared)
			InitializeConfigEdgeCasesScenario(sc, shared)
			InitializeWatchPathsScenario(sc, shared)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"."},
			Tags:     "~@wip", // Exclude work-in-progress scenarios
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
