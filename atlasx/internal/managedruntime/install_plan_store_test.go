package managedruntime

import (
	"errors"
	"os"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestInstallPlanSaveLoadAndStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	plan := mustInstallPlan(t)
	plan.CurrentPhase = InstallPhaseVerifying
	plan.LastError = "checksum retry pending"

	if err := SaveInstallPlan(paths, plan); err != nil {
		t.Fatalf("save install plan failed: %v", err)
	}

	loaded, err := LoadInstallPlan(paths)
	if err != nil {
		t.Fatalf("load install plan failed: %v", err)
	}
	if loaded.SourceURL != plan.SourceURL {
		t.Fatalf("unexpected source url: %+v", loaded)
	}

	status, err := InstallPlanInfo(paths)
	if err != nil {
		t.Fatalf("install plan info failed: %v", err)
	}
	if !status.Present {
		t.Fatalf("expected install plan present: %+v", status)
	}
	if status.CurrentPhase != InstallPhaseVerifying {
		t.Fatalf("unexpected phase: %+v", status)
	}
	if status.LastError != "checksum retry pending" {
		t.Fatalf("unexpected last error: %+v", status)
	}
}

func TestInstallPlanInfoWithoutFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	status, err := InstallPlanInfo(paths)
	if err != nil {
		t.Fatalf("install plan info failed: %v", err)
	}
	if status.Present {
		t.Fatalf("expected install plan absent: %+v", status)
	}
	if status.Path == "" {
		t.Fatalf("expected install plan path")
	}
}

func TestClearInstallPlan(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if err := SaveInstallPlan(paths, mustInstallPlan(t)); err != nil {
		t.Fatalf("save install plan failed: %v", err)
	}

	if err := ClearInstallPlan(paths); err != nil {
		t.Fatalf("clear install plan failed: %v", err)
	}

	if _, err := os.Stat(paths.RuntimeInstallPlanFile); !os.IsNotExist(err) {
		t.Fatalf("expected install plan removed, got: %v", err)
	}
}

func TestClearInstallPlanRejectsMissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	err = ClearInstallPlan(paths)
	if !errors.Is(err, ErrInstallPlanNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}
