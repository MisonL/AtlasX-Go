package managedruntime

import (
	"errors"
	"os"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestStatusAndClear(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	sourceBundle := createFakeChromiumBundle(t)
	if _, err := StageLocal(paths, StageOptions{
		BundlePath: sourceBundle,
		Version:    "123.0.0",
		Channel:    "local",
	}); err != nil {
		t.Fatalf("stage local failed: %v", err)
	}

	status, err := Status(paths)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !status.BundlePresent || !status.BinaryExecutable {
		t.Fatalf("unexpected status: %+v", status)
	}
	if status.ManifestSHA256 == "" {
		t.Fatalf("expected manifest sha256 in status: %+v", status)
	}
	if status.InstallPlanPresent {
		t.Fatalf("did not expect install plan in status: %+v", status)
	}
	if status.InstallPlanPath == "" {
		t.Fatalf("expected install plan path in status: %+v", status)
	}

	if err := Clear(paths); err != nil {
		t.Fatalf("clear failed: %v", err)
	}

	if _, err := os.Stat(status.StagedBundlePath); !os.IsNotExist(err) {
		t.Fatalf("expected bundle removed, got: %v", err)
	}
	if _, err := os.Stat(paths.RuntimeManifestFile); !os.IsNotExist(err) {
		t.Fatalf("expected manifest removed, got: %v", err)
	}
}

func TestStatusIncludesInstallPlan(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	plan := mustInstallPlan(t)
	plan.CurrentPhase = InstallPhaseDownloaded
	if err := SaveInstallPlan(paths, plan); err != nil {
		t.Fatalf("save install plan failed: %v", err)
	}

	status, err := Status(paths)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !status.InstallPlanPresent {
		t.Fatalf("expected install plan in status: %+v", status)
	}
	if status.InstallPlanPhase != InstallPhaseDownloaded {
		t.Fatalf("unexpected install plan phase: %+v", status)
	}
	if status.InstallPlanSourceURL != plan.SourceURL {
		t.Fatalf("unexpected install plan source url: %+v", status)
	}
	if status.InstallPlanArchivePath != plan.ArchivePath {
		t.Fatalf("unexpected install plan archive path: %+v", status)
	}
}

func TestClearRejectsMissingRuntime(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	err = Clear(paths)
	if !errors.Is(err, ErrStagedRuntimeNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}
