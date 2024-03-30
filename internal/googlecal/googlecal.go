package googlecal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

const TokenFile = "token.json"
const CredentialsFile = "credentials.json"
const CalendarName = "Google calendar"

func authorizeAccess(cfg *oauth2.Config) (*oauth2.Token, error) {
	authzURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	q := strings.Index(authzURL, "?")
	err := platform.OpenURL(fmt.Sprintf(`%s"%s"`, authzURL[:q+1], authzURL[q+2:]))
	if err != nil {
		return nil, err
	}

	c := make(chan *oauth2.Token)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		handleOAuthCallback(c, w, req, cfg)
	})

	go func() {
		if err := http.ListenAndServe("", nil); err != nil {
			slog.Error(err.Error())
		}
	}()

	d, err := platform.ConfigDir()
	if err != nil {
		return nil, err
	}

	tok := <-c
	return tok, saveToken(filepath.Join(d, TokenFile), tok)
}

func loadToken(path string) (*oauth2.Token, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	tok := &oauth2.Token{}
	return tok, json.NewDecoder(file).Decode(tok)
}

func saveToken(path string, tok *oauth2.Token) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(tok)
}

func handleOAuthCallback(c chan<- *oauth2.Token, w http.ResponseWriter, req *http.Request, cfg *oauth2.Config) {
	qv := req.URL.Query()
	code := qv.Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	tok, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to retrieve token: %v", err), http.StatusInternalServerError)
		return
	} else {
		fmt.Fprintf(w, "Authentication successful! You can now close this window.")
	}

	c <- tok
}

func initService() (*calendar.Service, error) {
	ctx := context.Background()

	c, err := platform.ConfigDir()
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(filepath.Join(c, CredentialsFile))
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json
	cfg, err := google.ConfigFromJSON(bytes, calendar.CalendarEventsReadonlyScope)
	if err != nil {
		return nil, err
	}

	tok, err := loadToken(filepath.Join(c, TokenFile))
	if err != nil {
		slog.Error(err.Error())

		tok, err = authorizeAccess(cfg)
		if err != nil {
			return nil, err
		}
	}

	client := cfg.Client(context.Background(), tok)
	return calendar.NewService(ctx, option.WithHTTPClient(client))
}

func waitNextHour() {
	sd := time.Until(time.Now().Truncate(time.Hour).Add(time.Hour))

	slog.Info("Waiting for the next hour to refetch events",
		slog.String("time.sleep", sd.String()),
		slog.String("calendar", CalendarName),
	)

	time.Sleep(sd)
}

// Fetch fetches calendar events and sends them through the provided channel.
// It periodically fetches events, sleeping until the beginning of the next hour between fetches.
func Fetch(ch chan<- *calendar.Events, errch chan<- error, retryLimit int) {
	srv, err := initService()
	if err != nil {
		errch <- err
		return
	}

	const (
		maxEvents  = 7
		hoursInDay = 24
	)

	for {
		now := time.Now()
		var events *calendar.Events

		for i := 0; i < retryLimit; i++ {
			events, err = srv.Events.List("primary").
				TimeMin(now.Format(time.RFC3339)).
				TimeMax(now.Truncate(hoursInDay * time.Hour).Add(hoursInDay * time.Hour).Format(time.RFC3339)).
				MaxResults(maxEvents).
				SingleEvents(true).
				OrderBy("startTime").
				Do()

			if err == nil {
				break
			}
		}

		if err != nil {
			errch <- err
			return
		}

		ch <- events
		waitNextHour()
	}
}

// Match checks if the given event matches the current time(hh:mm).
func Match(event *calendar.Event) (bool, error) {
	now := time.Now()

	t, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		return false, err
	}

	slog.Info("Matching event",
		slog.String("event.time", t.Format("15:04")),
		slog.String("now.time", now.Format("15:04")),
		slog.String("calendar", CalendarName),
	)

	return t.Format("15:04") == now.Format("15:04"), nil
}

// Execute executes actions associated with the given event, such as notifying and potentially starting a meeting.
func Execute(event *calendar.Event, autoStart bool) error {
	if err := platform.Notify(event.Summary, event.Location); err != nil {
		slog.Error(err.Error())
	}

	if autoStart {
		return platform.OpenURL(event.Location)
	}
	return nil
}
