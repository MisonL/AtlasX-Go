package managedruntime

import (
	"testing"

	"atlasx/internal/platform/macos"
)

func TestManifestSaveLoadAndStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	manifest := Manifest{
		Version:    "123.0.0",
		Channel:    "stable",
		SHA256:     "deadbeef",
		BundlePath: "/Applications/Chromium.app",
	}
	if err := SaveManifest(paths, manifest); err != nil {
		t.Fatalf("save manifest failed: %v", err)
	}

	loaded, err := LoadManifest(paths)
	if err != nil {
		t.Fatalf("load manifest failed: %v", err)
	}
	if loaded.Version != manifest.Version {
		t.Fatalf("unexpected version: %s", loaded.Version)
	}

	status, err := ManifestInfo(paths)
	if err != nil {
		t.Fatalf("manifest info failed: %v", err)
	}
	if !status.Present {
		t.Fatal("expected manifest present")
	}
	if status.BundlePath != manifest.BundlePath {
		t.Fatalf("unexpected bundle path: %s", status.BundlePath)
	}
}

func TestManifestInfoWithoutFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	status, err := ManifestInfo(paths)
	if err != nil {
		t.Fatalf("manifest info failed: %v", err)
	}
	if status.Present {
		t.Fatal("expected manifest to be absent")
	}
	if status.Path == "" {
		t.Fatal("expected manifest path")
	}
}
