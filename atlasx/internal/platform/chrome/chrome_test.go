package chrome

import (
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/macos"
)

func TestBuildLaunchArgs(t *testing.T) {
	args := BuildLaunchArgs("/Applications/Google Chrome.app", "https://chatgpt.com/atlas?get-started", "/tmp/atlasx-profile", false)
	if len(args) == 0 || args[0] != "-na" {
		t.Fatalf("unexpected isolated launch args: %#v", args)
	}
	shared := BuildLaunchArgs("/Applications/Google Chrome.app", "https://chatgpt.com/atlas?get-started", "", true)
	if len(shared) == 0 || shared[0] != "-a" {
		t.Fatalf("unexpected shared launch args: %#v", shared)
	}
}

func TestDetectWithPathsPrefersManagedRuntime(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	binaryPath := ManagedBinaryPath(paths)
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write binary failed: %v", err)
	}

	detection, err := DetectWithPaths("", paths)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}
	if detection.Source != "managed_auto" {
		t.Fatalf("unexpected source: %s", detection.Source)
	}
	if detection.BinaryPath != binaryPath {
		t.Fatalf("unexpected binary path: %s", detection.BinaryPath)
	}
}

func TestDetectWithPathsUsesManifestBinaryPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	sourceBundle := createFakeChromeBundle(t, "Google Chrome.app", "Google Chrome")
	if _, err := managedruntime.StageLocal(paths, managedruntime.StageOptions{
		BundlePath: sourceBundle,
		Version:    "123.0.0",
		Channel:    "local",
	}); err != nil {
		t.Fatalf("stage local failed: %v", err)
	}

	detection, err := DetectWithPaths("", paths)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}
	if detection.Source != "managed_auto" {
		t.Fatalf("unexpected source: %s", detection.Source)
	}
	if filepath.Base(detection.BinaryPath) != "Google Chrome" {
		t.Fatalf("unexpected binary path: %s", detection.BinaryPath)
	}
}

func createFakeChromeBundle(t *testing.T, bundleName string, binaryName string) string {
	t.Helper()

	bundlePath := filepath.Join(t.TempDir(), bundleName)
	binaryPath := filepath.Join(bundlePath, "Contents", "MacOS", binaryName)
	infoPath := filepath.Join(bundlePath, "Contents", "Info.plist")

	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write binary failed: %v", err)
	}
	if err := os.WriteFile(infoPath, []byte("<plist></plist>\n"), 0o644); err != nil {
		t.Fatalf("write plist failed: %v", err)
	}
	return bundlePath
}
