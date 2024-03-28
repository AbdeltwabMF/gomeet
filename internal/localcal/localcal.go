package localcal

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

const LocalCalendar = "Local calendar"

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

func LoadEvents(c chan *Events) {
	for {
		d, err := platform.ConfigDir()
		if err != nil {
			slog.Error(err.Error())
			return
		}

		file, err := os.OpenFile(filepath.Join(d, "config.json"), os.O_CREATE|os.O_RDONLY, 0640)
		if err != nil {
			slog.Error(err.Error())
			return
		}
		defer file.Close()

		var events Events
		if err := json.NewDecoder(file).Decode(&events); err != nil {
			slog.Error(err.Error())
			return
		}

		c <- &events
		st := time.Until(time.Now().Truncate(time.Hour).Add(time.Hour))
		slog.Info("Wait until the beginning of the next hour",
			slog.String("time.sleep", st.String()),
			slog.String("calendar", LocalCalendar),
		)
		time.Sleep(st)
	}
}

func Match(event *Event) bool {
	now := time.Now()
	slog.Info("Matching event",
		slog.String("event.time", event.Start.Time),
		slog.String("now.time", now.Format("15:04")),
		slog.String("calendar", LocalCalendar),
	)

	for _, d := range event.Start.Days {
		if d == now.Weekday().String() {
			return event.Start.Time == now.Format("15:04")
		}
	}

	return false
}

func Execute(event *Event, autoStart bool) error {
	if err := platform.NotifyMeeting(event.Summary, event.Url); err != nil {
		slog.Error(err.Error())
	}

	if autoStart {
		return platform.OpenURL(event.Url)
	}
	return nil
}
