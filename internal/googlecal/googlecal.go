package googlecal

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

// initService initializes the Google Calendar API service.
func initService(ctx context.Context) (*calendar.Service, error) {
	f, err := configs.OpenCredentials()
	if err != nil {
		return nil, fmt.Errorf("open credentials: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	// If modifying these scopes, delete your previously saved token.json
	cfg, err := google.ConfigFromJSON(b, calendar.CalendarEventsReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	tok, err := getToken(ctx, cfg)
	if err != nil {
		return nil, err
	}

	client := cfg.Client(ctx, tok)

	return calendar.NewService(ctx, option.WithHTTPClient(client))
}

// Monitor continuously monitors Google Calendar events until the context is
// cancelled.
func Monitor(ctx context.Context, cfg *configs.Config) error {
	srv, err := initService(ctx)
	if err != nil {
		return err
	}

	events := new(calendar.Events)

	fetchTicker := time.NewTicker(time.Minute + 7*time.Second)
	defer fetchTicker.Stop()

	checkTicker := time.NewTicker(time.Minute)
	defer checkTicker.Stop()

	c := make(chan *calendar.Events, 1)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case events = <-c:
			slog.Info("Received events", slog.Int("count", len(events.Items)))
		case <-fetchTicker.C:
			go Fetch(c, srv)
		case <-checkTicker.C:
			Check(ctx, events, cfg) // Check in this goroutine to prevent unsynchronized access to events
		}
	}
}

// Fetch retrieves upcoming Google Calendar events for the day.
func Fetch(c chan<- *calendar.Events, srv *calendar.Service) {
	now := time.Now()
	endofday := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)

	events, err := srv.Events.List("primary").
		TimeMin(now.Format(time.RFC3339)).
		TimeMax(endofday.Format(time.RFC3339)).
		MaxResults(7).
		SingleEvents(true).
		OrderBy("startTime").
		Do()
	if err != nil {
		slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
		return
	}

	c <- events
}

// Check checks all events for a match with the current time and executes actions accordingly.
func Check(ctx context.Context, events *calendar.Events, cfg *configs.Config) {
	for _, e := range events.Items {
		if Match(e) {
			if err := Execute(ctx, e, cfg.AutoStart); err != nil {
				slog.Error("execute event failed", slog.Any("error", err))
			}
		}
	}
}

// Match checks if the calendar event start time matches the current time.
func Match(e *calendar.Event) bool {
	st, err := time.Parse(time.RFC3339, e.Start.DateTime)
	if err != nil {
		slog.Error("parse event start failed", slog.Any("error", err))
		return false
	}

	now := time.Now()
	slog.Debug("Matching event",
		slog.String("now", now.Format("15:04")),
		slog.String("then", st.Format("15:04")),
	)

	return st.Format("15:04") == now.Format("15:04")
}

// Execute executes actions based on the given calendar event.
func Execute(ctx context.Context, e *calendar.Event, autoStart bool) error {
	if err := platform.Notify(ctx, e.Summary, e.Location); err != nil {
		slog.Error("send notification failed", slog.Any("error", err))
	}

	if autoStart {
		return platform.OpenURL(ctx, e.Location)
	}

	return nil
}
