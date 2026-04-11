package main

import (
	"errors"
	"testing"

	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/macos"
)

func TestRuntimeInstallCommandRendersReport(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	previousInstall := runManagedRuntimeInstall
	runManagedRuntimeInstall = func(paths macos.Paths) (managedruntime.InstallReport, error) {
		return managedruntime.InstallReport{
			InstallPlanPath:         paths.RuntimeInstallPlanFile,
			ArchivePath:             "/tmp/chromium.zip",
			ArchivePartPath:         "/tmp/chromium.zip.part",
			DownloadedArchiveSHA256: "deadbeef",
			ExtractedBundlePath:     "/tmp/extract/Chromium.app",
			StagedBundlePath:        paths.RuntimeRoot + "/Chromium.app",
			BinaryPath:              paths.RuntimeRoot + "/Chromium.app/Contents/MacOS/Chromium",
			ManifestPath:            paths.RuntimeManifestFile,
			Version:                 "123.0.0",
			Channel:                 "stable",
			CurrentPhase:            managedruntime.InstallPhaseStaged,
			Verified:                true,
		}, nil
	}
	t.Cleanup(func() {
		runManagedRuntimeInstall = previousInstall
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"runtime", "install"})
	})
	if err != nil {
		t.Fatalf("run runtime install failed: %v", err)
	}

	assertContainsAll(t, output,
		"archive=/tmp/chromium.zip",
		"current_phase=staged",
		"verified=true",
	)
}

func TestRuntimeInstallCommandSkipsEmptyReportOnFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	previousInstall := runManagedRuntimeInstall
	runManagedRuntimeInstall = func(paths macos.Paths) (managedruntime.InstallReport, error) {
		return managedruntime.InstallReport{}, errors.New("install failed")
	}
	t.Cleanup(func() {
		runManagedRuntimeInstall = previousInstall
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"runtime", "install"})
	})
	if err == nil {
		t.Fatal("expected runtime install to fail")
	}
	if err.Error() != "install failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Fatalf("expected no rendered report, got %q", output)
	}
}

func TestRuntimeVerifyCommandSkipsEmptyReportOnFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	previousVerify := runManagedRuntimeVerify
	runManagedRuntimeVerify = func(paths macos.Paths) (managedruntime.VerifyReport, error) {
		return managedruntime.VerifyReport{}, errors.New("verify failed")
	}
	t.Cleanup(func() {
		runManagedRuntimeVerify = previousVerify
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"runtime", "verify"})
	})
	if err == nil {
		t.Fatal("expected runtime verify to fail")
	}
	if err.Error() != "verify failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	if output != "" {
		t.Fatalf("expected no rendered report, got %q", output)
	}
}
