package googlecal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestRandomStateIsUnique(t *testing.T) {
	t.Parallel()

	first, err := randomState()
	if err != nil {
		t.Fatalf("randomState() error = %v", err)
	}
	second, err := randomState()
	if err != nil {
		t.Fatalf("randomState() error = %v", err)
	}
	if first == second {
		t.Fatal("randomState() must not repeat")
	}
	if len(first) < 32 {
		t.Fatalf("randomState() length = %d, want at least 32", len(first))
	}
}

func TestHandleOAuthCallbackRejectsBadState(t *testing.T) {
	t.Parallel()

	result := make(chan oauthResult, 1)
	cfg := &oauth2.Config{ClientID: "id", ClientSecret: "secret", Endpoint: oauth2.Endpoint{TokenURL: "https://example.com/token"}}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/oauth2/callback?code=abc&state=wrong", nil)

	handleOAuthCallback(context.Background(), result, cfg, "expected-state", recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	select {
	case tok := <-result:
		t.Fatalf("no token should be delivered, got %v", tok)
	default:
	}
}

func TestHandleOAuthCallbackRejectsMissingCode(t *testing.T) {
	t.Parallel()

	result := make(chan oauthResult, 1)
	cfg := &oauth2.Config{ClientID: "id", ClientSecret: "secret"}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/oauth2/callback?state=expected-state", nil)

	handleOAuthCallback(context.Background(), result, cfg, "expected-state", recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestHandleOAuthCallbackRejectsProviderError(t *testing.T) {
	t.Parallel()

	result := make(chan oauthResult, 1)
	cfg := &oauth2.Config{ClientID: "id", ClientSecret: "secret"}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/oauth2/callback?error=access_denied&state=expected-state", nil)

	handleOAuthCallback(context.Background(), result, cfg, "expected-state", recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	select {
	case res := <-result:
		if res.err == nil {
			t.Fatal("a denial must be reported to the caller")
		}
	default:
		t.Fatal("a denial must be reported to the caller")
	}
}

func TestHandleOAuthCallbackRejectsWrongMethod(t *testing.T) {
	t.Parallel()

	result := make(chan oauthResult, 1)
	cfg := &oauth2.Config{ClientID: "id", ClientSecret: "secret"}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/oauth2/callback?code=abc&state=expected-state", nil)

	handleOAuthCallback(context.Background(), result, cfg, "expected-state", recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleOAuthCallbackRejectsExchangeFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	result := make(chan oauthResult, 1)
	cfg := &oauth2.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		Endpoint:     oauth2.Endpoint{TokenURL: server.URL},
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/oauth2/callback?code=abc&state=expected-state", nil)

	handleOAuthCallback(context.Background(), result, cfg, "expected-state", recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	select {
	case res := <-result:
		if res.err == nil {
			t.Fatal("expected an error result")
		}
		if res.token != nil {
			t.Fatal("no token should be delivered")
		}
	default:
		t.Fatal("an error result should be delivered")
	}
}

func TestAuthorizeAccessDoesNotBlockForever(t *testing.T) {
	t.Parallel()

	cfg := &oauth2.Config{ClientID: "id", ClientSecret: "secret", Endpoint: oauth2.Endpoint{AuthURL: "https://example.com/auth"}}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := authorizeAccess(ctx, cfg)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("authorizeAccess() error = nil, want an error")
		}
		if !strings.Contains(err.Error(), "open browser") && !strings.Contains(err.Error(), "timed out") {
			t.Fatalf("authorizeAccess() error = %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("authorizeAccess() did not return")
	}
}
