//go:build windows
// +build windows

package platform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-toast/toast"
	"golang.org/x/sys/windows"
)

const LocalDirKey = "LOCALAPPDATA"

// Notify displays a notification with the given summary and URL.
func Notify(ctx context.Context, summary string, url string) error {
	if err := ValidateURL(url); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

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
func OpenURL(ctx context.Context, url string) error {
	if err := ValidateURL(url); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	target, err := windows.UTF16PtrFromString(url)
	if err != nil {
		return err
	}

	return windows.ShellExecute(0, verb, target, nil, nil, windows.SW_SHOWNORMAL)
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
