package runner

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"sync"
	"time"
)

const defaultTailBytes = 64 * 1024

type Command struct {
	Executable string
	Args       []string
	Directory  string
}

type Result struct {
	ExitCode       int
	HasExitCode    bool
	Duration       time.Duration
	Tail           string
	Interrupted    bool
	StartError     error
	TerminateError error
}

type lockedWriter struct {
	mu      sync.Mutex
	writers []io.Writer
}

func (w *lockedWriter) Write(payload []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, writer := range w.writers {
		if _, err := writer.Write(payload); err != nil {
			return 0, err
		}
	}
	return len(payload), nil
}

func Run(ctx context.Context, command Command, log io.Writer) Result {
	startedAt := time.Now()
	tail := NewTailBuffer(defaultTailBytes)
	output := &lockedWriter{writers: []io.Writer{log, tail}}

	child := exec.Command(command.Executable, command.Args...)
	child.Dir = command.Directory
	child.Stdout = output
	child.Stderr = output
	configureProcess(child)

	if err := child.Start(); err != nil {
		return Result{
			Duration:   time.Since(startedAt),
			Tail:       tail.String(),
			StartError: err,
		}
	}

	done := make(chan error, 1)
	go func() {
		done <- child.Wait()
	}()

	var waitError error
	result := Result{}
	select {
	case waitError = <-done:
	case <-ctx.Done():
		result.Interrupted = true
		result.TerminateError = terminateProcess(child.Process)
		select {
		case waitError = <-done:
		case <-time.After(2 * time.Second):
			if killError := killProcess(child.Process); result.TerminateError == nil {
				result.TerminateError = killError
			}
			waitError = <-done
		}
	}

	result.Duration = time.Since(startedAt)
	result.Tail = tail.String()
	if child.ProcessState != nil {
		result.ExitCode, result.HasExitCode = processExitCode(child.ProcessState, waitError, result.Interrupted)
	}
	if waitError != nil && !result.HasExitCode && !errors.Is(waitError, context.Canceled) {
		result.StartError = waitError
	}
	return result
}
