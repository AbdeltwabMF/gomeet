//go:build darwin
// +build darwin

package platform

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

var ErrNotImpl = errors.New("not implemented on Darwin platform")

// NotifyMeeting sends actionable notification with the specified meeting topic and URL
func NotifyMeeting(topic, url string) error {
	return ErrNotImpl
}

// OpenURL opens the specified URL in the default web browser
func OpenURL(url string) error {
	cmd := exec.Command("open", url)
	return cmd.Run()
}

// LogDir returns the application-specific log directory
func LogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	logDir := filepath.Join(home, "Library", "Logs", "gomeet")
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return "", err
	}

	return logDir, nil
}
