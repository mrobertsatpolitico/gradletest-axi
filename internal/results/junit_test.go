package results

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFilesAggregatesSuitesAndFailures(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	first := writeResult(t, root, "module-a", `<?xml version="1.0"?>
<testsuite name="WidgetTest">
  <testcase name="passes" classname="com.example.WidgetTest" time="0.1"/>
  <testcase name="ignored" classname="com.example.WidgetTest" time="0.2"><skipped/></testcase>
  <testcase name="fails" classname="com.example.WidgetTest" time="0.3">
    <failure message="expected true but was false">java.lang.AssertionError
	at com.example.WidgetTest.fails(WidgetTest.kt:42)
    </failure>
  </testcase>
</testsuite>`)
	second := writeResult(t, root, "module-b", `<?xml version="1.0"?>
<testsuites>
  <testsuite name="OtherTest">
    <testcase name="errors" classname="com.example.OtherTest" time="0.4">
      <error>configuration exploded</error>
    </testcase>
  </testsuite>
</testsuites>`)

	got := ParseFiles([]string{second, first})
	if got.Total != 4 || got.Passed != 1 || got.Failed != 2 || got.Skipped != 1 {
		t.Fatalf("ParseFiles() counts = total:%d passed:%d failed:%d skipped:%d", got.Total, got.Passed, got.Failed, got.Skipped)
	}
	if len(got.Failures) != 2 {
		t.Fatalf("ParseFiles() failures = %d, want 2", len(got.Failures))
	}
	if got.Failures[1].Location != "WidgetTest.kt:42" {
		t.Fatalf("failure location = %q, want WidgetTest.kt:42", got.Failures[1].Location)
	}
	if !strings.Contains(got.Failures[1].Detail, "AssertionError") {
		t.Fatalf("failure detail = %q, want stack trace", got.Failures[1].Detail)
	}
}

func TestParseFilesReportsMalformedXML(t *testing.T) {
	t.Parallel()

	path := writeResult(t, t.TempDir(), "module", `<testsuite><testcase>`)
	got := ParseFiles([]string{path})
	if got.Files != 1 || len(got.ParseErrors) != 1 {
		t.Fatalf("ParseFiles() = files:%d errors:%d, want 1 and 1", got.Files, len(got.ParseErrors))
	}
}

func TestChangedFilesExcludesStaleReports(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	stale := writeResult(t, root, "stale", `<testsuite><testcase name="old"/></testsuite>`)
	changed := writeResult(t, root, "changed", `<testsuite><testcase name="before"/></testsuite>`)
	before, err := TakeSnapshot(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(changed, []byte(`<testsuite><testcase name="after"/></testsuite>`), 0o644); err != nil {
		t.Fatal(err)
	}
	fresh := writeResult(t, root, "fresh", `<testsuite><testcase name="new"/></testsuite>`)

	files, err := DiscoverFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ChangedFiles(before, files)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != changed || got[1] != fresh {
		t.Fatalf("ChangedFiles() = %q, want [%q %q]; stale was %q", got, changed, fresh, stale)
	}
}

func writeResult(t *testing.T, root, module, content string) string {
	t.Helper()
	directory := filepath.Join(root, module, "build", "test-results", "test")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "TEST-result.xml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
