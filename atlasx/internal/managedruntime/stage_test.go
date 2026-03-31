package managedruntime

import (
	"os"
	"path/filepath"
	"testing"

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
	if manifest.BinaryPath == "" {
		t.Fatal("expected manifest binary path")
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

func TestStageLocalSupportsGoogleChromeBundle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	sourceBundle := createFakeBundle(t, "Google Chrome.app", "Google Chrome")
	report, err := StageLocal(paths, StageOptions{
		BundlePath: sourceBundle,
		Version:    "123.0.0",
		Channel:    "local",
	})
	if err != nil {
		t.Fatalf("stage local failed: %v", err)
	}
	if filepath.Base(report.BinaryPath) != "Google Chrome" {
		t.Fatalf("unexpected staged binary: %s", report.BinaryPath)
	}
}

func createFakeChromiumBundle(t *testing.T) string {
	t.Helper()

	return createFakeBundle(t, "Chromium.app", "Chromium")
}

func createFakeBundle(t *testing.T, bundleName string, binaryName string) string {
	t.Helper()

	bundlePath := filepath.Join(t.TempDir(), bundleName)
	binaryPath := filepath.Join(bundlePath, "Contents", "MacOS", binaryName)
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
