package localcal

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

const CalendarName = "Local calendar"

type Start struct {
	Time string   `json:"time"`
	Days []string `json:"days"`
}

type Event struct {
	Summary string `json:"summary"`
	Url     string `json:"url"`
	Start   Start  `json:"start"`
}

type Events struct {
	Items []*Event `json:"events"`
}

func waitNextHour() {
	sd := time.Until(time.Now().Truncate(time.Hour).Add(time.Hour))

	slog.Info("Waiting for the next hour to reload events",
		slog.String("time.sleep", sd.String()),
		slog.String("calendar", CalendarName),
	)

	time.Sleep(sd)
}

// Load loads events from the local calendar and sends them to the provided channel.
func Load(ch chan<- *Events, errch chan<- error, retryLimit int) {
	d, err := platform.ConfigDir()
	if err != nil {
		errch <- err
		return
	}

	file, err := os.OpenFile(filepath.Join(d, "config.json"), os.O_CREATE|os.O_RDONLY, 0640)
	if err != nil {
		errch <- err
		return
	}
	defer file.Close()

	for {
		var events Events
		var err error

		for i := 0; i < retryLimit; i++ {
			if err = json.NewDecoder(file).Decode(&events); err == nil {
				break
			}
		}

		if err != nil {
			errch <- err
			return
		}

		ch <- &events
		waitNextHour()
	}
}

// Match checks if the given event matches the current time(hh:mm) and day.
func Match(event *Event) bool {
	now := time.Now()
	for _, d := range event.Start.Days {
		if d == now.Weekday().String() {
			slog.Info("Matching event",
				slog.String("event.time", event.Start.Time),
				slog.String("now.time", now.Format("15:04")),
				slog.String("calendar", CalendarName),
			)

			return event.Start.Time == now.Format("15:04")
		}
	}

	return false
}

// Execute executes actions associated with the given event, such as notifying and potentially starting a meeting.
func Execute(event *Event, autoStart bool) error {
	if err := platform.Notify(event.Summary, event.Url); err != nil {
		slog.Error(err.Error())
	}

	if autoStart {
		return platform.OpenURL(event.Url)
	}
	return nil
}
