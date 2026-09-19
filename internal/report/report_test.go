package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"gradletest-axi/internal/results"
)

func TestRenderDocumentGolden(t *testing.T) {
	t.Parallel()

	gradleExit := 1
	document := Document{
		Status:     "failed",
		Kind:       "test",
		ExitCode:   1,
		GradleExit: &gradleExit,
		Duration:   1250 * time.Millisecond,
		Report:     "junit-fresh",
		Tests:      &Counts{Total: 2, Passed: 1, Failed: 1, DurationMS: 750},
		Failures: []Failure{{
			Test:     "com.example.WidgetTest.fails",
			Location: "WidgetTest.kt:42",
			Message:  "expected true",
		}},
	}
	var output bytes.Buffer
	if err := Render(&output, document); err != nil {
		t.Fatal(err)
	}
	want := `status: failed
kind: test
exit_code: 1
gradle_exit: 1
duration_ms: 1250
report: junit-fresh
tests:
  total: 2
  passed: 1
  failed: 1
  skipped: 0
  duration_ms: 750
failures[1]{test,location,message}:
  com.example.WidgetTest.fails,"WidgetTest.kt:42",expected true
`
	if output.String() != want {
		t.Fatalf("Render() output:\n%s\nwant:\n%s", output.String(), want)
	}
}

func TestFailuresFromAggregateBoundsDefaultAndExpandsFull(t *testing.T) {
	t.Parallel()

	aggregate := results.Aggregate{}
	for i := 0; i < 6; i++ {
		aggregate.Failures = append(aggregate.Failures, results.Failure{
			Test:    "test",
			Message: strings.Repeat("m", defaultMessageRunes+1),
			Detail:  "complete detail",
		})
	}

	bounded, truncated := FailuresFromAggregate(aggregate, false)
	if len(bounded) != defaultFailureLimit || !truncated {
		t.Fatalf("bounded failures = %d, truncated = %v", len(bounded), truncated)
	}
	full, truncated := FailuresFromAggregate(aggregate, true)
	if len(full) != 6 || truncated {
		t.Fatalf("full failures = %d, truncated = %v", len(full), truncated)
	}
	if full[0].Message != "complete detail" {
		t.Fatalf("full message = %q, want complete detail", full[0].Message)
	}
}

func TestSummarizeGradleFailure(t *testing.T) {
	t.Parallel()

	got := SummarizeGradleFailure("\x1b[31mFAILURE\x1b[0m\n* What went wrong:\nCompilation failed\nbad source\n\n* Try:\n--stacktrace\n")
	if got != "Compilation failed bad source" {
		t.Fatalf("SummarizeGradleFailure() = %q", got)
	}
}
