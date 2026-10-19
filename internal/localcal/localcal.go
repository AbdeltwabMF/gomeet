// Package localcal monitors the recurring events defined in the local
// configuration file.
package localcal

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/platform"
	"github.com/AbdeltwabMF/gomeet/internal/schedule"
)

const (
	// checkInterval is how often occurrences are evaluated.
	checkInterval = time.Minute
	// reloadInterval is how often the configuration file is reloaded.
	reloadInterval = time.Minute + 7*time.Second
	// startupGrace covers the small delay between a meeting starting and the
	// monitor noticing it.
	startupGrace = 5 * time.Minute
)

// Monitor watches the local configuration for meeting occurrences until the
// context is cancelled.
func Monitor(ctx context.Context, cfg *configs.Config) error {
	loc := time.Local
	deduper := schedule.NewDeduper(256)

	events := cfg.Events
	autoStart := cfg.AutoStart
	lastCheck := time.Now().Add(-startupGrace)

	check := func(now time.Time) error {
		for i := range events {
			for _, occurrence := range occurrences(&events[i], lastCheck, now, loc) {
				key := localKey(&events[i], occurrence)
				if !deduper.First(key, occurrence, now) {
					continue
				}
				if err := Execute(ctx, &events[i], autoStart); err != nil {
					// A failed reminder must not stop the monitor.
					slog.Error("execute event failed", slog.Any("error", err), slog.String("summary", events[i].Summary))
				}
			}
		}
		lastCheck = now

		return nil
	}

	// Evaluate the initial configuration immediately so that occurrences
	// shortly after startup are not missed.
	if err := check(time.Now()); err != nil {
		return err
	}

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
		case <-reloadTicker.C:
			reloaded, err := configs.LoadConfig()
			if err != nil {
				// Keep using the previous configuration when a reload fails.
				slog.Error("reload configuration failed", slog.Any("error", err))
				continue
			}
			events = reloaded.Events
			autoStart = reloaded.AutoStart
			slog.Info("Reloaded events", slog.Int("count", len(events)))
		}
	}
}

// occurrences returns the start times of e that fall inside (after, through].
func occurrences(e *configs.Event, after, through time.Time, loc *time.Location) []time.Time {
	if e.Start.Time == "" || len(e.Start.Days) == 0 {
		return nil
	}

	hour, minute, ok := parseClock(e.Start.Time)
	if !ok {
		return nil
	}

	days := make(map[time.Weekday]bool, len(e.Start.Days))
	for _, day := range e.Start.Days {
		for weekday := time.Sunday; weekday <= time.Saturday; weekday++ {
			if day == weekday.String() {
				days[weekday] = true
			}
		}
	}

	// Walk the dates that may contain an occurrence. An occurrence inside
	// (after, through] must fall on a day between the day of after and the
	// day of through.
	from := startOfDay(after, loc)
	to := startOfDay(through, loc)

	// Bound the amount of work so that a long running process cannot scan
	// an unbounded number of days.
	const maxDays = 366

	var result []time.Time
	for date := from; !date.After(to) && len(result) < maxDays; date = date.AddDate(0, 0, 1) {
		day := date.In(loc)
		if !days[day.Weekday()] {
			continue
		}

		occurrence := time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, loc)
		if schedule.Occurrence(occurrence, after, through) {
			result = append(result, occurrence)
		}
	}

	return result
}

// startOfDay returns the start of the day that contains t in loc.
func startOfDay(t time.Time, loc *time.Location) time.Time {
	day := t.In(loc)

	return time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
}

// parseClock parses the strict "15:04" configuration format.
func parseClock(value string) (hour, minute int, ok bool) {
	if len(value) != 5 || value[2] != ':' {
		return 0, 0, false
	}

	hour = int(value[0]-'0')*10 + int(value[1]-'0')
	minute = int(value[3]-'0')*10 + int(value[4]-'0')
	if value[0] < '0' || value[0] > '9' || value[1] < '0' || value[1] > '9' ||
		value[3] < '0' || value[3] > '9' || value[4] < '0' || value[4] > '9' {
		return 0, 0, false
	}
	if hour > 23 || minute > 59 {
		return 0, 0, false
	}

	return hour, minute, true
}

// localKey builds a stable key for a single occurrence.
func localKey(e *configs.Event, occurrence time.Time) string {
	return e.Summary + "\x00" + e.URL + "\x00" + e.Start.Time + "\x00" + occurrence.Format(time.RFC3339)
}

// Execute performs the reminder actions for an event. The meeting link is
// opened automatically only when autoStart is enabled.
func Execute(ctx context.Context, e *configs.Event, autoStart bool) error {
	var errs []error

	if err := platform.Notify(ctx, e.Summary, e.URL); err != nil {
		errs = append(errs, err)
	}
	if autoStart {
		if err := platform.OpenURL(ctx, e.URL); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
