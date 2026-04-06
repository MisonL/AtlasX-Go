package daemon

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/imports"
	"atlasx/internal/managedruntime"
	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/tabs"
)

type stubTabsClient struct {
	openedURL  string
	context    tabs.PageContext
	captureErr error
}

func (s *stubTabsClient) List() ([]tabs.Target, error) {
	return nil, nil
}

func (s *stubTabsClient) Open(targetURL string) (tabs.Target, error) {
	s.openedURL = targetURL
	return tabs.Target{ID: "tab-1", URL: targetURL}, nil
}

func (s *stubTabsClient) Activate(string) error {
	return nil
}

func (s *stubTabsClient) Close(string) error {
	return nil
}

func (s *stubTabsClient) Navigate(string, string) error {
	return nil
}

func (s *stubTabsClient) Capture(string) (tabs.PageContext, error) {
	return s.context, s.captureErr
}

func TestHistoryOpenEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	snapshot := mirror.Snapshot{
		HistoryRows: []mirror.HistoryEntry{
			{URL: "https://example.com/history"},
		},
	}
	if err := mirror.Save(paths, snapshot); err != nil {
		t.Fatalf("save mirror failed: %v", err)
	}

	client := &stubTabsClient{}
	restoreDaemonHooks(t, client)

	body := bytes.NewBufferString(`{"index":0}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/history/open", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if client.openedURL != "https://example.com/history" {
		t.Fatalf("unexpected opened url: %s", client.openedURL)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["opened_history_index"].(float64) != 0 {
		t.Fatalf("unexpected response payload: %+v", payload)
	}
}

func TestDownloadsOpenEndpointRejectsEmptyTabURL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	snapshot := mirror.Snapshot{
		DownloadRows: []mirror.DownloadEntry{
			{TargetPath: "/tmp/file.zip"},
		},
	}
	if err := mirror.Save(paths, snapshot); err != nil {
		t.Fatalf("save mirror failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{})

	body := bytes.NewBufferString(`{"index":0}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/downloads/open", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestBookmarksOpenEndpointRejectsGet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	importRoot := imports.DefaultChromeImportRoot(paths)
	if err := os.MkdirAll(importRoot, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	payload := `{"roots":{"bookmark_bar":{"children":[{"type":"url","name":"OpenAI","url":"https://openai.com"}]}}}`
	if err := os.WriteFile(filepath.Join(importRoot, "Bookmarks.json"), []byte(payload), 0o644); err != nil {
		t.Fatalf("write bookmarks failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodGet, "/v1/bookmarks/open", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestTabContextEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-06T10:00:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/context?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"text":"Atlas context"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"captured_at":"2026-04-06T10:00:00Z"`)) {
		t.Fatalf("missing captured_at: %s", recorder.Body.String())
	}
}

func TestTabContextEndpointReturnsStructuredFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	context := tabs.PageContext{
		ID:           "tab-1",
		Title:        "Atlas",
		URL:          "https://chatgpt.com/atlas",
		CapturedAt:   "2026-04-06T10:00:00Z",
		TextLimit:    4096,
		CaptureError: "cdp error -32000: capture failed",
	}
	restoreDaemonHooks(t, &stubTabsClient{
		context: context,
		captureErr: &tabs.CaptureError{
			Context: context,
			Cause:   errors.New("cdp error -32000: capture failed"),
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/context?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["capture_error"] != "cdp error -32000: capture failed" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload["captured_at"] != "2026-04-06T10:00:00Z" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload["error"] != "cdp error -32000: capture failed" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestStatusEndpointIncludesRuntimeManifestDetails(t *testing.T) {
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

	request := httptest.NewRequest(http.MethodGet, "/v1/status", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["runtime_manifest_version"] != "123.0.0" {
		t.Fatalf("unexpected manifest version: %+v", payload)
	}
	if payload["runtime_manifest_channel"] != "local" {
		t.Fatalf("unexpected manifest channel: %+v", payload)
	}
	if payload["runtime_manifest_sha256"] != stageReport.SHA256 {
		t.Fatalf("unexpected manifest sha256: %+v", payload)
	}
	if payload["runtime_manifest_binary_path"] != stageReport.BinaryPath {
		t.Fatalf("unexpected manifest binary path: %+v", payload)
	}
}

func TestStatusEndpointLeavesRuntimeManifestDetailsEmptyWhenAbsent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/status", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["runtime_manifest_version"] != "" {
		t.Fatalf("expected empty manifest version: %+v", payload)
	}
	if payload["runtime_manifest_channel"] != "" {
		t.Fatalf("expected empty manifest channel: %+v", payload)
	}
	if payload["runtime_manifest_sha256"] != "" {
		t.Fatalf("expected empty manifest sha256: %+v", payload)
	}
	if payload["runtime_manifest_binary_path"] != "" {
		t.Fatalf("expected empty manifest binary path: %+v", payload)
	}
}

func TestRuntimeVerifyEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if _, err := managedruntime.StageLocal(paths, managedruntime.StageOptions{
		BundlePath: createDaemonFakeChromiumBundle(t),
		Version:    "123.0.0",
		Channel:    "local",
	}); err != nil {
		t.Fatalf("stage local failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/runtime/verify", bytes.NewBuffer(nil))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"verified":true`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestRuntimeVerifyEndpointRejectsMissingManifest(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodPost, "/v1/runtime/verify", bytes.NewBuffer(nil))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`managed runtime manifest is not present`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestRuntimeVerifyEndpointRejectsGet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/runtime/verify", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSidebarStatusEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	bootstrapConfig(t)

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodGet, "/v1/sidebar/status", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"configured":false`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"providers":[]`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestSidebarStatusEndpointIncludesProviderRegistry(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   "https://api.openai.com/v1",
				APIKeyEnv: "OPENAI_API_KEY",
			},
			{
				ID:        "backup",
				Provider:  "openrouter",
				Model:     "openai/gpt-5",
				BaseURL:   "https://openrouter.ai/api/v1",
				APIKeyEnv: "OPENROUTER_API_KEY",
			},
		},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodGet, "/v1/sidebar/status", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"default_provider":"primary"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"providers":[`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"api_key_env":"OPENAI_API_KEY"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestSidebarAskEndpointRejectsUnconfiguredBackend(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	bootstrapConfig(t)

	restoreDaemonHooks(t, &stubTabsClient{})

	body := bytes.NewBufferString(`{"tab_id":"tab-1","question":"summarize this page"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/ask", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSidebarAskEndpointRejectsUnimplementedBackend(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
		SidebarProvider: "openai",
		SidebarModel:    "gpt-5.4",
		SidebarBaseURL:  "https://api.openai.com/v1",
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{})

	body := bytes.NewBufferString(`{"tab_id":"tab-1","question":"summarize this page"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/ask", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotImplemented {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func restoreDaemonHooks(t *testing.T, client tabClient) {
	t.Helper()

	previousDiscoverPaths := discoverPaths
	previousNewTabsClient := newTabsClient

	discoverPaths = macos.DiscoverPaths
	newTabsClient = func(paths macos.Paths) (tabClient, error) {
		return client, nil
	}

	t.Cleanup(func() {
		discoverPaths = previousDiscoverPaths
		newTabsClient = previousNewTabsClient
	})
}

func bootstrapConfig(t *testing.T) {
	t.Helper()

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if _, err := settings.NewStore(paths.ConfigFile).Bootstrap(); err != nil {
		t.Fatalf("bootstrap config failed: %v", err)
	}
}
