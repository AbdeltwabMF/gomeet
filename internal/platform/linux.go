//go:build linux
// +build linux

package platform

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

const ToolName = "gomeet"

var ErrNotImplemented = errors.New("feature: not implemented on Linux platform")

// Notify displays a notification with the given summary and URL.
func Notify(summary string, url string) error {
	return ErrNotImplemented
}

// OpenURL opens the provided URL in the default web browser.
func OpenURL(url string) error {
	cmd := exec.Command("xdg-open", url)
	return cmd.Run()
}

// LogDir returns the path to the log directory.
func LogDir() (path string, err error) {
	d := filepath.Join("/", "var", "log", ToolName)
	if err := os.MkdirAll(d, 0750); err != nil {
		return "", err
	}

	return d, nil
}

// ConfigDir returns the path to the configuration directory.
func ConfigDir() (path string, err error) {
	d, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	d = filepath.Join(d, ToolName)
	if err := os.MkdirAll(d, 0750); err != nil {
		return "", err
	}

	return d, nil
}
