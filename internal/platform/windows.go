//go:build windows
// +build windows

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-toast/toast"
)

const ToolName = "gomeet"
const LocalDirKey = "LOCALAPPDATA"

// Notify displays a notification with the given summary and URL.
func Notify(summary string, url string) error {
	n := toast.Notification{
		AppID:    ToolName,
		Title:    "Join Meeting: " + summary,
		Message:  "Click to join the meeting now.",
		Actions:  []toast.Action{{Type: "protocol", Label: "Join", Arguments: url}},
		Duration: toast.Long,
	}

	return n.Push()
}

// OpenURL opens the provided URL in the default web browser.
func OpenURL(url string) error {
	cmd := exec.Command("cmd", "/c", "start", url)
	return cmd.Run()
}

// LogDir returns the path to the log directory.
func LogDir() (path string, err error) {
	d := os.Getenv(LocalDirKey)
	if d == "" {
		return "", fmt.Errorf("'%s' environment variable is not set", LocalDirKey)
	}

	d = filepath.Join(d, ToolName, "logs")
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
