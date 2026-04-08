package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"atlasx/internal/defaultbrowser"
	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/macos"
)

func TestDefaultBrowserStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	previous := readDefaultBrowserStatus
	readDefaultBrowserStatus = func() (defaultbrowser.Status, error) {
		return defaultbrowser.Status{
			Source:        "launchservices",
			HTTPBundleID:  "org.mozilla.firefox",
			HTTPRole:      "all",
			HTTPKnown:     true,
			HTTPSBundleID: "org.mozilla.firefox",
			HTTPSRole:     "all",
			HTTPSKnown:    true,
			Consistent:    true,
		}, nil
	}
	t.Cleanup(func() {
		readDefaultBrowserStatus = previous
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"default-browser", "status"})
	})
	if err != nil {
		t.Fatalf("run default-browser status failed: %v", err)
	}

	assertContainsAll(t, output,
		"source=launchservices",
		"http_bundle_id=org.mozilla.firefox",
		"http_role=all",
		"https_bundle_id=org.mozilla.firefox",
		"https_role=all",
		"consistent=true",
	)
}

func TestLogsStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"logs", "status"})
	})
	if err != nil {
		t.Fatalf("run logs status failed: %v", err)
	}

	assertContainsAll(t, output,
		"logs_root=",
		"present=",
		"file_count=",
		"total_bytes=",
		"returned=",
	)
}

func TestRuntimeStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"runtime", "status"})
	})
	if err != nil {
		t.Fatalf("run runtime status failed: %v", err)
	}

	assertContainsAll(t, output,
		"runtime_root=",
		"manifest_present=",
		"install_plan_present=",
		"binary_executable=",
	)
}

func TestRuntimePlanStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	plan, err := managedruntime.NewInstallPlan(managedruntime.InstallPlanOptions{
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
	plan.CurrentPhase = managedruntime.InstallPhaseVerifying
	if err := managedruntime.SaveInstallPlan(paths, plan); err != nil {
		t.Fatalf("save install plan failed: %v", err)
	}

	output, err := captureStdout(t, func() error {
		return run([]string{"runtime", "plan", "status"})
	})
	if err != nil {
		t.Fatalf("run runtime plan status failed: %v", err)
	}

	assertContainsAll(t, output,
		"install_plan_present=true",
		"install_plan_source_url=https://example.com/chromium.zip",
		"install_plan_phase=verifying",
	)
}

func TestRuntimeVerifyContractWithoutManifest(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"runtime", "verify"})
	})
	if err == nil {
		t.Fatal("expected runtime verify to fail without manifest")
	}
	if !strings.Contains(err.Error(), "managed runtime manifest is not present") {
		t.Fatalf("unexpected runtime verify error: %v", err)
	}

	assertContainsAll(t, output,
		"manifest_present=false",
		"verified=false",
	)
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	previousStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe failed: %v", err)
	}

	os.Stdout = writer
	runErr := fn()
	_ = writer.Close()
	os.Stdout = previousStdout

	data, readErr := io.ReadAll(reader)
	_ = reader.Close()
	if readErr != nil {
		t.Fatalf("read captured stdout failed: %v", readErr)
	}
	return string(data), runErr
}

func assertContainsAll(t *testing.T, output string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, output=%s", fragment, output)
		}
	}
}
