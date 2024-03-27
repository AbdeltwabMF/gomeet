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

func waitMin() {
	st := time.Until(time.Now().Truncate(time.Minute).Add(time.Minute))
	slog.Info("Done iteration; time to sleep", slog.String("sleep.time", st.String()))
	time.Sleep(st)
}

func main() {
	file, err := initLogger()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	cfg, err := loadConfig()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	var gEvents *calendar.Events
	var lEvents []localcal.Event
	gc := make(chan *calendar.Events)
	lc := make(chan []localcal.Event)

	go googlecal.FetchEvents(gc)
	go localcal.LoadEvents(lc)

	gEvents = <-gc
	lEvents = <-lc

	go func() {
		for {
			select {
			case gEvents = <-gc:
				slog.Info("Google calendar channel is ready")
			default:
				for i, item := range gEvents.Items {
					fmt.Printf("%d. %v %v (%s)\n", i, item.Summary, item.Start.DateTime, item.Location)
					if googlecal.Match(*item) {
						googlecal.Execute(*item, cfg.AutoStart)
					}
				}
				waitMin()
			}
		}
	}()

	go func() {
		for {
			select {
			case lEvents = <-lc:
				slog.Info("Local calendar channel is ready")
			default:
				for i, item := range lEvents {
					fmt.Printf("%d. %v %v (%s)\n", i, item.Summary, item.Start.Time, item.Url)
					if localcal.Match(item) {
						localcal.Execute(item, cfg.AutoStart)
					}
				}
				waitMin()
			}
		}
	}()

	select {}
}
