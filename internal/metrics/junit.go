package metrics

import (
	"github.com/drew/devpipe/internal/model"
	"github.com/joshdk/go-junit"
)

// ParseJUnitXML parses a JUnit XML file and returns metrics
// Uses github.com/joshdk/go-junit for robust parsing of all JUnit XML variants
// Supports: single <testsuite>, <testsuites>, multiple root elements, etc.
func ParseJUnitXML(path string) (*model.TaskMetrics, error) {
	// Ingest JUnit XML file using battle-tested library
	suites, err := junit.IngestFile(path)
	if err != nil {
		return nil, err
	}

	// Aggregate metrics across all suites
	var totalTests, totalFailures, totalErrors, totalSkipped int
	var totalTime float64
	var testCases []map[string]interface{}

	for _, suite := range suites {
		totalTests += len(suite.Tests)
		for _, test := range suite.Tests {
			// Collect test case details
			testCase := map[string]interface{}{
				"name":      test.Name,
				"classname": test.Classname,
				"time":      test.Duration.Seconds(),
				"status":    string(test.Status),
			}

			// Add error/failure message if present
			if test.Error != nil {
				testCase["message"] = test.Error.Error()
			}

			testCases = append(testCases, testCase)

			switch test.Status {
			case junit.StatusFailed:
				totalFailures++
			case junit.StatusError:
				totalErrors++
			case junit.StatusSkipped:
				totalSkipped++
			}
			totalTime += test.Duration.Seconds()
		}
	}

	return &model.TaskMetrics{
		Kind:          "test",
		SummaryFormat: "junit",
		Data: map[string]interface{}{
			"tests":     totalTests,
			"failures":  totalFailures,
			"errors":    totalErrors,
			"skipped":   totalSkipped,
			"time":      totalTime,
			"testcases": testCases,
		},
	}, nil
}
