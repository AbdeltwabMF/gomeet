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

func runDoctor(w io.Writer) error {
	fmt.Fprintln(w, "GoMeet doctor")
	fmt.Fprintln(w, "=============")
	fmt.Fprintln(w, "Local and Google calendars are independent; you can use either or both.")
	fmt.Fprintln(w, "The auto_start setting applies to meetings from both sources.")
	problems := 0
	localEventCount := -1

	fmt.Fprintln(w, "\nConfiguration")
	fmt.Fprintln(w, "-------------")
	configPath, created, err := configs.EnsureConfig()
	if err != nil {
		fmt.Fprintf(w, "[FAIL] Configuration: %v\n", err)
		problems++
	} else {
		if created {
			fmt.Fprintf(w, "[WARN] Configuration was missing, so an empty template was created at:\n       %s\n", configPath)
		} else {
			fmt.Fprintf(w, "[OK]   Configuration file: %s\n", configPath)
		}
		cfg, loadErr := configs.LoadConfig()
		if loadErr != nil {
			fmt.Fprintf(w, "[FAIL] Configuration is invalid: %v\n", loadErr)
			problems++
		} else {
			localEventCount = len(cfg.Events)
			fmt.Fprintln(w, "[OK]   Configuration is valid.")
		}
	}

	fmt.Fprintln(w, "\nLocal Calendar")
	fmt.Fprintln(w, "--------------")
	if localEventCount < 0 {
		fmt.Fprintln(w, "[FAIL] Local events could not be checked because the configuration is invalid.")
	} else if localEventCount == 0 {
		fmt.Fprintln(w, "[INFO] Local Calendar is disabled because no local events are configured.")
	} else {
		fmt.Fprintf(w, "[OK]   Local Calendar has %d configured event(s).\n", localEventCount)
	}
	printLocalConfigurationHelp(w)

	fmt.Fprintln(w, "\nGoogle Calendar (optional)")
	fmt.Fprintln(w, "--------------------------")
	credentialsPath, err := configs.CredentialsPath()
	if err != nil {
		fmt.Fprintf(w, "[WARN] Cannot determine the optional Google credentials path: %v\n", err)
	} else if info, statErr := os.Stat(credentialsPath); statErr == nil && !info.IsDir() {
		if credentialsErr := googlecal.CheckCredentials(); credentialsErr != nil {
			fmt.Fprintf(w, "[FAIL] Google Calendar credentials are invalid: %v\n", credentialsErr)
			printGoogleCredentialsHelp(w, credentialsPath)
			problems++
		} else {
			fmt.Fprintf(w, "[OK]   Google Calendar credentials: %s\n", credentialsPath)
		}
	} else {
		fmt.Fprintln(w, "[INFO] Google Calendar is disabled.")
		printGoogleCredentialsHelp(w, credentialsPath)
	}

	fmt.Fprintln(w, "\nSystem")
	fmt.Fprintln(w, "------")
	logPath, err := configs.LogPath()
	if err != nil {
		fmt.Fprintf(w, "[FAIL] Log directory: %v\n", err)
		problems++
	} else {
		f, openErr := configs.OpenLog()
		if openErr != nil {
			fmt.Fprintf(w, "[FAIL] Cannot write logs at %s: %v\n", logPath, openErr)
			problems++
		} else {
			f.Close()
			fmt.Fprintf(w, "[OK]   Log file is writable: %s\n", logPath)
		}
	}

	for _, requirement := range commandRequirements() {
		path, lookupErr := exec.LookPath(requirement.name)
		if lookupErr != nil {
			fmt.Fprintf(w, "[FAIL] %s is missing (%s). %s\n", requirement.name, requirement.purpose, requirement.install)
			problems++
		} else {
			fmt.Fprintf(w, "[OK]   %s: %s\n", requirement.name, path)
		}
	}

	if problems != 0 {
		return fmt.Errorf("doctor found %d problem(s); follow the [FAIL] instructions above", problems)
	}
	fmt.Fprintln(w, "\nNo problems found.")
	return nil
}

func printLocalConfigurationHelp(w io.Writer) {
	fmt.Fprint(w, `
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

func printGoogleCredentialsHelp(w io.Writer, credentialsPath string) {
	fmt.Fprintf(w, `
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
