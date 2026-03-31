package managedruntime

import (
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestResolveBundleBinaryPath(t *testing.T) {
	bundlePath := createFakeBundle(t, "Google Chrome.app", "Google Chrome")

	binaryPath, err := ResolveBundleBinaryPath(bundlePath)
	if err != nil {
		t.Fatalf("resolve bundle binary path failed: %v", err)
	}
	if filepath.Base(binaryPath) != "Google Chrome" {
		t.Fatalf("unexpected binary path: %s", binaryPath)
	}
}

func TestDetectManagedBinaryPathUsesManifestBinaryPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	bundlePath := createFakeBundle(t, "Google Chrome.app", "Google Chrome")
	binaryPath, err := ResolveBundleBinaryPath(bundlePath)
	if err != nil {
		t.Fatalf("resolve binary path failed: %v", err)
	}

	if err := SaveManifest(paths, Manifest{
		Version:    "123.0.0",
		Channel:    "local",
		BundlePath: bundlePath,
		BinaryPath: binaryPath,
	}); err != nil {
		t.Fatalf("save manifest failed: %v", err)
	}

	detectedPath, err := DetectManagedBinaryPath(paths)
	if err != nil {
		t.Fatalf("detect managed binary path failed: %v", err)
	}
	if detectedPath != binaryPath {
		t.Fatalf("unexpected detected path: %s", detectedPath)
	}
}

func TestResolveBundleBinaryPathRejectsMultipleExecutables(t *testing.T) {
	bundlePath := filepath.Join(t.TempDir(), "Multi.app")
	macosDir := filepath.Join(bundlePath, "Contents", "MacOS")
	if err := os.MkdirAll(macosDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(macosDir, "one"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write first binary failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(macosDir, "two"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write second binary failed: %v", err)
	}

	if _, err := ResolveBundleBinaryPath(bundlePath); err == nil {
		t.Fatal("expected multiple executables failure")
	}
}
