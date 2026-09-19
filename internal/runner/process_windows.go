//go:build windows

package runner

import (
	"os"
	"os/exec"
)

func configureProcess(command *exec.Cmd) {}

func terminateProcess(process *os.Process) error {
	return process.Kill()
}
func killProcess(process *os.Process) error {
	return process.Kill()
}

func processExitCode(state *os.ProcessState, waitError error, interrupted bool) (int, bool) {
	if interrupted {
		return 130, true
	}
	code := state.ExitCode()
	return code, code >= 0
}
