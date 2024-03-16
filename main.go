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
	Url   string   `json:"url"`
	When  string   `json:"when"`
	Days  []string `json:"days"`
}

type Config struct {
	AutoStart bool      `json:"auto_start"`
	Meetings  []Meeting `json:"meetings"`
}

// Loads configurations from the specified file path and returns a Config instance
func loadConfig(filePath string) (*Config, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Sends a notification with the specified meeting topic and URL,
// allowing users to directly open the meeting from the notification
func notifyMeeting(topic string, url string) error {
	switch runtime.GOOS {
	case "windows":
		notification := toast.Notification{
			AppID:    "gomeet",
			Title:    "Join Meeting: " + topic,
			Message:  "Click to join the meeting now.",
			Actions:  []toast.Action{{Type: "protocol", Label: "Join", Arguments: url}},
			Duration: toast.Long,
		}

		return notification.Push()
	case "darwin":
		return nil
	default:
		return nil
	}
}

// Opens the specified URL in the default web browser
func openUrl(url string) error {
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
		log.Fatal(err)
	}

	config, err := loadConfig(filepath.Join(configDir, "meetings.json"))
	if err != nil {
		log.Fatal(err)
	}

	for {
		now := time.Now()
		for _, meeting := range config.Meetings {
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

			when, err := time.Parse("15:04", meeting.When)
			if err != nil {
				fmt.Println(err)
				continue
			}

			// If the current time matches the [when], attempt to start the meeting
			if now.Hour() == when.Hour() && now.Minute() == when.Minute() {
				if err := notifyMeeting(meeting.Topic, meeting.Url); err != nil {
					fmt.Println(err)
				}

				if config.AutoStart {
					if err := openUrl(meeting.Url); err != nil {
						fmt.Println(err)
					}
				}

				// Prevents checking the meeting more than once in the same minute
				time.Sleep(time.Minute)
			}
		}

		// Sleep for a while but wake up before a minute has passed
		time.Sleep(20 * time.Second)
	}
}
