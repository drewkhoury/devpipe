package metrics

import (
	"encoding/xml"
	"os"

	"github.com/drew/devpipe/internal/model"
)

// JUnitTestSuite represents a JUnit XML test suite
type JUnitTestSuite struct {
	XMLName  xml.Name `xml:"testsuite"`
	Name     string   `xml:"name,attr"`
	Tests    int      `xml:"tests,attr"`
	Failures int      `xml:"failures,attr"`
	Errors   int      `xml:"errors,attr"`
	Skipped  int      `xml:"skipped,attr"`
	Time     float64  `xml:"time,attr"`
}

// ParseJUnitXML parses a JUnit XML file and returns metrics
func ParseJUnitXML(path string) (*model.StageMetrics, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var suite JUnitTestSuite
	if err := xml.Unmarshal(data, &suite); err != nil {
		return nil, err
	}

	return &model.StageMetrics{
		Kind:          "test", // Inferred from format
		SummaryFormat: "junit",
		Data: map[string]interface{}{
			"tests":    suite.Tests,
			"failures": suite.Failures,
			"errors":   suite.Errors,
			"skipped":  suite.Skipped,
			"time":     suite.Time,
		},
	}, nil
}
