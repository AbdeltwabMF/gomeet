//go:build darwin
// +build darwin

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

// Notify sends a meeting notification with the specified summary and URL.
func Notify(summary string, url string) error {
	return ErrNotImpl
}

// OpenURL opens the specified URL in the default web browser.
func OpenURL(url string) error {
	cmd := exec.Command("open", url)
	return cmd.Run()
}

// LogDir returns the directory path for storing logs related to the tool.
func LogDir() (string, error) {
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	d := filepath.Join(h, "Library", "Logs", ToolName)
	if err := os.MkdirAll(d, 0750); err != nil {
		return "", err
	}

	return d, nil
}

// ConfigDir returns the directory path for storing configuration files related to the tool.
func ConfigDir() (string, error) {
	c, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	d := filepath.Join(c, ToolName)
	if err := os.MkdirAll(d, 0750); err != nil {
		return "", err
	}

	return d, nil
}
