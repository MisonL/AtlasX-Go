package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"atlasx/internal/launcher"
	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/macos"
	"atlasx/internal/profile"
)

func TestBootstrapIncludesRuntimeBinaryState(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	sourceBundle := createDaemonFakeChromiumBundle(t)
	stageReport, err := managedruntime.StageLocal(paths, managedruntime.StageOptions{
		BundlePath: sourceBundle,
		Version:    "123.0.0",
		Channel:    "local",
	})
	if err != nil {
		t.Fatalf("stage local failed: %v", err)
	}

	status, err := Bootstrap()
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	if !status.RuntimeBundlePresent || !status.RuntimeBinaryPresent || !status.RuntimeBinaryExecutable {
		t.Fatalf("unexpected runtime state: %+v", status)
	}
	if status.RuntimeManifestVersion != "123.0.0" || status.RuntimeManifestChannel != "local" {
		t.Fatalf("unexpected manifest version/channel: %+v", status)
	}
	if status.RuntimeManifestSHA256 != stageReport.SHA256 {
		t.Fatalf("unexpected manifest sha256: %+v", status)
	}
	if status.RuntimeManifestBinaryPath != stageReport.BinaryPath {
		t.Fatalf("unexpected manifest binary path: %+v", status)
	}
}

func TestBootstrapMarksStaleManagedSessionAsNotLive(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if err := launcher.SaveState(paths, launcher.State{
		Mode:        profile.ModeIsolated,
		Managed:     true,
		UserDataDir: filepath.Join(t.TempDir(), "profile"),
		StartedAt:   time.Now().Add(-10 * time.Second).UTC().Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("save state failed: %v", err)
	}

	status, err := Bootstrap()
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	if status.ManagedSessionLive {
		t.Fatalf("expected stale managed session to be not live: %+v", status)
	}
	if !status.ManagedSessionStale || !status.ManagedSessionStateCleaned {
		t.Fatalf("expected stale session cleanup state: %+v", status)
	}
	if _, err := os.Stat(paths.SessionFile); !os.IsNotExist(err) {
		t.Fatalf("expected stale session file removed, got: %v", err)
	}
}
