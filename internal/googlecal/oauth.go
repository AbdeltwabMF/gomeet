package googlecal

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/AbdeltwabMF/gomeet/internal/platform"
)

const (
	// authorizeTimeout bounds the whole browser authorization flow.
	authorizeTimeout = 5 * time.Minute
	// exchangeTimeout bounds the authorization code exchange.
	exchangeTimeout = 30 * time.Second
	// shutdownTimeout bounds the graceful shutdown of the callback server.
	shutdownTimeout = 5 * time.Second
)

// authorizeAccess opens a browser window for the user to authorize access and
// returns the resulting token.
func authorizeAccess(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(ctx, authorizeTimeout)
	defer cancel()

	// Bind an ephemeral port on the loopback interface only.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("listen for oauth callback: %w", err)
	}

	// Use the allocated port for the redirect URI.
	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		listener.Close()
		return nil, fmt.Errorf("unexpected listener address %v", listener.Addr())
	}

	oauthCfg := *cfg
	oauthCfg.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/oauth2/callback", addr.Port)

	state, err := randomState()
	if err != nil {
		listener.Close()
		return nil, err
	}

	result := make(chan *oauth2.Token, 1)

	// Use a private mux so that no other handler is affected.
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		handleOAuthCallback(result, &oauthCfg, state, w, r)
	})

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown oauth server failed", slog.Any("error", err))
		}
	}()

	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	if err := platform.OpenURL(ctx, authURL); err != nil {
		return nil, fmt.Errorf("open browser: %w", err)
	}

	select {
	case tok := <-result:
		return tok, nil
	case err := <-serveErr:
		return nil, fmt.Errorf("oauth callback server failed: %w", err)
	case <-ctx.Done():
		return nil, fmt.Errorf("authorization timed out: %w", ctx.Err())
	}
}

// handleOAuthCallback exchanges the authorization code for a token and sends
// it to the provided channel.
func handleOAuthCallback(c chan<- *oauth2.Token, cfg *oauth2.Config, state string, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	if errCode := query.Get("error"); errCode != "" {
		http.Error(w, "Authorization denied: "+errCode, http.StatusBadRequest)
		return
	}

	code := query.Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	if subtle.ConstantTimeCompare([]byte(query.Get("state")), []byte(state)) != 1 {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), exchangeTimeout)
	defer cancel()

	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		http.Error(w, "Unable to retrieve token", http.StatusInternalServerError)
		slog.Error("exchange authorization code failed", slog.Any("error", err))
		return
	}

	fmt.Fprint(w, "Access granted! You can now close this window.")

	// Never block the handler if a token was already delivered.
	select {
	case c <- tok:
	default:
	}
}

// randomState returns a cryptographically random OAuth state value.
func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
