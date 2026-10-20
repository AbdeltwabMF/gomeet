//go:build windows

package platform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteStartupLauncher(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), startupLauncherFile)
	if err := writeStartupLauncher(path); err != nil {
		t.Fatalf("writeStartupLauncher() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read startup launcher: %v", err)
	}
	if string(content) != startupLauncherScript {
		t.Fatalf("startup launcher = %q, want %q", content, startupLauncherScript)
	}
}
