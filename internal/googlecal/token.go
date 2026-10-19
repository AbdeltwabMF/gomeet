package googlecal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/oauth2"

	"github.com/AbdeltwabMF/gomeet/configs"
)

// getToken retrieves the OAuth2 token. It starts an interactive authorization
// only when no token has been stored yet.
func getToken(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	tok, err := loadToken()
	if err == nil {
		return tok, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		// A corrupt or unreadable token cannot be fixed by authorizing
		// again, so report it with actionable guidance.
		return nil, fmt.Errorf("load token: %w (delete %s to authorize again)", err, configs.TokenFile)
	}

	// No token exists yet, so ask the user to authorize this application.
	// The callback stores the token before reporting success.
	tok, err = authorizeAccess(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return tok, nil
}

// loadToken loads the OAuth2 token from a file.
func loadToken() (*oauth2.Token, error) {
	f, err := configs.OpenToken()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(tok); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}

	return tok, nil
}

// readToken loads the OAuth2 token from the given path.
func readToken(path string) (*oauth2.Token, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	tok := &oauth2.Token{}
	if err := json.Unmarshal(b, tok); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}

	return tok, nil
}

// saveToken saves the OAuth2 token to a file. The token is written to a
// temporary file and renamed into place, so an interrupted write cannot
// destroy a valid token.
func saveToken(tok *oauth2.Token) error {
	dir, err := configs.TokenDir()
	if err != nil {
		return err
	}

	return writeToken(filepath.Join(dir, configs.TokenFile), tok)
}

// writeToken stores the token at path atomically.
func writeToken(path string, tok *oauth2.Token) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, "token-*.json")
	if err != nil {
		return fmt.Errorf("create temporary token file: %w", err)
	}
	tmpName := tmp.Name()

	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()

	if err := tmp.Chmod(0600); err != nil {
		return fmt.Errorf("set token permissions: %w", err)
	}
	if err := json.NewEncoder(tmp).Encode(tok); err != nil {
		return fmt.Errorf("encode token: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync token: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close token: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace token: %w", err)
	}

	return nil
}

// tokenSource persists refreshed tokens atomically, so that a restart does
// not require the user to authorize again.
type tokenSource struct {
	source oauth2.TokenSource
	mu     sync.Mutex
}

// Token returns a valid token and stores it when it has been refreshed.
func (s *tokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tok, err := s.source.Token()
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	if err := saveToken(tok); err != nil {
		return nil, fmt.Errorf("save refreshed token: %w", err)
	}

	return tok, nil
}
