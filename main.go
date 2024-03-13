//go:build prod

// main is a special name declaring an executable rather than a library
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gen2brain/beeep"
)

type Meeting struct {
	Topic string `json:"topic"`
	Link  string `json:"link"`
	When  string `json:"when"`
}

// loadMeetings loads meetings from the given JSON file.
func loadMeetings(filePath string) ([]Meeting, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening JSON file: %w", err)
	}
	defer file.Close()

	var meetings []Meeting
	if err := json.NewDecoder(file).Decode(&meetings); err != nil {
		return nil, fmt.Errorf("error decoding JSON file: %w", err)
	}

	return meetings, nil
}

// notifyMeeting sends a notification for the given meeting.
func notifyMeeting(meeting Meeting) error {
	err := beeep.Notify("GoMeet", "Your meeting link is set to launch shortly", "assets/information.png")
	if err != nil {
		return fmt.Errorf("error sending notification: %w", err)
	}
	return nil
}

// openMeetingLink opens the meeting link in the default web browser.
func openMeetingLink(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// main is the entry point for the executable.
func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Println("Error finding config directory:", err)
		return
	}

	meetings, err := loadMeetings(filepath.Join(configDir, "meetings.json"))
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		now := time.Now()
		hour, minute := now.Hour(), now.Minute()

		for _, meeting := range meetings {
			fmt.Println(meeting.Topic)

			mt, err := time.Parse("15:04", meeting.When)
			if err != nil {
				fmt.Println("Error parsing time:", err)
				continue
			}

			mh, mm := mt.Hour(), mt.Minute()
			if hour == mh && minute == mm {
				if err := notifyMeeting(meeting); err != nil {
					fmt.Println(err)
				}

				// Retry opening the link if failed
				for {
					if err := openMeetingLink(meeting.Link); err != nil {
						fmt.Println("Error opening link:", err)
						continue
					}
					break
				}

				// Wait before checking the next meeting
				time.Sleep(time.Minute)
			}
		}

		// Sleep until the next check time
		time.Sleep(20 * time.Second)
	}
}
