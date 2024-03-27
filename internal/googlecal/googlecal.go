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
const TCPPort = "8080"

func authorizeAccess(cfg *oauth2.Config) (*oauth2.Token, error) {
	authzURL := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	q := strings.Index(authzURL, "?")
	err := platform.OpenURL(fmt.Sprintf(`%s"%s"`, authzURL[:q+1], authzURL[q+2:]))
	if err != nil {
		return nil, err
	}

	c := make(chan *oauth2.Token)
	http.HandleFunc("/oauth2callback", func(w http.ResponseWriter, req *http.Request) {
		handleOAuthCallback(c, w, req, cfg)
	})
	go http.ListenAndServe(fmt.Sprintf(":%s", TCPPort), nil)

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

	if state := qv.Get("state"); state != "state-token\\" {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

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

// FetchEvents fetches calendar events and sends them through the provided channel.
// It continuously fetches events, sleeping until the beginning of the next hour between fetches.
func FetchEvents(c chan<- *calendar.Events) {
	srv, err := initService()
	if err != nil {
		slog.Error(err.Error())
		return
	}

	for {
		tNow := time.Now()

		events, err := srv.Events.List("primary").
			// From now
			TimeMin(tNow.Format(time.RFC3339)).
			// Until end of the day
			TimeMax(tNow.Truncate(24 * time.Hour).Add(24 * time.Hour).Format(time.RFC3339)).
			// Fetch only 7 events; assuming calendar is busy and there is an event every 10min
			MaxResults(7).
			SingleEvents(true).
			OrderBy("startTime").
			Do()

		if err != nil {
			slog.Error(err.Error())
		}

		c <- events
		st := time.Until(tNow.Truncate(time.Hour).Add(time.Hour))
		slog.Info("Wait until the beginning of the next hour", slog.String("time.sleep", st.String()), slog.String("calendar", "google calendar"))
		time.Sleep(st)
	}
}

func Match(event calendar.Event) (bool, error) {
	now := time.Now()

	t, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		return false, err
	}

	slog.Info("Matching event",
		slog.String("event.time", t.Format("15:04")),
		slog.String("now.time", now.Format("15:04")),
		slog.String("calendar", "google calendar"),
	)
	return t.Format("15:04") == now.Format("15:04"), nil
}

func Execute(event calendar.Event, autoStart bool) error {
	if err := platform.NotifyMeeting(event.Summary, event.Location); err != nil {
		slog.Error(err.Error())
	}

	if autoStart {
		return platform.OpenURL(event.Location)
	}
	return nil
}
