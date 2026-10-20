// Package googlecal monitors Google Calendar events and reminds the user
// when a meeting is about to start.
package googlecal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/platform"
	"github.com/AbdeltwabMF/gomeet/internal/schedule"
)

const (
	// fetchInterval is how often the calendar is refreshed.
	fetchInterval = time.Minute + 7*time.Second
	// checkInterval is how often occurrences are evaluated.
	checkInterval = time.Minute
	// reloadInterval is how often the configuration file is reloaded.
	reloadInterval = time.Minute + 7*time.Second
	// startupGrace covers the small delay between a meeting starting and the
	// monitor noticing it.
	startupGrace = 5 * time.Minute
	// requestTimeout bounds a single Calendar API request.
	requestTimeout = 30 * time.Second
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

	// Persist refreshed tokens so that a restart does not require the user to
	// authorize again. The stored token is remembered, so that it is only
	// written again when it actually changed.
	source := &tokenSource{source: cfg.TokenSource(ctx, tok), last: tok}
	client := oauth2.NewClient(ctx, source)

	return calendar.NewService(ctx, option.WithHTTPClient(client))
}

// Monitor watches Google Calendar for meeting occurrences until the context
// is cancelled.
func Monitor(ctx context.Context, cfg *configs.Config) error {
	srv, err := initService(ctx)
	if err != nil {
		return err
	}

	deduper := schedule.NewDeduper(256)
	autoStart := cfg.AutoStart

	// Events are merged instead of replaced, so that an event which appears
	// shortly before its start time is still reported.
	known := make(map[string]*calendar.Event)

	// Fetch before the first check so that occurrences shortly after startup
	// are not missed.
	fetched, err := Fetch(ctx, srv, time.Now())
	if err != nil {
		return err
	}
	mergeEvents(known, fetched, time.Now())

	check := func(now time.Time) error {
		for key, e := range known {
			start, ok := eventStart(e)
			if !ok {
				continue
			}
			if !due(start, now) {
				continue
			}

			if !deduper.First(key, start, now) {
				continue
			}
			if err := Execute(ctx, e, autoStart); err != nil {
				// A failed reminder must not stop the monitor.
				slog.Error("execute event failed", slog.Any("error", err), slog.String("event", e.Id))
			}
		}

		return nil
	}

	if err := check(time.Now()); err != nil {
		return err
	}

	fetchTicker := time.NewTicker(fetchInterval)
	defer fetchTicker.Stop()

	checkTicker := time.NewTicker(checkInterval)
	defer checkTicker.Stop()

	reloadTicker := time.NewTicker(reloadInterval)
	defer reloadTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-checkTicker.C:
			if err := check(now); err != nil {
				return err
			}
		case <-fetchTicker.C:
			fetched, err := Fetch(ctx, srv, time.Now())
			if err != nil {
				// Keep the previous events when a refresh fails.
				slog.Error("fetch calendar events failed", slog.Any("error", err))
				continue
			}
			mergeEvents(known, fetched, time.Now())
			slog.Info("Received events", slog.Int("count", len(known)))
		case <-reloadTicker.C:
			reloaded, err := configs.LoadConfig()
			if err != nil {
				// Keep the previous settings when a reload fails.
				slog.Error("reload configuration failed", slog.Any("error", err))
				continue
			}
			autoStart = reloaded.AutoStart
		}
	}
}

// Fetch retrieves the upcoming events for the local day, following pagination
// so that busy days are fully covered.
func Fetch(ctx context.Context, srv *calendar.Service, now time.Time) ([]*calendar.Event, error) {
	loc := now.Location()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)

	// Start the window slightly in the past so that a meeting which has
	// already begun is still returned.
	timeMin := now.Add(-startupGrace)

	var events []*calendar.Event
	pageToken := ""
	for {
		call := srv.Events.List("primary").
			TimeMin(timeMin.Format(time.RFC3339)).
			TimeMax(endOfDay.Format(time.RFC3339)).
			SingleEvents(true).
			OrderBy("startTime")
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		reqCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		page, err := call.Context(reqCtx).Do()
		cancel()
		if err != nil {
			return nil, fmt.Errorf("list events: %w", err)
		}

		for _, e := range page.Items {
			if e == nil || e.Status == "cancelled" {
				continue
			}
			events = append(events, e)
		}

		pageToken = page.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return events, nil
}

// due reports whether an occurrence that starts at start should be reported
// at the time now. An occurrence is reported once it has started, and stays
// reportable for a short grace period so that events which appear shortly
// before their start time are not missed.
func due(start, now time.Time) bool {
	if start.After(now) {
		return false
	}

	return now.Sub(start) <= startupGrace
}

// mergeEvents adds the fetched events to the known events and drops the
// events that are too old to be reported.
func mergeEvents(known map[string]*calendar.Event, fetched []*calendar.Event, now time.Time) {
	for _, e := range fetched {
		start, ok := eventStart(e)
		if !ok {
			continue
		}
		known[eventKey(e, start)] = e
	}

	for key, e := range known {
		start, ok := eventStart(e)
		if !ok || now.Sub(start) > startupGrace {
			delete(known, key)
		}
	}
}

// eventStart returns the absolute start time of an event.
func eventStart(e *calendar.Event) (time.Time, bool) {
	if e == nil || e.Start == nil {
		return time.Time{}, false
	}

	// All-day events report a date without a time and are not meetings.
	if e.Start.DateTime == "" {
		return time.Time{}, false
	}

	start, err := time.Parse(time.RFC3339, e.Start.DateTime)
	if err != nil {
		slog.Error("parse event start failed", slog.Any("error", err), slog.String("event", e.Id))
		return time.Time{}, false
	}

	return start, true
}

// eventKey builds a stable key for a single occurrence.
func eventKey(e *calendar.Event, start time.Time) string {
	id := e.Id
	if id == "" {
		id = e.ICalUID
	}

	return fmt.Sprintf("%s\x00%s\x00%s", id, e.Location, start.UTC().Format(time.RFC3339))
}

// Execute performs the reminder actions for an event. The meeting link is
// opened automatically only when autoStart is enabled.
func Execute(ctx context.Context, e *calendar.Event, autoStart bool) error {
	if e.Location == "" {
		// The meeting link is stored in the location field, so an event
		// without one has nothing to open.
		slog.Debug("event has no meeting link", slog.String("event", e.Id))
		return nil
	}

	var errs []error

	if err := platform.Notify(ctx, e.Summary, e.Location); err != nil {
		errs = append(errs, err)
	}
	if autoStart {
		if err := platform.OpenURL(ctx, e.Location); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
