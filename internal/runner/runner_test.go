//go:build !windows

package runner

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunCapturesOutputAndExitCode(t *testing.T) {
	t.Parallel()

	script := writeScript(t, "#!/bin/sh\nprintf stdout\nprintf stderr >&2\nexit 7\n")
	var output bytes.Buffer
	got := Run(context.Background(), Command{Executable: script, Directory: filepath.Dir(script)}, &output)
	if !got.HasExitCode || got.ExitCode != 7 {
		t.Fatalf("Run() exit = %d (%v), want 7", got.ExitCode, got.HasExitCode)
	}
	if string(got.Tail) != output.String() {
		t.Fatalf("Run() tail = %q, output = %q", got.Tail, output.String())
	}
	if !strings.Contains(output.String(), "stdout") || !strings.Contains(output.String(), "stderr") {
		t.Fatalf("Run() output = %q, want combined streams", output.String())
	}
}

func TestRunTerminatesProcessGroupOnCancellation(t *testing.T) {
	t.Parallel()

	script := writeScript(t, "#!/bin/sh\ntrap 'exit 0' TERM\nwhile :; do sleep 1; done\n")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	got := Run(ctx, Command{Executable: script, Directory: filepath.Dir(script)}, &bytes.Buffer{})
	if !got.Interrupted {
		t.Fatal("Run() Interrupted = false, want true")
	}
}

func writeScript(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "script")
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}
