package report

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mrobertsatpolitico/gradletest-axi/internal/results"
)

const (
	defaultFailureLimit = 5
	defaultMessageRunes = 800
)

type Document struct {
	Status     string
	Kind       string
	Task       string
	ExitCode   int
	GradleExit *int
	Duration   time.Duration
	Report     string
	Tests      *Counts
	Failures   []Failure
	Error      string
	Warning    string
	Log        string
	Truncated  bool
	Help       []string
}

type Counts struct {
	Total      int
	Passed     int
	Failed     int
	Skipped    int
	DurationMS int64
}

type Failure struct {
	Test     string
	Location string
	Message  string
}

type ErrorDocument struct {
	Error    string
	ExitCode int
	Help     []string
}

type HelpDocument struct {
	Command     string
	Description string
	Usage       string
	Flags       []Flag
	Help        []string
}

type Flag struct {
	Flag        string
	Description string
}

type VersionDocument struct {
	Name    string
	Version string
}

func CountsFromAggregate(aggregate results.Aggregate) *Counts {
	return &Counts{
		Total:      aggregate.Total,
		Passed:     aggregate.Passed,
		Failed:     aggregate.Failed,
		Skipped:    aggregate.Skipped,
		DurationMS: aggregate.Duration.Milliseconds(),
	}
}

func FailuresFromAggregate(aggregate results.Aggregate, full bool) ([]Failure, bool) {
	source := aggregate.Failures
	truncated := false
	if !full && len(source) > defaultFailureLimit {
		source = source[:defaultFailureLimit]
		truncated = true
	}

	failures := make([]Failure, 0, len(source))
	for _, failure := range source {
		message := failure.Message
		if full && failure.Detail != "" {
			message = failure.Detail
		} else if !full {
			var messageTruncated bool
			message, messageTruncated = truncateRunes(message, defaultMessageRunes)
			truncated = truncated || messageTruncated
		}
		failures = append(failures, Failure{
			Test:     failure.Test,
			Location: failure.Location,
			Message:  message,
		})
	}
	return failures, truncated
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

func SummarizeGradleFailure(tail string) string {
	cleaned := ansiPattern.ReplaceAllString(tail, "")
	lines := strings.Split(cleaned, "\n")

	for index, line := range lines {
		if strings.TrimSpace(line) != "* What went wrong:" {
			continue
		}
		summary := collectMeaningfulLines(lines[index+1:], 6)
		if summary != "" {
			value, _ := truncateRunes(summary, 1200)
			return value
		}
	}

	for index := len(lines) - 1; index >= 0; index-- {
		line := strings.TrimSpace(lines[index])
		if strings.HasPrefix(line, "Execution failed for task ") ||
			strings.Contains(line, "Compilation failed") ||
			strings.Contains(line, "Could not compile") {
			value, _ := truncateRunes(line, 1200)
			return value
		}
	}
	return "Gradle test command failed without machine-readable test failures"
}

func collectMeaningfulLines(lines []string, limit int) string {
	collected := make([]string, 0, limit)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "* Try:") ||
			strings.HasPrefix(trimmed, "* Exception is:") ||
			strings.HasPrefix(trimmed, "BUILD FAILED") {
			break
		}
		if trimmed == "" {
			continue
		}
		collected = append(collected, trimmed)
		if len(collected) == limit {
			break
		}
	}
	return strings.Join(collected, " ")
}

func truncateRunes(value string, limit int) (string, bool) {
	runes := []rune(value)
	if len(runes) <= limit {
		return value, false
	}
	return fmt.Sprintf("%s… (truncated, %d chars total)", string(runes[:limit]), len(runes)), true
}
