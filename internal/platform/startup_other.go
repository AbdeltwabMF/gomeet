//go:build !windows

package platform

import (
	"context"
	"errors"
)

var errStartupUnsupported = errors.New("automatic startup installation is currently supported only on Windows")

// SupportsStartup reports whether automatic startup can be managed here.
func SupportsStartup() bool { return false }

// InstallStartup reports that startup management is unsupported.
func InstallStartup(context.Context, string, bool) error { return errStartupUnsupported }

// UninstallStartup reports that startup management is unsupported.
func UninstallStartup(context.Context) error { return errStartupUnsupported }
