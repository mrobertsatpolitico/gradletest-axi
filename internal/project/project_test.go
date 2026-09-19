package project

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDiscoverFindsNearestWrapper(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	wrapper := filepath.Join(root, wrapperName())
	writeExecutable(t, wrapper)
	nested := filepath.Join(root, "module", "src")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Discover(nested)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if got.Root != root {
		t.Fatalf("Discover().Root = %q, want %q", got.Root, root)
	}
	if got.Wrapper != wrapper {
		t.Fatalf("Discover().Wrapper = %q, want %q", got.Wrapper, wrapper)
	}
}

func TestDiscoverRejectsMissingWrapper(t *testing.T) {
	t.Parallel()

	if _, err := Discover(t.TempDir()); err == nil {
		t.Fatal("Discover() error = nil, want missing-wrapper error")
	}
}

func TestDiscoverRejectsNonExecutableWrapper(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not use executable mode bits")
	}
	t.Parallel()

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "gradlew"), []byte("#!/bin/sh\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Discover(root); err == nil {
		t.Fatal("Discover() error = nil, want non-executable error")
	}
}

func wrapperName() string {
	if runtime.GOOS == "windows" {
		return "gradlew.bat"
	}
	return "gradlew"
}

func writeExecutable(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("wrapper"), 0o755); err != nil {
		t.Fatal(err)
	}
}
