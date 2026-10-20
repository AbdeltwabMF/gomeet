package googlecal

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestSaveTokenIsAtomicAndPrivate(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "token.json")
	tok := &oauth2.Token{AccessToken: "access", RefreshToken: "refresh"}

	if err := writeToken(path, tok); err != nil {
		t.Fatalf("writeToken() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat token: %v", err)
	}
	// Windows represents file access with ACLs and only exposes the writable
	// bit through os.FileMode, so a private file commonly reports 0666 there.
	// Exact Unix permission bits are meaningful only on Unix-like systems.
	if runtime.GOOS != "windows" {
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Fatalf("token permissions = %o, want 600", perm)
		}
	}

	// No temporary files may be left behind.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("directory contains %d entries, want 1", len(entries))
	}
}

func TestTokenSourceDoesNotRewriteUnchangedToken(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	tok := &oauth2.Token{AccessToken: "access", Expiry: time.Now().Add(time.Hour)}
	if err := writeToken(path, tok); err != nil {
		t.Fatalf("writeToken() error = %v", err)
	}

	before, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat token: %v", err)
	}

	source := &tokenSource{
		source: oauth2.StaticTokenSource(tok),
		last:   tok,
	}

	for i := 0; i < 3; i++ {
		if _, err := source.Token(); err != nil {
			t.Fatalf("Token() error = %v", err)
		}
	}

	after, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat token: %v", err)
	}
	if !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("the token file must not be rewritten when the token is unchanged")
	}
}

func TestWriteTokenKeepsPreviousFileOnFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	original := &oauth2.Token{AccessToken: "original"}
	if err := writeToken(path, original); err != nil {
		t.Fatalf("writeToken() error = %v", err)
	}

	// A directory in place of the target file makes the rename fail.
	blocked := filepath.Join(dir, "blocked")
	if err := os.Mkdir(blocked, 0750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := writeToken(blocked, &oauth2.Token{AccessToken: "other"}); err == nil {
		t.Fatal("writeToken() error = nil, want an error")
	}

	// The original token must still be readable.
	loaded, err := readToken(path)
	if err != nil {
		t.Fatalf("readToken() error = %v", err)
	}
	if loaded.AccessToken != "original" {
		t.Fatalf("AccessToken = %q, want %q", loaded.AccessToken, "original")
	}
}

func TestValidateToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		token   *oauth2.Token
		wantErr bool
	}{
		{name: "current access token", token: &oauth2.Token{AccessToken: "access", Expiry: time.Now().Add(time.Hour)}},
		{name: "refreshable expired token", token: &oauth2.Token{RefreshToken: "refresh", Expiry: time.Now().Add(-time.Hour)}},
		{name: "empty token", token: &oauth2.Token{}, wantErr: true},
		{name: "expired without refresh", token: &oauth2.Token{AccessToken: "access", Expiry: time.Now().Add(-time.Hour)}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
