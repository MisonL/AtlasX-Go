package managedruntime

import (
	"errors"
	"testing"
)

func TestNewInstallPlanRejectsInvalidSourceURL(t *testing.T) {
	_, err := NewInstallPlan(InstallPlanOptions{
		Version:          "123.0.0",
		Channel:          "stable",
		SourceURL:        "http://example.com/chromium.zip",
		ExpectedSHA256:   "deadbeef",
		ArchivePath:      "/tmp/chromium.zip",
		StagedBundlePath: "/tmp/Chromium.app",
	})
	if !errors.Is(err, ErrInstallPlanSourceURLInvalid) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAdvanceInstallPlanHappyPath(t *testing.T) {
	plan := mustInstallPlan(t)

	var err error
	plan, err = AdvanceInstallPlan(plan, InstallEventStartDownload, "")
	if err != nil {
		t.Fatalf("start download failed: %v", err)
	}
	plan, err = AdvanceInstallPlan(plan, InstallEventFinishDownload, "")
	if err != nil {
		t.Fatalf("finish download failed: %v", err)
	}
	plan, err = AdvanceInstallPlan(plan, InstallEventStartVerify, "")
	if err != nil {
		t.Fatalf("start verify failed: %v", err)
	}
	plan, err = AdvanceInstallPlan(plan, InstallEventFinishVerify, "")
	if err != nil {
		t.Fatalf("finish verify failed: %v", err)
	}
	plan, err = AdvanceInstallPlan(plan, InstallEventStartStage, "")
	if err != nil {
		t.Fatalf("start stage failed: %v", err)
	}
	plan, err = AdvanceInstallPlan(plan, InstallEventFinishStage, "")
	if err != nil {
		t.Fatalf("finish stage failed: %v", err)
	}
	if plan.CurrentPhase != InstallPhaseStaged {
		t.Fatalf("unexpected final phase: %+v", plan)
	}
}

func TestAdvanceInstallPlanRollbackPath(t *testing.T) {
	plan := mustInstallPlan(t)

	var err error
	plan, err = AdvanceInstallPlan(plan, InstallEventStartDownload, "")
	if err != nil {
		t.Fatalf("start download failed: %v", err)
	}
	plan, err = AdvanceInstallPlan(plan, InstallEventFail, "download checksum precheck failed")
	if err != nil {
		t.Fatalf("fail transition failed: %v", err)
	}
	if plan.CurrentPhase != InstallPhaseFailed || plan.LastError == "" {
		t.Fatalf("unexpected failed plan: %+v", plan)
	}
	plan, err = AdvanceInstallPlan(plan, InstallEventStartRollback, "")
	if err != nil {
		t.Fatalf("start rollback failed: %v", err)
	}
	plan, err = AdvanceInstallPlan(plan, InstallEventFinishRollback, "")
	if err != nil {
		t.Fatalf("finish rollback failed: %v", err)
	}
	if plan.CurrentPhase != InstallPhaseRolledBack {
		t.Fatalf("unexpected rollback phase: %+v", plan)
	}
}

func TestAdvanceInstallPlanRejectsInvalidTransition(t *testing.T) {
	plan := mustInstallPlan(t)

	_, err := AdvanceInstallPlan(plan, InstallEventFinishVerify, "")
	if !errors.Is(err, ErrInstallPlanTransitionInvalid) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustInstallPlan(t *testing.T) InstallPlan {
	t.Helper()

	plan, err := NewInstallPlan(InstallPlanOptions{
		Version:          "123.0.0",
		Channel:          "stable",
		SourceURL:        "https://example.com/chromium.zip",
		ExpectedSHA256:   "deadbeef",
		ArchivePath:      "/tmp/chromium.zip",
		StagedBundlePath: "/tmp/Chromium.app",
	})
	if err != nil {
		t.Fatalf("new install plan failed: %v", err)
	}
	return plan
}
