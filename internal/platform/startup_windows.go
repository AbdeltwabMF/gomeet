//go:build windows

package platform

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const installStartupScript = `$ErrorActionPreference = 'Stop'
$user = [System.Security.Principal.WindowsIdentity]::GetCurrent().Name
$wscript = Join-Path $env:WINDIR 'System32\wscript.exe'
$arguments = '//B //NoLogo "' + $env:GOMEET_LAUNCHER + '" "' + $env:GOMEET_EXECUTABLE + '"'
$action = New-ScheduledTaskAction -Execute $wscript -Argument $arguments
$trigger = New-ScheduledTaskTrigger -AtLogOn -User $user
$trigger.Delay = 'PT30S'
$principal = New-ScheduledTaskPrincipal -UserId $user -LogonType Interactive -RunLevel Limited
$settings = New-ScheduledTaskSettingsSet -MultipleInstances IgnoreNew -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1) -ExecutionTimeLimit ([TimeSpan]::Zero) -StartWhenAvailable -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries
Register-ScheduledTask -TaskName 'GoMeet' -Description 'Start GoMeet after user sign-in' -Action $action -Trigger $trigger -Principal $principal -Settings $settings -Force | Out-Null
`

const startupLauncherFile = "startup.vbs"

// Windows Terminal can ignore PowerShell's -WindowStyle Hidden. WScript starts
// the console application with SW_HIDE, waits for it, and returns its exit code
// so Task Scheduler can still apply the configured restart policy.
const startupLauncherScript = `Option Explicit
Dim shell, executable, exitCode
executable = WScript.Arguments(0)
Set shell = CreateObject("WScript.Shell")
exitCode = shell.Run(Chr(34) & executable & Chr(34), 0, True)
WScript.Quit exitCode
`

const uninstallStartupScript = `$ErrorActionPreference = 'Stop'
$task = Get-ScheduledTask -TaskName 'GoMeet' -ErrorAction SilentlyContinue
if ($null -ne $task) {
    Unregister-ScheduledTask -TaskName 'GoMeet' -Confirm:$false
}
`

// SupportsStartup reports whether automatic startup can be managed here.
func SupportsStartup() bool { return true }

// InstallStartup installs a per-user Windows scheduled task.
func InstallStartup(ctx context.Context, executable string) error {
	launcher, err := startupLauncherPath()
	if err != nil {
		return fmt.Errorf("locate startup launcher: %w", err)
	}
	if err := writeStartupLauncher(launcher); err != nil {
		return err
	}

	return runStartupPowerShell(ctx, installStartupScript, launcher, executable)
}

// UninstallStartup removes the Windows scheduled task when present.
func UninstallStartup(ctx context.Context) error {
	if err := runStartupPowerShell(ctx, uninstallStartupScript, "", ""); err != nil {
		return err
	}
	launcher, err := startupLauncherPath()
	if err != nil {
		return fmt.Errorf("locate startup launcher: %w", err)
	}
	if err := os.Remove(launcher); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove startup launcher: %w", err)
	}

	return nil
}

func startupLauncherPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, startupLauncherFile), nil
}

func writeStartupLauncher(path string) error {
	if err := os.WriteFile(path, []byte(startupLauncherScript), 0600); err != nil {
		return fmt.Errorf("write startup launcher: %w", err)
	}
	return nil
}

func runStartupPowerShell(ctx context.Context, script, launcher, executable string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.Env = os.Environ()
	if launcher != "" {
		cmd.Env = append(cmd.Env, "GOMEET_LAUNCHER="+launcher)
	}
	if executable != "" {
		cmd.Env = append(cmd.Env, "GOMEET_EXECUTABLE="+executable)
	}
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("manage scheduled task: timed out")
	}
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return fmt.Errorf("manage scheduled task: %w", err)
	}

	return fmt.Errorf("manage scheduled task: %w: %s", err, detail)
}
