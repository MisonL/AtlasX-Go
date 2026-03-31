package managedruntime

import (
	"errors"
	"os"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestVerifySucceedsForStagedRuntime(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	stageReport, err := StageLocal(paths, StageOptions{
		BundlePath: createFakeChromiumBundle(t),
		Version:    "123.0.0",
		Channel:    "local",
	})
	if err != nil {
		t.Fatalf("stage local failed: %v", err)
	}

	report, err := Verify(paths)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !report.Verified {
		t.Fatalf("expected verified report: %+v", report)
	}
	if report.ActualSHA256 != stageReport.SHA256 {
		t.Fatalf("unexpected actual sha256: %+v", report)
	}
}

func TestVerifyRejectsMissingManifest(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	report, err := Verify(paths)
	if !errors.Is(err, ErrRuntimeManifestNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ManifestPresent {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestVerifyRejectsBinaryPathDrift(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if _, err := StageLocal(paths, StageOptions{
		BundlePath: createFakeChromiumBundle(t),
		Version:    "123.0.0",
		Channel:    "local",
	}); err != nil {
		t.Fatalf("stage local failed: %v", err)
	}

	manifest, err := LoadManifest(paths)
	if err != nil {
		t.Fatalf("load manifest failed: %v", err)
	}
	manifest.BinaryPath = manifest.BinaryPath + ".missing"
	if err := SaveManifest(paths, manifest); err != nil {
		t.Fatalf("save manifest failed: %v", err)
	}

	report, err := Verify(paths)
	if !errors.Is(err, ErrRuntimeBinaryNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.BinaryPresent {
		t.Fatalf("expected binary to be missing: %+v", report)
	}
}

func TestVerifyRejectsSHA256Mismatch(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	stageReport, err := StageLocal(paths, StageOptions{
		BundlePath: createFakeChromiumBundle(t),
		Version:    "123.0.0",
		Channel:    "local",
	})
	if err != nil {
		t.Fatalf("stage local failed: %v", err)
	}

	if err := os.WriteFile(stageReport.BinaryPath, []byte("#!/bin/sh\necho drift\n"), 0o755); err != nil {
		t.Fatalf("rewrite binary failed: %v", err)
	}

	report, err := Verify(paths)
	if !errors.Is(err, ErrRuntimeSHA256Mismatch) {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ActualSHA256 == report.ManifestSHA256 {
		t.Fatalf("expected sha256 mismatch: %+v", report)
	}
}
