package googlecal

import (
	"os"
	"path/filepath"
	"testing"

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
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("token permissions = %o, want 600", perm)
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
