package managedruntime

import (
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/platform/chrome"
	"atlasx/internal/platform/macos"
)

func TestStageLocalCopiesBundleAndWritesManifest(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	sourceBundle := createFakeChromiumBundle(t)
	report, err := StageLocal(paths, StageOptions{
		BundlePath: sourceBundle,
		Version:    "123.0.0",
		Channel:    "local",
	})
	if err != nil {
		t.Fatalf("stage local failed: %v", err)
	}

	if _, err := os.Stat(report.BinaryPath); err != nil {
		t.Fatalf("staged binary missing: %v", err)
	}

	manifest, err := LoadManifest(paths)
	if err != nil {
		t.Fatalf("load manifest failed: %v", err)
	}
	if manifest.Version != "123.0.0" {
		t.Fatalf("unexpected manifest version: %s", manifest.Version)
	}

	detection, err := chrome.DetectWithPaths("", paths)
	if err != nil {
		t.Fatalf("detect managed runtime failed: %v", err)
	}
	if detection.Source != "managed_auto" {
		t.Fatalf("unexpected detection source: %s", detection.Source)
	}
}

func TestStageLocalRejectsInvalidBundle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	invalidBundle := filepath.Join(t.TempDir(), "Broken.app")
	if err := os.MkdirAll(invalidBundle, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	if _, err := StageLocal(paths, StageOptions{
		BundlePath: invalidBundle,
		Version:    "123.0.0",
	}); err == nil {
		t.Fatal("expected invalid bundle failure")
	}
}

func createFakeChromiumBundle(t *testing.T) string {
	t.Helper()

	bundlePath := filepath.Join(t.TempDir(), "Chromium.app")
	binaryPath := filepath.Join(bundlePath, "Contents", "MacOS", "Chromium")
	resourcePath := filepath.Join(bundlePath, "Contents", "Info.plist")

	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write binary failed: %v", err)
	}
	if err := os.WriteFile(resourcePath, []byte("<plist></plist>\n"), 0o644); err != nil {
		t.Fatalf("write resource failed: %v", err)
	}
	return bundlePath
}
