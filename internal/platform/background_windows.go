//go:build windows

package platform

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// StartBackground starts a detached copy of the executable.
func StartBackground(executable string, args []string) (pid int, retErr error) {
	null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return 0, fmt.Errorf("open null device: %w", err)
	}
	defer func() {
		if err := null.Close(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("close null device: %w", err))
		}
	}()

	cmd := exec.Command(executable, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = null, null, null
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS,
		HideWindow:    true,
	}
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start background process: %w", err)
	}
	pid = cmd.Process.Pid
	if err := cmd.Process.Release(); err != nil {
		return 0, fmt.Errorf("release background process: %w", err)
	}

	return pid, nil
}
