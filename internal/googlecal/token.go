package googlecal

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"

	"github.com/AbdeltwabMF/gomeet/configs"
)

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

// saveToken saves the OAuth2 token to a file. The token is written to a
// temporary file and renamed into place, so an interrupted write cannot
// destroy a valid token.
func saveToken(tok *oauth2.Token) error {
	dir, err := configs.TokenDir()
	if err != nil {
		return err
	}

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

	if err := os.Rename(tmpName, filepath.Join(dir, configs.TokenFile)); err != nil {
		return fmt.Errorf("replace token: %w", err)
	}

	return nil
}
