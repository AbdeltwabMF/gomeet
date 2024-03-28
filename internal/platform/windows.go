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

const (
	ToolName = "gomeet"
)

// Notify sends a meeting notification with the specified summary and URL.
func Notify(summary string, url string) error {
	notification := toast.Notification{
		AppID:    "gomeet",
		Title:    "Join Meeting: " + summary,
		Message:  "Click to join the meeting now.",
		Actions:  []toast.Action{{Type: "protocol", Label: "Join", Arguments: url}},
		Duration: toast.Long,
	}

	return notification.Push()
}

// OpenURL opens the specified URL in the default web browser.
func OpenURL(url string) error {
	cmd := exec.Command("cmd", "/c", "start", url)
	return cmd.Run()
}

// LogDir returns the directory path for storing logs related to the tool.
func LogDir() (string, error) {
	l := os.Getenv("LOCALAPPDATA")
	if l == "" {
		return "", fmt.Errorf("'LOCALAPPDATA' is not defined in the environment variables")
	}

	d := filepath.Join(l, ToolName, "logs")
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
