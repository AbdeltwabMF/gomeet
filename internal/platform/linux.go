//go:build linux
// +build linux

package platform

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	ToolName = "gomeet"
)

var (
	ErrNotImpl = errors.New("not implemented on Linux platform")
)

func NotifyMeeting(summary string, url string) error {
	return ErrNotImpl
}

func OpenURL(url string) error {
	cmd := exec.Command("xdg-open", url)
	return cmd.Run()
}

func LogDir() (string, error) {
	tDir := filepath.Join("/", "var", "log", ToolName)
	if err := os.MkdirAll(tDir, 0750); err != nil {
		return "", err
	}

	return tDir, nil
}

func ConfigDir() (string, error) {
	c, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	tDir := filepath.Join(c, ToolName)
	if err := os.MkdirAll(tDir, 0750); err != nil {
		return "", err
	}

	return tDir, nil
}
