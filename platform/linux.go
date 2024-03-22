//go:build linux
// +build linux

package platform

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

var ErrNotImpl = errors.New("not implemented on Linux platform")

// NotifyMeeting sends actionable notification with the specified meeting topic and URL
func NotifyMeeting(string, string) error {
	return ErrNotImpl
}

// OpenURL opens the specified URL in the default web browser
func OpenURL(url string) error {
	cmd := exec.Command("xdg-open", url)
	return cmd.Run()
}

// LogDir returns the application-specific log directory
func LogDir() (string, error) {
	logDir := filepath.Join("/var/log", "gomeet")
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return "", err
	}

	return logDir, nil
}
