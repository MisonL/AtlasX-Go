package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/macos"
)

func TestRuntimeStatusEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/runtime/status", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"manifest_present":false`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"install_plan_present":false`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestRuntimeStatusEndpointIncludesInstallPlan(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	plan := mustDaemonInstallPlan(t)
	plan.CurrentPhase = managedruntime.InstallPhaseVerifying
	if err := managedruntime.SaveInstallPlan(paths, plan); err != nil {
		t.Fatalf("save install plan failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/runtime/status", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"install_plan_present":true`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"install_plan_phase":"verifying"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"install_plan_source_url":"https://example.com/chromium.zip"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestRuntimeStageEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	bundlePath := createDaemonFakeChromiumBundle(t)
	body := bytes.NewBufferString(`{"bundle_path":"` + bundlePath + `","version":"123.0.0","channel":"local"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/runtime/stage", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"version":"123.0.0"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestRuntimeStageEndpointRejectsInvalidBundle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	body := bytes.NewBufferString(`{"bundle_path":"/tmp/missing.app","version":"123.0.0"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/runtime/stage", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestRuntimeClearEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	bundlePath := createDaemonFakeChromiumBundle(t)
	stageBody := bytes.NewBufferString(`{"bundle_path":"` + bundlePath + `","version":"123.0.0","channel":"local"}`)
	stageRequest := httptest.NewRequest(http.MethodPost, "/v1/runtime/stage", stageBody)
	stageRecorder := httptest.NewRecorder()
	NewMux(Status{}).ServeHTTP(stageRecorder, stageRequest)
	if stageRecorder.Code != http.StatusOK {
		t.Fatalf("stage failed: %d body=%s", stageRecorder.Code, stageRecorder.Body.String())
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/runtime/clear", nil)
	recorder := httptest.NewRecorder()
	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	statusRequest := httptest.NewRequest(http.MethodGet, "/v1/runtime/status", nil)
	statusRecorder := httptest.NewRecorder()
	NewMux(Status{}).ServeHTTP(statusRecorder, statusRequest)
	if !bytes.Contains(statusRecorder.Body.Bytes(), []byte(`"manifest_present":false`)) {
		t.Fatalf("unexpected status body after clear: %s", statusRecorder.Body.String())
	}
}

func createDaemonFakeChromiumBundle(t *testing.T) string {
	t.Helper()

	bundlePath := filepath.Join(t.TempDir(), "Chromium.app")
	binaryPath := filepath.Join(bundlePath, "Contents", "MacOS", "Chromium")
	infoPath := filepath.Join(bundlePath, "Contents", "Info.plist")

	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(binaryPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write binary failed: %v", err)
	}
	if err := os.WriteFile(infoPath, []byte("<plist></plist>\n"), 0o644); err != nil {
		t.Fatalf("write plist failed: %v", err)
	}
	return bundlePath
}

func mustDaemonInstallPlan(t *testing.T) managedruntime.InstallPlan {
	t.Helper()

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
	return plan
}
