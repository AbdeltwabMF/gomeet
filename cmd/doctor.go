package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/googlecal"
)

type commandRequirement struct {
	name    string
	purpose string
	install string
}

type doctorWriter struct {
	w   io.Writer
	err error
}

func (w *doctorWriter) print(args ...any) {
	if w.err == nil {
		_, w.err = fmt.Fprint(w.w, args...)
	}
}

func (w *doctorWriter) printf(format string, args ...any) {
	if w.err == nil {
		_, w.err = fmt.Fprintf(w.w, format, args...)
	}
}

func (w *doctorWriter) println(args ...any) {
	if w.err == nil {
		_, w.err = fmt.Fprintln(w.w, args...)
	}
}

func runDoctor(w io.Writer) error {
	report := &doctorWriter{w: w}
	report.println("GoMeet doctor")
	report.println("=============")
	report.println("Local and Google calendars are independent; you can use either or both.")
	report.println("The auto_start setting applies to meetings from both sources.")
	problems := 0
	localEventCount := -1

	report.println("\nConfiguration")
	report.println("-------------")
	configPath, created, err := configs.EnsureConfig()
	if err != nil {
		report.printf("[FAIL] Configuration: %v\n", err)
		problems++
	} else {
		if created {
			report.printf("[WARN] Configuration was missing, so an empty template was created at:\n       %s\n", configPath)
		} else {
			report.printf("[OK]   Configuration file: %s\n", configPath)
		}
		cfg, loadErr := configs.LoadConfig()
		if loadErr != nil {
			report.printf("[FAIL] Configuration is invalid: %v\n", loadErr)
			problems++
		} else {
			localEventCount = len(cfg.Events)
			report.println("[OK]   Configuration is valid.")
		}
	}

	report.println("\nLocal Calendar")
	report.println("--------------")
	if localEventCount < 0 {
		report.println("[FAIL] Local events could not be checked because the configuration is invalid.")
	} else if localEventCount == 0 {
		report.println("[INFO] Local Calendar is disabled because no local events are configured.")
	} else {
		report.printf("[OK]   Local Calendar has %d configured event(s).\n", localEventCount)
	}
	printLocalConfigurationHelp(report)

	report.println("\nGoogle Calendar (optional)")
	report.println("--------------------------")
	credentialsPath, err := configs.CredentialsPath()
	if err != nil {
		report.printf("[WARN] Cannot determine the optional Google credentials path: %v\n", err)
	} else if info, statErr := os.Stat(credentialsPath); statErr == nil && !info.IsDir() {
		if credentialsErr := googlecal.CheckCredentials(); credentialsErr != nil {
			report.printf("[FAIL] Google Calendar credentials are invalid: %v\n", credentialsErr)
			printGoogleCredentialsHelp(report, credentialsPath)
			problems++
		} else {
			report.printf("[OK]   Google Calendar credentials: %s\n", credentialsPath)
		}
	} else {
		report.println("[INFO] Google Calendar is disabled.")
		printGoogleCredentialsHelp(report, credentialsPath)
	}

	report.println("\nSystem")
	report.println("------")
	logPath, err := configs.LogPath()
	if err != nil {
		report.printf("[FAIL] Log directory: %v\n", err)
		problems++
	} else {
		f, openErr := configs.OpenLog()
		if openErr != nil {
			report.printf("[FAIL] Cannot write logs at %s: %v\n", logPath, openErr)
			problems++
		} else {
			if closeErr := f.Close(); closeErr != nil {
				report.printf("[FAIL] Cannot close log file at %s: %v\n", logPath, closeErr)
				problems++
			} else {
				report.printf("[OK]   Log file is writable: %s\n", logPath)
			}
		}
	}

	for _, requirement := range commandRequirements() {
		path, lookupErr := exec.LookPath(requirement.name)
		if lookupErr != nil {
			report.printf("[FAIL] %s is missing (%s). %s\n", requirement.name, requirement.purpose, requirement.install)
			problems++
		} else {
			report.printf("[OK]   %s: %s\n", requirement.name, path)
		}
	}

	if report.err != nil {
		return fmt.Errorf("write doctor report: %w", report.err)
	}
	if problems != 0 {
		return fmt.Errorf("doctor found %d problem(s); follow the [FAIL] instructions above", problems)
	}
	report.println("\nNo problems found.")
	if report.err != nil {
		return fmt.Errorf("write doctor report: %w", report.err)
	}
	return nil
}

func printLocalConfigurationHelp(w *doctorWriter) {
	w.print(`
Configuration example:
{
  "auto_start": false,
  "events": [
    {
      "summary": "Standup",
      "url": "https://meet.example.com/room",
      "start": {
        "time": "09:30",
        "days": ["Monday"]
      }
    }
  ]
}

Field guide:
  auto_start  Open local and Google meeting links automatically when they start.
  summary     Name displayed in the notification.
  url         An http:// or https:// meeting link.
  time        Local start time in 24-hour HH:MM format.
  days        Full English weekday names, such as Monday or Wednesday.
`)
}

func printGoogleCredentialsHelp(w *doctorWriter, credentialsPath string) {
	w.printf(`
       To enable Google Calendar:
       1. Create or select a Google Cloud project:
          https://console.cloud.google.com/projectcreate
       2. Enable the Google Calendar API:
          https://console.cloud.google.com/apis/library/calendar-json.googleapis.com
       3. Configure the OAuth consent screen. If the app is in testing mode,
          add your Google account as a test user.
       4. Create an OAuth client ID with application type "Desktop app":
          https://console.cloud.google.com/apis/credentials
       5. Download its JSON file, rename it to credentials.json, and save it as:
          %s

       Official guide:
       https://developers.google.com/workspace/guides/create-credentials#desktop-app
`, credentialsPath)
}

func commandRequirements() []commandRequirement {
	switch runtime.GOOS {
	case "linux":
		return []commandRequirement{
			{name: "notify-send", purpose: "desktop notifications", install: "Install libnotify (for Debian/Ubuntu: sudo apt install libnotify-bin)."},
			{name: "xdg-open", purpose: "opening meeting links", install: "Install xdg-utils (for Debian/Ubuntu: sudo apt install xdg-utils)."},
		}
	case "darwin":
		return []commandRequirement{
			{name: "osascript", purpose: "desktop notifications", install: "Restore the macOS system command or reinstall macOS command-line components."},
			{name: "open", purpose: "opening meeting links", install: "Restore the macOS system command or reinstall macOS command-line components."},
		}
	default:
		return nil
	}
}
