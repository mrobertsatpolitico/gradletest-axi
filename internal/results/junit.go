package results

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Aggregate struct {
	Files       int
	Total       int
	Passed      int
	Failed      int
	Skipped     int
	Duration    time.Duration
	Failures    []Failure
	ParseErrors []error
}

type Failure struct {
	Test     string
	Location string
	Message  string
	Detail   string
}

type testSuitesXML struct {
	Suites []testSuiteXML `xml:"testsuite"`
}

type testSuiteXML struct {
	Name      string        `xml:"name,attr"`
	TestCases []testCaseXML `xml:"testcase"`
}

type testCaseXML struct {
	Name      string      `xml:"name,attr"`
	ClassName string      `xml:"classname,attr"`
	Time      string      `xml:"time,attr"`
	Failure   *problemXML `xml:"failure"`
	Error     *problemXML `xml:"error"`
	Skipped   *struct{}   `xml:"skipped"`
}

type problemXML struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

var sourceLocationPattern = regexp.MustCompile(`\(([^()\s]+:\d+)\)`)

func ParseFiles(paths []string) Aggregate {
	aggregate := Aggregate{Files: len(paths)}
	for _, path := range paths {
		cases, err := parseFile(path)
		if err != nil {
			aggregate.ParseErrors = append(aggregate.ParseErrors, err)
			continue
		}
		for _, testCase := range cases {
			aggregate.Total++
			aggregate.Duration += parseDuration(testCase.Time)
			switch {
			case testCase.Failure != nil:
				aggregate.Failed++
				aggregate.Failures = append(aggregate.Failures, toFailure(testCase, testCase.Failure))
			case testCase.Error != nil:
				aggregate.Failed++
				aggregate.Failures = append(aggregate.Failures, toFailure(testCase, testCase.Error))
			case testCase.Skipped != nil:
				aggregate.Skipped++
			default:
				aggregate.Passed++
			}
		}
	}
	sort.Slice(aggregate.Failures, func(i, j int) bool {
		return aggregate.Failures[i].Test < aggregate.Failures[j].Test
	})
	return aggregate
}

func parseFile(path string) ([]testCaseXML, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read JUnit XML %s: %w", path, err)
	}
	var root struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal(payload, &root); err != nil {
		return nil, fmt.Errorf("parse JUnit XML %s: %w", path, err)
	}

	switch root.XMLName.Local {
	case "testsuite":
		var suite testSuiteXML
		if err := xml.Unmarshal(payload, &suite); err != nil {
			return nil, fmt.Errorf("parse JUnit suite %s: %w", path, err)
		}
		return suite.TestCases, nil
	case "testsuites":
		var suites testSuitesXML
		if err := xml.Unmarshal(payload, &suites); err != nil {
			return nil, fmt.Errorf("parse JUnit suites %s: %w", path, err)
		}
		var cases []testCaseXML
		for _, suite := range suites.Suites {
			cases = append(cases, suite.TestCases...)
		}
		return cases, nil
	default:
		return nil, fmt.Errorf("parse JUnit XML %s: unsupported root element %q", path, root.XMLName.Local)
	}
}

func parseDuration(value string) time.Duration {
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || seconds < 0 {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}

func toFailure(testCase testCaseXML, problem *problemXML) Failure {
	testName := testCase.Name
	if testCase.ClassName != "" {
		testName = testCase.ClassName + "." + testCase.Name
	}
	body := strings.TrimSpace(problem.Body)
	message := strings.TrimSpace(problem.Message)
	if message == "" {
		message = firstNonEmptyLine(body)
	}
	if message == "" {
		message = strings.TrimSpace(problem.Type)
	}
	if message == "" {
		message = "test failed without a message"
	}

	location := ""
	if match := sourceLocationPattern.FindStringSubmatch(body); len(match) == 2 {
		location = filepath.ToSlash(match[1])
	}
	return Failure{
		Test:     testName,
		Location: location,
		Message:  message,
		Detail:   body,
	}
}

func firstNonEmptyLine(value string) string {
	for _, line := range strings.Split(value, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
