package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/AbdeltwabMF/gomeet/platform"
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

func loadConfig() (*Config, error) {
	cDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	gomeetDir := filepath.Join(cDir, "gomeet")
	if err := os.MkdirAll(gomeetDir, 0750); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(gomeetDir, "config.json"), os.O_CREATE|os.O_RDONLY, 0640)
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

func initLogger() (*os.File, error) {
	lDir, err := platform.LogDir()
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(lDir, "log.txt"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewJSONHandler(file, nil))
	slog.SetDefault(logger)

	return file, nil
}

func meetingMatch(meeting Meeting) bool {
	now := time.Now()

	for _, wDay := range meeting.Days {
		if wDay == now.Weekday().String() {
			when, err := time.Parse("15:04", meeting.When)
			if err != nil {
				slog.Error(err.Error())
				return false
			}

			// If the current time matches the [when], attempt to start the meeting
			if now.Hour() == when.Hour() && now.Minute() == when.Minute() {
				return true
			} else {
				return false
			}
		}
	}

	return false
}

func main() {
	logFile, err := initLogger()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer logFile.Close()

	config, err := loadConfig()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	for {
		for _, meeting := range config.Meetings {
			if meetingMatch(meeting) {
				if err := platform.NotifyMeeting(meeting.Topic, meeting.Url); err != nil {
					slog.Error(err.Error())
				}

				if config.AutoStart {
					if err := platform.OpenURL(meeting.Url); err != nil {
						slog.Error(err.Error())
					}
				}
			}
		}

		// Wait till the next minute
		time.Sleep(time.Minute - time.Duration(time.Now().Second()))
	}
}
