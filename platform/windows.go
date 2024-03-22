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

// NotifyMeeting sends actionable notification with the specified meeting topic and URL
func NotifyMeeting(topic string, url string) error {
	notification := toast.Notification{
		AppID:    "gomeet",
		Title:    "Join Meeting: " + topic,
		Message:  "Click to join the meeting now.",
		Actions:  []toast.Action{{Type: "protocol", Label: "Join", Arguments: url}},
		Duration: toast.Long,
	}

	return notification.Push()
}

// OpenURL opens the specified URL in the default web browser
func OpenURL(url string) error {
	cmd := exec.Command("cmd", "/c", "start", url)
	return cmd.Run()
}

// LogDir returns the application-specific log directory
func LogDir() (string, error) {
	appDataDir := os.Getenv("LOCALAPPDATA")
	if appDataDir == "" {
		return "", fmt.Errorf("'LOCALAPPDATA' is not defined in the environment variables")
	}

	logDir := filepath.Join(appDataDir, "gomeet", "logs")
	if err := os.MkdirAll(logDir, 0750); err != nil {
		return "", err
	}

	return logDir, nil
}
