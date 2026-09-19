//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package runner

import (
	"os"
	"os/exec"
	"syscall"
)

func configureProcess(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func terminateProcess(process *os.Process) error {
	return syscall.Kill(-process.Pid, syscall.SIGTERM)
}
func killProcess(process *os.Process) error {
	return syscall.Kill(-process.Pid, syscall.SIGKILL)
}

func processExitCode(state *os.ProcessState, waitError error, interrupted bool) (int, bool) {
	if status, ok := state.Sys().(syscall.WaitStatus); ok {
		if status.Signaled() {
			return 128 + int(status.Signal()), true
		}
		return status.ExitStatus(), true
	}
	code := state.ExitCode()
	return code, code >= 0
}
