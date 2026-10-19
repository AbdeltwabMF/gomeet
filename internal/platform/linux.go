//go:build linux
// +build linux

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

	return exec.CommandContext(ctx, "notify-send", "--app-name", ToolName, "--",
		"Join Meeting: "+summary, "Click to join: "+url).Run()
}

// OpenURL opens the provided URL in the default web browser.
func OpenURL(ctx context.Context, url string) error {
	if err := ValidateURL(url); err != nil {
		return err
	}

	return exec.CommandContext(ctx, "xdg-open", url).Run()
}

// LogDir returns the path to the log directory.
func LogDir() (path string, err error) {
	d, err := os.UserCacheDir()
	if err != nil {
		return "", err
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
