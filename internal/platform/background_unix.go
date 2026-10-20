//go:build !windows

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// StartBackground starts a detached copy of the executable.
func StartBackground(executable string, args []string) (int, error) {
	null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return 0, fmt.Errorf("open null device: %w", err)
	}
	defer null.Close()

	cmd := exec.Command(executable, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = null, null, null
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start background process: %w", err)
	}
	pid := cmd.Process.Pid
	if err := cmd.Process.Release(); err != nil {
		return 0, fmt.Errorf("release background process: %w", err)
	}

	return pid, nil
}
