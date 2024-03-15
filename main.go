package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-toast/toast"
)

type Meeting struct {
	Topic string   `json:"topic"`
	Link  string   `json:"link"`
	When  string   `json:"when"`
	Days  []string `json:"days"`
}

// loadMeetings loads meetings from the given JSON file
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

// Sends a notification with the specified meeting topic and URL,
// allowing users to directly open the meeting from the notification
func notifyMeeting(topic string, url string) error {
	notification := toast.Notification{
		AppID:    "gomeet",
		Title:    "Join Meeting: " + topic,
		Message:  "Click to join the meeting now.",
		Actions:  []toast.Action{{Type: "protocol", Label: "Join", Arguments: url}},
		Duration: toast.Long,
	}

	return notification.Push()
}

// openMeetingLink opens the meeting link in the default web browser
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

func main() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Error finding config directory: %v\n", err)
	}

	meetings, err := loadMeetings(filepath.Join(configDir, "meetings.json"))
	if err != nil {
		log.Fatalln(err)
	}

	for {
		now := time.Now()
		for _, meeting := range meetings {
			fmt.Printf("%s when %s\n", meeting.Topic, meeting.When)

			// Continue if today is not a working day
			isWorkingDay := func() bool {
				for _, weekday := range meeting.Days {
					if weekday == now.Weekday().String() {
						return true
					}
				}
				return false
			}()

			if !isWorkingDay {
				continue
			}

			meetingTime, err := time.Parse("15:04", meeting.When)
			if err != nil {
				fmt.Printf("Error parsing time: %v\n", err)
				continue
			}

			// If the current time matches the [when], attempt to start the meeting
			if now.Hour() == when.Hour() && now.Minute() == when.Minute() {
				if err := notifyMeeting(meeting.Topic, meeting.Url); err != nil {
					fmt.Println(err)
				}

				// Retry opening the link if failed
				for i := 0; i < 100; i++ {
					if err := openMeetingLink(meeting.Link); err != nil {
						fmt.Printf("Error opening link: %v\n", err)
						continue
					}
					break
				}

				// Wait for a minute to avoid starting the meeting more than once in the same minute
				time.Sleep(time.Minute)
			}
		}

		// Sleep for a while but wake up before a minute has passed
		time.Sleep(20 * time.Second)
	}
}
