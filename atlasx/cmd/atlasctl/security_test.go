package main

import (
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestMirrorScanRejectsOutsideProfileDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	outsideDir := filepath.Join(t.TempDir(), "outside-profile")
	if err := os.MkdirAll(outsideDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	_, err := captureStdout(t, func() error {
		return run([]string{"mirror-scan", "--profile-dir", outsideDir})
	})
	if err == nil {
		t.Fatal("expected mirror-scan to reject outside profile dir")
	}
}

func TestImportChromeRejectsOutsideProfileDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	outsideDir := filepath.Join(t.TempDir(), "outside-chrome-profile")
	if err := os.MkdirAll(outsideDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	_, err := captureStdout(t, func() error {
		return run([]string{"import-chrome", "--source-profile-dir", outsideDir})
	})
	if err == nil {
		t.Fatal("expected import-chrome to reject outside profile dir")
	}
}

func TestMirrorScanAllowsAtlasXProfileDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	profileDir := filepath.Join(paths.ProfilesRoot, "webapp-isolated")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	output, err := captureStdout(t, func() error {
		return run([]string{"mirror-scan", "--profile-dir", profileDir})
	})
	if err != nil {
		t.Fatalf("expected mirror-scan to allow atlasx profile dir: %v", err)
	}
	if output == "" {
		t.Fatal("expected mirror-scan output")
	}
}
