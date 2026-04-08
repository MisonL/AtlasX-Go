package updates

import (
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/macos"
)

func TestLoadStatusReturnsEmptyWhenRuntimeAbsent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	status, err := LoadStatus(paths)
	if err != nil {
		t.Fatalf("load status failed: %v", err)
	}

	if status.ManifestPresent || status.StagedReady || status.PlanPresent || status.PlanPending || status.PlanInFlight {
		t.Fatalf("unexpected empty status: %+v", status)
	}
}

func TestLoadStatusIncludesStagedRuntimeAndPendingPlan(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	bundlePath := createFakeChromiumBundle(t)
	if _, err := managedruntime.StageLocal(paths, managedruntime.StageOptions{
		BundlePath: bundlePath,
		Version:    "123.0.0",
		Channel:    "stable",
	}); err != nil {
		t.Fatalf("stage local failed: %v", err)
	}

	plan, err := managedruntime.NewInstallPlan(managedruntime.InstallPlanOptions{
		Version:          "124.0.0",
		Channel:          "stable",
		SourceURL:        "https://example.com/chromium-124.zip",
		ExpectedSHA256:   "deadbeef",
		ArchivePath:      "/tmp/chromium-124.zip",
		StagedBundlePath: "/tmp/Chromium.app",
	})
	if err != nil {
		t.Fatalf("new install plan failed: %v", err)
	}
	plan.CurrentPhase = managedruntime.InstallPhaseVerifying
	if err := managedruntime.SaveInstallPlan(paths, plan); err != nil {
		t.Fatalf("save install plan failed: %v", err)
	}

	status, err := LoadStatus(paths)
	if err != nil {
		t.Fatalf("load status failed: %v", err)
	}

	if !status.StagedReady || status.StagedVersion != "123.0.0" {
		t.Fatalf("unexpected staged status: %+v", status)
	}
	if !status.PlanPresent || status.PlanVersion != "124.0.0" || status.PlanPhase != managedruntime.InstallPhaseVerifying {
		t.Fatalf("unexpected plan status: %+v", status)
	}
	if !status.PlanPending || !status.PlanInFlight {
		t.Fatalf("expected pending in-flight plan: %+v", status)
	}
}

func createFakeChromiumBundle(t *testing.T) string {
	t.Helper()

	root := filepath.Join(t.TempDir(), "Chromium.app", "Contents", "MacOS")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir fake bundle failed: %v", err)
	}

	binaryPath := filepath.Join(root, "Chromium")
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write fake binary failed: %v", err)
	}
	return filepath.Dir(filepath.Dir(root))
}
