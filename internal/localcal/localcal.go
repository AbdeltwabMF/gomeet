package localcal

import (
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

// Monitor continuously monitors local calendar events.
func Monitor(cfg *configs.Config) {
	events := new(configs.Events)

	loadTicker := time.NewTicker(time.Minute + time.Second*7)
	checkTicker := time.NewTicker(time.Minute)

	c := make(chan *configs.Events, 1)

	for {
		select {
		case events = <-c:
			slog.Info("Received events", slog.Int("count", len(events.Items)), slog.Any("func", configs.CallerInfo()))
		case <-loadTicker.C:
			go Load(c)
		case <-checkTicker.C:
			Check(events, cfg) // Check in this goroutine to prevent unsynchronized access to events
		}
	}
}

// Load loads local calendar events from the configuration file.
func Load(c chan<- *configs.Events) {
	f, err := configs.OpenConfig(os.O_RDONLY)
	if err != nil {
		slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
		return
	}
	defer f.Close()

	events := new(configs.Events)
	if err = json.NewDecoder(f).Decode(events); err != nil {
		slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
		return
	}

	c <- events
}

// Check checks all events for a match with the current day and time and executes actions accordingly.
func Check(events *configs.Events, cfg *configs.Config) {
	for _, e := range events.Items {
		if Match(e) {
			if err := Execute(e, cfg.AutoStart); err != nil {
				slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
			}
		}
	}
}

// Match checks if the calendar event matches the current day and time.
func Match(e *configs.Event) (ok bool) {
	now := time.Now()

	for _, d := range e.Start.Days {
		if d == now.Weekday().String() {
			slog.Debug("Matching event",
				slog.String("now", now.Format("15:04")),
				slog.String("then", e.Start.Time),
				slog.Any("func", configs.CallerInfo()),
			)

			return e.Start.Time == now.Format("15:04")
		}
	}

	return false
}

// Execute executes actions based on the given calendar event.
func Execute(e *configs.Event, autoStart bool) error {
	if err := platform.Notify(e.Summary, e.Url); err != nil {
		slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
	}

	if autoStart {
		return platform.OpenURL(e.Url)
	}

	return nil
}
