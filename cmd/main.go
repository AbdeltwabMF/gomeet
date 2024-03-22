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

func matchMeeting(meeting Meeting) bool {
	now := time.Now()

	for _, wDay := range meeting.Days {
		if wDay == now.Weekday().String() {
			when, err := time.Parse("15:04", meeting.When)
			if err != nil {
				slog.Error(err.Error())
				return false
			}

			// If the current time matches the [when], attempt to start the meeting
			slog.Info(fmt.Sprintf("Matching %s", meeting.Topic), slog.String("parsed_time", when.String()))
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
			if matchMeeting(meeting) {
				if err := platform.NotifyMeeting(meeting.Topic, meeting.Url); err != nil {
					slog.Error(err.Error())
				}

				if config.AutoStart {
					if err := platform.OpenURL(meeting.Url); err != nil {
						slog.Error(err.Error())
					} else {
						slog.Info(fmt.Sprintf("%s started", meeting.Topic), slog.String("at", time.Now().String()))
					}
				}
			}
		}

		// Wait until the next minute to start the next iteration of meetings
		// This ensures that we are checking meetings at the beginning of each minute
		st := time.Minute - time.Duration(time.Now().Second())*time.Second
		slog.Info("Sleeping until the next minute", slog.String("sleep_time", st.String()))
		time.Sleep(st)
	}
}
