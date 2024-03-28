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

func NotifyMeeting(summary string, url string) error {
	notification := toast.Notification{
		AppID:    "gomeet",
		Title:    "Join Meeting: " + summary,
		Message:  "Click to join the meeting now.",
		Actions:  []toast.Action{{Type: "protocol", Label: "Join", Arguments: url}},
		Duration: toast.Long,
	}

	return notification.Push()
}

func OpenURL(url string) error {
	cmd := exec.Command("cmd", "/c", "start", url)
	return cmd.Run()
}

func LogDir() (string, error) {
	d := os.Getenv("LOCALAPPDATA")
	if d == "" {
		return "", fmt.Errorf("'LOCALAPPDATA' is not defined in the environment variables")
	}

	tDir := filepath.Join(d, ToolName, "logs")
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
