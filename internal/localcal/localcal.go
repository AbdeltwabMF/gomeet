// Package localcal monitors the recurring events defined in the local
// configuration file.
package localcal

import (
	"context"
	"log/slog"
	"time"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

// Monitor continuously monitors local calendar events until the context is
// cancelled.
func Monitor(ctx context.Context, cfg *configs.Config) error {
	events := cfg.Events

	loadTicker := time.NewTicker(time.Minute + 7*time.Second)
	defer loadTicker.Stop()

	checkTicker := time.NewTicker(time.Minute)
	defer checkTicker.Stop()

	c := make(chan []configs.Event, 1)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case events = <-c:
			slog.Info("Received events", slog.Int("count", len(events)))
		case <-loadTicker.C:
			go Load(c)
		case <-checkTicker.C:
			Check(ctx, events, cfg) // Check in this goroutine to prevent unsynchronized access to events
		}
	}
}

// Load loads local calendar events from the configuration file.
func Load(c chan<- []configs.Event) {
	cfg, err := configs.LoadConfig()
	if err != nil {
		slog.Error("load configuration failed", slog.Any("error", err))
		return
	}

	c <- cfg.Events
}

// Check checks all events for a match with the current day and time and executes actions accordingly.
func Check(ctx context.Context, events []configs.Event, cfg *configs.Config) {
	for i := range events {
		if Match(&events[i]) {
			if err := Execute(ctx, &events[i], cfg.AutoStart); err != nil {
				slog.Error("execute event failed", slog.Any("error", err))
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
			)

			return e.Start.Time == now.Format("15:04")
		}
	}

	return false
}

// Execute executes actions based on the given calendar event.
func Execute(ctx context.Context, e *configs.Event, autoStart bool) error {
	if err := platform.Notify(ctx, e.Summary, e.URL); err != nil {
		slog.Error("send notification failed", slog.Any("error", err))
	}

	if autoStart {
		return platform.OpenURL(ctx, e.URL)
	}

	return nil
}
