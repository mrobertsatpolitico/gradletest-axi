//go:build !windows

package app

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	toon "github.com/toon-format/toon-go"
)

func TestApplicationPassesFixedTaskArgumentsAndIsolatesGradleNoise(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeGradlew(t, root, `#!/bin/sh
printf '%s\n' "$@" > args.txt
printf 'raw stdout noise\n'
printf 'raw stderr noise\n' >&2
mkdir -p build/test-results/test
cat > build/test-results/test/TEST-pass.xml <<'XML'
<testsuite><testcase classname="WidgetTest" name="passes" time="0.25"/></testsuite>
XML
`)
	stdout, stderr, exitCode := execute(t, root, []string{"--", "--tests", "WidgetTest"})
	if exitCode != 0 {
		t.Fatalf("Execute() exit = %d, output:\n%s", exitCode, stdout)
	}
	document := decodeDocument(t, stdout)
	assertField(t, document, "status", "passed")
	assertField(t, document, "kind", "test")
	assertField(t, document, "gradle_exit", "0")
	if strings.Contains(stdout, "raw stdout noise") || strings.Contains(stdout, "raw stderr noise") {
		t.Fatalf("structured stdout leaked Gradle noise:\n%s", stdout)
	}
	if !strings.Contains(stderr, "running Gradle tests") {
		t.Fatalf("stderr = %q, want progress message", stderr)
	}
	args, err := os.ReadFile(filepath.Join(root, "args.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(args) != "test\n--console=plain\n--tests\nWidgetTest\n" {
		t.Fatalf("Gradle args = %q", args)
	}
}

func TestApplicationDoesNotReuseStaleJUnitOnBuildFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeJUnit(t, root, `<testsuite><testcase classname="OldTest" name="fails"><failure message="stale failure"/></testcase></testsuite>`)
	writeGradlew(t, root, `#!/bin/sh
printf '* What went wrong:\nCompilation failed\n\n* Try:\n--stacktrace\n'
exit 1
`)
	stdout, _, exitCode := execute(t, root, nil)
	if exitCode != 1 {
		t.Fatalf("Execute() exit = %d, output:\n%s", exitCode, stdout)
	}
	document := decodeDocument(t, stdout)
	assertField(t, document, "kind", "build")
	assertField(t, document, "report", "unavailable")
	if _, ok := document["failures"]; ok {
		t.Fatalf("stale failures present in output:\n%s", stdout)
	}
	if !strings.Contains(fmt.Sprint(document["error"]), "Compilation failed") {
		t.Fatalf("error = %v, want compilation summary", document["error"])
	}
	logPath := fmt.Sprint(document["log"])
	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log %q: %v", logPath, err)
	}
	if !strings.Contains(string(logContent), "Compilation failed") {
		t.Fatalf("log = %q, want complete Gradle output", logContent)
	}
}

func TestApplicationTrustsFreshJUnitFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeJUnit(t, root, `<testsuite><testcase classname="OldTest" name="passes"/></testsuite>`)
	writeGradlew(t, root, `#!/bin/sh
mkdir -p build/test-results/test
cat > build/test-results/test/TEST-result.xml <<'XML'
<testsuite><testcase classname="WidgetTest" name="fails"><failure message="expected true">at WidgetTest.fails(WidgetTest.kt:9)</failure></testcase></testsuite>
XML
exit 1
`)
	stdout, _, exitCode := execute(t, root, nil)
	if exitCode != 1 {
		t.Fatalf("Execute() exit = %d, output:\n%s", exitCode, stdout)
	}
	document := decodeDocument(t, stdout)
	assertField(t, document, "kind", "test")
	assertField(t, document, "report", "junit-fresh")
	if _, ok := document["log"]; ok {
		t.Fatalf("complete JUnit-backed failure unexpectedly advertised raw log:\n%s", stdout)
	}
}

func TestApplicationReturnsStructuredUsageError(t *testing.T) {
	t.Parallel()

	stdout, _, exitCode := execute(t, t.TempDir(), []string{"--wat"})
	if exitCode != 2 {
		t.Fatalf("Execute() exit = %d, want 2", exitCode)
	}
	document := decodeDocument(t, stdout)
	assertField(t, document, "exit_code", "2")
	if !strings.Contains(fmt.Sprint(document["error"]), "unknown argument") {
		t.Fatalf("error = %v", document["error"])
	}
}

func TestApplicationReturnsStructuredMissingWrapperError(t *testing.T) {
	t.Parallel()

	stdout, _, exitCode := execute(t, t.TempDir(), nil)
	if exitCode != 1 {
		t.Fatalf("Execute() exit = %d, want 1", exitCode)
	}
	document := decodeDocument(t, stdout)
	if !strings.Contains(fmt.Sprint(document["error"]), "no Gradle wrapper found") {
		t.Fatalf("error = %v", document["error"])
	}
}

func TestApplicationReportsWrapperStartFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeGradlew(t, root, "#!/definitely/missing/interpreter\n")
	stdout, _, exitCode := execute(t, root, nil)
	if exitCode != 1 {
		t.Fatalf("Execute() exit = %d, want 1", exitCode)
	}
	document := decodeDocument(t, stdout)
	assertField(t, document, "kind", "wrapper")
	if _, ok := document["gradle_exit"]; ok {
		t.Fatalf("start failure unexpectedly has gradle_exit:\n%s", stdout)
	}
}

func TestApplicationMakesMalformedReportExplicit(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeGradlew(t, root, `#!/bin/sh
mkdir -p build/test-results/test
printf '<testsuite><testcase>' > build/test-results/test/TEST-bad.xml
`)
	stdout, _, exitCode := execute(t, root, nil)
	if exitCode != 0 {
		t.Fatalf("Execute() exit = %d, output:\n%s", exitCode, stdout)
	}
	document := decodeDocument(t, stdout)
	assertField(t, document, "status", "passed")
	assertField(t, document, "report", "malformed")
	if !strings.Contains(fmt.Sprint(document["warning"]), "could not be parsed") {
		t.Fatalf("warning = %v", document["warning"])
	}
}

func TestApplicationReportsZeroTestsExplicitly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeJUnit(t, root, `<testsuite/>`)
	writeGradlew(t, root, "#!/bin/sh\nexit 0\n")
	stdout, _, exitCode := execute(t, root, nil)
	if exitCode != 0 {
		t.Fatalf("Execute() exit = %d, output:\n%s", exitCode, stdout)
	}
	document := decodeDocument(t, stdout)
	tests, ok := document["tests"].(map[string]any)
	if !ok {
		t.Fatalf("tests = %T, want map", document["tests"])
	}
	assertField(t, tests, "total", "0")
	if !strings.Contains(stdout, "0 tests were reported") {
		t.Fatalf("output lacks explicit zero-test guidance:\n%s", stdout)
	}
}
func execute(t *testing.T, root string, args []string) (string, string, int) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	application := Application{
		Version:   "test",
		Stdout:    &stdout,
		Stderr:    &stderr,
		CacheBase: t.TempDir(),
	}
	exitCode := application.Execute(context.Background(), args, root)
	return stdout.String(), stderr.String(), exitCode
}

func writeGradlew(t *testing.T, root, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "gradlew"), []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeJUnit(t *testing.T, root, content string) {
	t.Helper()
	directory := filepath.Join(root, "build", "test-results", "test")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "TEST-result.xml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func decodeDocument(t *testing.T, value string) map[string]any {
	t.Helper()
	decoded, err := toon.DecodeString(value)
	if err != nil {
		t.Fatalf("decode TOON: %v\n%s", err, value)
	}
	document, ok := decoded.(map[string]any)
	if !ok {
		t.Fatalf("decoded type = %T, want map", decoded)
	}
	return document
}

func assertField(t *testing.T, document map[string]any, field, want string) {
	t.Helper()
	if got := fmt.Sprint(document[field]); got != want {
		t.Fatalf("%s = %q, want %q", field, got, want)
	}
}
