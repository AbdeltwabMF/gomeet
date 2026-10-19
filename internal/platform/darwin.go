//go:build darwin
// +build darwin

package platform

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// Notify displays a notification with the given summary and URL.
func Notify(ctx context.Context, summary string, url string) error {
	if err := ValidateURL(url); err != nil {
		return err
	}

	const script = `on run argv
display notification (item 2 of argv) with title ("Join Meeting: " & item 1 of argv)
end run`
	return exec.CommandContext(ctx, "osascript", "-e", script, "--", summary, "Click to join: "+url).Run()
}

// OpenURL opens the provided URL in the default web browser.
func OpenURL(ctx context.Context, url string) error {
	if err := ValidateURL(url); err != nil {
		return err
	}

	return exec.CommandContext(ctx, "open", url).Run()
}

// LogDir returns the path to the log directory.
func LogDir() (path string, err error) {
	d, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	d = filepath.Join(d, "Library", "Logs", ToolName)
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
