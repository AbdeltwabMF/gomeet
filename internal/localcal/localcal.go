package localcal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

var CalAttr = slog.String("calendar", "Local")

func waitNextMinute() {
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
}

func Monitor(cfg *configs.Config) {
	var events *configs.Events
	c := make(chan *configs.Events, 2)
	errc := make(chan error, 2)

	go Load(c, errc)

	for {
		select {
		case events = <-c:
			slog.Info("Received events", slog.Int("events.count", len(events.Items)), CalAttr)
		case err := <-errc:
			slog.Error(fmt.Sprintf("Received error: %v", err), CalAttr)
			go func() {
				waitNextMinute()
				Load(c, errc)
			}()
		default:
			if events != nil {
				for _, item := range events.Items {
					if Match(item) {
						err := Execute(item, cfg.AutoStart)
						if err != nil {
							slog.Error(fmt.Sprintf("Execute: %v", err.Error()), CalAttr)
						}
					}
				}
			}
			waitNextMinute()
		}
	}
}

// Load loads events from the local calendar and sends them to the provided channel.
func Load(ch chan<- *configs.Events, errch chan<- error) {
	d, err := platform.ConfigDir()
	if err != nil {
		errch <- err
		return
	}

	file, err := os.OpenFile(filepath.Join(d, configs.ConfigFile), os.O_CREATE|os.O_RDONLY, 0640)
	if err != nil {
		errch <- err
		return
	}
	defer file.Close()

	for {
		var events configs.Events
		_, err := file.Seek(0, 0)
		if err != nil {
			errch <- err
			return
		}

		err = json.NewDecoder(file).Decode(&events)
		if err != nil {
			errch <- err
			return
		}

		ch <- &events
		waitNextMinute()
	}
}

// Match checks if the given event matches the current time(hh:mm) and day.
func Match(event *configs.Event) bool {
	now := time.Now()
	for _, d := range event.Start.Days {
		if d == now.Weekday().String() {
			slog.Info("Match",
				slog.String("event.time", event.Start.Time),
				slog.String("now.time", now.Format("15:04")),
				CalAttr,
			)

			return event.Start.Time == now.Format("15:04")
		}
	}

	return false
}

// Execute executes actions associated with the given event, such as notifying and potentially starting a meeting.
func Execute(event *configs.Event, autoStart bool) error {
	if err := platform.Notify(event.Summary, event.Url); err != nil {
		slog.Error(err.Error())
	}

	if autoStart {
		return platform.OpenURL(event.Url)
	}
	return nil
}
