package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"atlasx/internal/imports"
	"atlasx/internal/launcher"
	"atlasx/internal/managedruntime"
	"atlasx/internal/mirror"
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

func TestBootstrapIncludesMirrorAndImportRefreshStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if err := mirror.SaveScanStatus(paths, mirror.ScanStatus{
		GeneratedAt: "2026-04-06T00:00:00Z",
		ProfileDir:  "/tmp/profile",
		Result:      "failed",
		Error:       "boom",
	}); err != nil {
		t.Fatalf("save mirror status failed: %v", err)
	}
	if err := imports.SaveChromeImportStatus(paths, imports.OperationStatus{
		GeneratedAt: "2026-04-06T01:00:00Z",
		Source:      "/tmp/chrome-default",
		Result:      "succeeded",
	}); err != nil {
		t.Fatalf("save chrome import status failed: %v", err)
	}
	if err := imports.SaveSafariImportStatus(paths, imports.OperationStatus{
		GeneratedAt: "2026-04-06T02:00:00Z",
		Source:      "/tmp/Bookmarks.plist",
		Result:      "failed",
		Error:       "missing",
	}); err != nil {
		t.Fatalf("save safari import status failed: %v", err)
	}

	status, err := Bootstrap()
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	if status.MirrorLastScanAt != "2026-04-06T00:00:00Z" || status.MirrorLastScanError != "boom" {
		t.Fatalf("unexpected mirror refresh status: %+v", status)
	}
	if status.ChromeImportLastAt != "2026-04-06T01:00:00Z" || status.ChromeImportLastResult != "succeeded" {
		t.Fatalf("unexpected chrome import refresh status: %+v", status)
	}
	if status.SafariImportLastAt != "2026-04-06T02:00:00Z" || status.SafariImportLastError != "missing" {
		t.Fatalf("unexpected safari import refresh status: %+v", status)
	}
}
