package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/AbdeltwabMF/gomeet/internal/googlecal"
	"github.com/AbdeltwabMF/gomeet/internal/localcal"
	"github.com/AbdeltwabMF/gomeet/internal/platform"
	"google.golang.org/api/calendar/v3"
)

type Config struct {
	AutoStart bool `json:"auto_start"`
}

func loadConfig() (*Config, error) {
	c, err := platform.ConfigDir()
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filepath.Join(c, "config.json"), os.O_CREATE|os.O_RDONLY, 0640)
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

func waitNextMinute(calendarName string) {
	sd := time.Until(time.Now().Truncate(time.Minute).Add(time.Minute))

	slog.Info("Waiting for the next minute to recheck events",
		slog.String("time.sleep", sd.String()),
		slog.String("calendar", calendarName),
	)

	time.Sleep(sd)
}

func main() {
	file, err := initLogger()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintf(file, "\n---------- ---------- (%v) ---------- ----------\n", time.Now().Local().Format(time.RFC3339))
	defer file.Close()

	cfg, err := loadConfig()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	var gevents *calendar.Events
	var levents *localcal.Events

	gch := make(chan *calendar.Events)
	lch := make(chan *localcal.Events)
	gerrch := make(chan error)
	lerrch := make(chan error)

	retryLimit := 32
	go googlecal.Fetch(gch, gerrch, retryLimit)
	go localcal.Load(lch, lerrch, retryLimit)

	// Monitor Google calendar events.
	go func() {
	loop:
		for {
			select {
			case gevents = <-gch:
				slog.Info("Read from Google calendar channel")
			case err := <-gerrch:
				slog.Error(fmt.Sprintf("Unable to fetch events: %v", err), slog.String("calendar", "Google calendar"))
				break loop
			default:
				if gevents != nil {
					for _, item := range gevents.Items {
						matched, err := googlecal.Match(item)
						if err != nil {
							slog.Error(err.Error())
							continue
						}

						if matched {
							err := googlecal.Execute(item, cfg.AutoStart)
							if err != nil {
								slog.Error(err.Error())
							}
						}
					}
				}
				waitNextMinute("Google calendar")
			}
		}
	}()

	// Monitor local calendar events.
	go func() {
	loop:
		for {
			select {
			case levents = <-lch:
				slog.Info("Read from Local calendar channel")
			case err := <-lerrch:
				slog.Error(fmt.Sprintf("Unable to load events: %v", err), slog.String("calendar", "Local calendar"))
				break loop
			default:
				if levents != nil {
					for _, item := range levents.Items {
						if localcal.Match(item) {
							err := localcal.Execute(item, cfg.AutoStart)
							if err != nil {
								slog.Error(err.Error())
							}
						}
					}
				}
				waitNextMinute("Local calendar")
			}
		}
	}()

	// Block indefinitely to keep the main goroutine running.
	select {}
}
