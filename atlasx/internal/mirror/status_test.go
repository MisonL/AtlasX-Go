package mirror

import (
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestScanWritesSuccessStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	profileDir := filepath.Join(t.TempDir(), "Default")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(profileDir, "Bookmarks"), []byte(`{"roots":{"bookmark_bar":{"children":[]}}}`), 0o644); err != nil {
		t.Fatalf("write bookmarks failed: %v", err)
	}

	if _, err := Scan(paths, profileDir); err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	status, err := LoadScanStatus(paths)
	if err != nil {
		t.Fatalf("load scan status failed: %v", err)
	}
	if status.Result != scanResultSucceeded || status.ProfileDir != profileDir {
		t.Fatalf("unexpected scan status: %+v", status)
	}
}

func TestScanWritesFailureStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	profileDir := filepath.Join(t.TempDir(), "Default")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(profileDir, "Bookmarks"), []byte(`{`), 0o644); err != nil {
		t.Fatalf("write bookmarks failed: %v", err)
	}

	if _, err := Scan(paths, profileDir); err == nil {
		t.Fatal("expected scan failure")
	}

	status, err := LoadScanStatus(paths)
	if err != nil {
		t.Fatalf("load scan status failed: %v", err)
	}
	if status.Result != scanResultFailed || status.Error == "" {
		t.Fatalf("unexpected scan status: %+v", status)
	}
}
