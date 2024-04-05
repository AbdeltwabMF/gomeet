package googlecal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/AbdeltwabMF/gomeet/configs"
	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

// handleOAuthCallback exchanges the authorization code for a token and sends it to the provided channel.
func handleOAuthCallback(c chan<- *oauth2.Token, cfg *oauth2.Config, w http.ResponseWriter, r *http.Request) {
	qv := r.URL.Query()
	code := qv.Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	tok, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Unable to retrieve token: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Access granted! You can now close this window.")

	c <- tok
}

// authorizeAccess opens a browser window for the user to authenticate and authorize access.
func authorizeAccess(cfg *oauth2.Config) error {
	c := make(chan *oauth2.Token, 1)

	server := &http.Server{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleOAuthCallback(c, cfg, w, r)
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
		}
	}()

	defer func() {
		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
		}
	}()

	url := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	q := strings.Index(url, "?")
	if err := platform.OpenURL(fmt.Sprintf(`%s"%s"`, url[:q+1], url[q+2:])); err != nil {
		return err
	}

	tok := <-c
	return saveToken(tok)
}

// saveToken saves the OAuth2 token to a file.
func saveToken(tok *oauth2.Token) error {
	f, err := configs.OpenToken(os.O_CREATE | os.O_TRUNC | os.O_WRONLY)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(tok)
}

// loadToken loads the OAuth2 token from a file.
func loadToken() (*oauth2.Token, error) {
	f, err := configs.OpenToken(os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)

	return tok, err
}

// getToken retrieves the OAuth2 token. If it does not exist, it creates a new token.
func getToken(cfg *oauth2.Config) (*oauth2.Token, error) {
	tok, err := loadToken()
	if err != nil {
		slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))

		// Create a new access token
		if err := authorizeAccess(cfg); err != nil {
			return nil, err
		}

		tok, err = loadToken()
	}

	return tok, err
}

// initService initializes the Google Calendar API service.
func initService() (*calendar.Service, error) {
	f, err := configs.OpenCredentials(os.O_RDONLY)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	// If modifying these scopes, delete your previously saved token.json
	cfg, err := google.ConfigFromJSON(b, calendar.CalendarEventsReadonlyScope)
	if err != nil {
		return nil, err
	}

	tok, err := getToken(cfg)
	if err != nil {
		return nil, err
	}

	client := cfg.Client(context.Background(), tok)
	ctx := context.Background()

	return calendar.NewService(ctx, option.WithHTTPClient(client))
}

// Monitor continuously monitors Google Calendar events.
func Monitor(cfg *configs.Config) {
	srv, err := initService()
	if err != nil {
		slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
		return
	}

	events := new(calendar.Events)

	fetchTicker := time.NewTicker(time.Minute + time.Second*7)
	checkTicker := time.NewTicker(time.Minute)

	c := make(chan *calendar.Events, 1)

	for {
		select {
		case events = <-c:
			slog.Info("Received events", slog.Int("count", len(events.Items)), slog.Any("func", configs.CallerInfo()))
		case <-fetchTicker.C:
			go Fetch(c, srv)
		case <-checkTicker.C:
			Check(events, cfg) // Check in this goroutine to prevent unsynchronized access to events
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
func Check(events *calendar.Events, cfg *configs.Config) {
	for _, e := range events.Items {
		if Match(e) {
			if err := Execute(e, cfg.AutoStart); err != nil {
				slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
			}
		}
	}
}

// Match checks if the calendar event start time matches the current time.
func Match(e *calendar.Event) bool {
	st, err := time.Parse(time.RFC3339, e.Start.DateTime)
	if err != nil {
		slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
		return false
	}

	now := time.Now()
	slog.Debug("Matching event",
		slog.String("now", now.Format("15:04")),
		slog.String("then", st.Format("15:04")),
		slog.Any("func", configs.CallerInfo()),
	)

	return st.Format("15:04") == now.Format("15:04")
}

// Execute executes actions based on the given calendar event.
func Execute(e *calendar.Event, autoStart bool) error {
	if err := platform.Notify(e.Summary, e.Location); err != nil {
		slog.Error(err.Error(), slog.Any("func", configs.CallerInfo()))
	}

	if autoStart {
		return platform.OpenURL(e.Location)
	}

	return nil
}
