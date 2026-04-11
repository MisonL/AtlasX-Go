//go:build darwin

package main

import (
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/macos"
)

func TestRuntimePlanResolveCommandCreatesInstallPlan(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	catalogPath := filepath.Join(t.TempDir(), "catalog.json")
	if err := os.WriteFile(catalogPath, []byte(`{
  "entries": [
    {
      "platform": "darwin-amd64",
      "channel": "stable",
      "version": "123.0.0",
      "url": "https://example.com/chromium-123.zip",
      "sha256": "deadbeef",
      "bundle_name": "Chromium.app"
    }
  ]
}
`), 0o644); err != nil {
		t.Fatalf("write catalog failed: %v", err)
	}

	output, err := captureStdout(t, func() error {
		return run([]string{
			"runtime", "plan", "resolve",
			"--catalog", catalogPath,
			"--version", "123.0.0",
			"--channel", "stable",
		})
	})
	if err != nil {
		t.Fatalf("run runtime plan resolve failed: %v", err)
	}

	assertContainsAll(t, output,
		"install_plan_present=true",
		"install_plan_bundle_name=Chromium.app",
		"install_plan_source_url=https://example.com/chromium-123.zip",
	)

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	plan, err := managedruntime.LoadInstallPlan(paths)
	if err != nil {
		t.Fatalf("load install plan failed: %v", err)
	}
	if plan.BundleName != "Chromium.app" {
		t.Fatalf("unexpected plan bundle name: %+v", plan)
	}
}

func TestRuntimePlanResolveCommandRejectsMissingVersionWithoutArchivePath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	catalogPath := filepath.Join(t.TempDir(), "catalog.json")
	if err := os.WriteFile(catalogPath, []byte(`{
  "entries": [
    {
      "platform": "darwin-amd64",
      "channel": "stable",
      "version": "123.0.0",
      "url": "https://example.com/chromium-123.zip",
      "sha256": "deadbeef",
      "bundle_name": "Chromium.app"
    }
  ]
}
`), 0o644); err != nil {
		t.Fatalf("write catalog failed: %v", err)
	}

	_, err := captureStdout(t, func() error {
		return run([]string{
			"runtime", "plan", "resolve",
			"--catalog", catalogPath,
			"--channel", "stable",
		})
	})
	if err == nil {
		t.Fatal("expected runtime plan resolve to fail without version")
	}
	if err.Error() != "runtime plan resolve requires --version when --archive-path is not provided" {
		t.Fatalf("unexpected error: %v", err)
	}
}
