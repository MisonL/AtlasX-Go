package daemon

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atlasx/internal/imports"
	"atlasx/internal/managedruntime"
	"atlasx/internal/memory"
	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
	"atlasx/internal/tabs"
)

type stubTabsClient struct {
	openedURL                    string
	targets                      []tabs.Target
	listErr                      error
	searchTargets                []tabs.Target
	searchErr                    error
	windows                      []tabs.WindowSummary
	windowsErr                   error
	duplicateClose               tabs.CloseDuplicatesResult
	duplicateCloseErr            error
	windowOpen                   tabs.WindowOpenResult
	windowOpenErr                error
	windowMove                   tabs.WindowMoveResult
	windowMoveErr                error
	windowMoveByID               map[string]tabs.WindowMoveResult
	windowMoveNew                tabs.WindowMoveToNewResult
	windowMoveNewErr             error
	windowMoveNewByID            map[string]tabs.WindowMoveToNewResult
	windowMerge                  tabs.WindowMergeResult
	windowMergeErr               error
	windowActivate               tabs.WindowActivateResult
	windowActivateErr            error
	windowClose                  tabs.WindowCloseResult
	windowCloseErr               error
	windowState                  tabs.WindowBounds
	windowStateErr               error
	windowBounds                 tabs.WindowBounds
	windowBoundsErr              error
	openDevTools                 tabs.Target
	openDevToolsErr              error
	openDevToolsInWindow         tabs.WindowOpenResult
	openDevToolsInWindowErr      error
	openDevToolsPanel            tabs.Target
	openDevToolsPanelErr         error
	openDevToolsPanelInWindow    tabs.WindowOpenResult
	openDevToolsPanelInWindowErr error
	activatedTargetID            string
	activateErr                  error
	context                      tabs.PageContext
	captureErr                   error
	semanticContext              tabs.SemanticContext
	semanticErr                  error
	selection                    tabs.SelectionContext
	selectionErr                 error
	devTools                     tabs.DevToolsTarget
	devToolsErr                  error
	devToolsPanel                tabs.DevToolsTarget
	devToolsPanelErr             error
	deviceResult                 tabs.DeviceEmulationResult
	deviceErr                    error
}

func (s *stubTabsClient) List() ([]tabs.Target, error) {
	return s.targets, s.listErr
}

func (s *stubTabsClient) Search(query string) ([]tabs.Target, error) {
	if s.searchErr != nil {
		return nil, s.searchErr
	}
	return s.searchTargets, nil
}

func (s *stubTabsClient) Windows() ([]tabs.WindowSummary, error) {
	return s.windows, s.windowsErr
}

func (s *stubTabsClient) CloseDuplicates() (tabs.CloseDuplicatesResult, error) {
	if s.duplicateCloseErr != nil {
		return tabs.CloseDuplicatesResult{}, s.duplicateCloseErr
	}
	return s.duplicateClose, nil
}

func (s *stubTabsClient) OpenInWindow(windowID int, targetURL string) (tabs.WindowOpenResult, error) {
	if s.windowOpenErr != nil {
		return tabs.WindowOpenResult{}, s.windowOpenErr
	}
	return s.windowOpen, nil
}

func (s *stubTabsClient) MoveToWindow(targetID string, targetWindowID int) (tabs.WindowMoveResult, error) {
	if s.windowMoveErr != nil {
		return tabs.WindowMoveResult{}, s.windowMoveErr
	}
	if result, ok := s.windowMoveByID[targetID]; ok {
		return result, nil
	}
	return s.windowMove, nil
}

func (s *stubTabsClient) MoveToNewWindow(targetID string) (tabs.WindowMoveToNewResult, error) {
	if s.windowMoveNewErr != nil {
		return tabs.WindowMoveToNewResult{}, s.windowMoveNewErr
	}
	if result, ok := s.windowMoveNewByID[targetID]; ok {
		return result, nil
	}
	return s.windowMoveNew, nil
}

func (s *stubTabsClient) MergeWindow(sourceWindowID int, targetWindowID int) (tabs.WindowMergeResult, error) {
	if s.windowMergeErr != nil {
		return tabs.WindowMergeResult{}, s.windowMergeErr
	}
	return s.windowMerge, nil
}

func (s *stubTabsClient) ActivateWindow(windowID int) (tabs.WindowActivateResult, error) {
	if s.windowActivateErr != nil {
		return tabs.WindowActivateResult{}, s.windowActivateErr
	}
	return s.windowActivate, nil
}

func (s *stubTabsClient) CloseWindow(windowID int) (tabs.WindowCloseResult, error) {
	if s.windowCloseErr != nil {
		return tabs.WindowCloseResult{}, s.windowCloseErr
	}
	return s.windowClose, nil
}

func (s *stubTabsClient) SetWindowState(windowID int, state string) (tabs.WindowBounds, error) {
	if s.windowStateErr != nil {
		return tabs.WindowBounds{}, s.windowStateErr
	}
	return s.windowState, nil
}

func (s *stubTabsClient) SetWindowBounds(windowID int, left int, top int, width int, height int) (tabs.WindowBounds, error) {
	if s.windowBoundsErr != nil {
		return tabs.WindowBounds{}, s.windowBoundsErr
	}
	return s.windowBounds, nil
}

func (s *stubTabsClient) OpenDevToolsWindow(targetID string) (tabs.Target, error) {
	if s.openDevToolsErr != nil {
		return tabs.Target{}, s.openDevToolsErr
	}
	return s.openDevTools, nil
}

func (s *stubTabsClient) OpenDevToolsInWindow(targetID string, windowID int) (tabs.WindowOpenResult, error) {
	if s.openDevToolsInWindowErr != nil {
		return tabs.WindowOpenResult{}, s.openDevToolsInWindowErr
	}
	return s.openDevToolsInWindow, nil
}

func (s *stubTabsClient) OpenDevToolsPanelWindow(targetID string, panel string) (tabs.Target, error) {
	if s.openDevToolsPanelErr != nil {
		return tabs.Target{}, s.openDevToolsPanelErr
	}
	return s.openDevToolsPanel, nil
}

func (s *stubTabsClient) OpenDevToolsPanelInWindow(targetID string, panel string, windowID int) (tabs.WindowOpenResult, error) {
	if s.openDevToolsPanelInWindowErr != nil {
		return tabs.WindowOpenResult{}, s.openDevToolsPanelInWindowErr
	}
	return s.openDevToolsPanelInWindow, nil
}

func (s *stubTabsClient) Open(targetURL string) (tabs.Target, error) {
	s.openedURL = targetURL
	return tabs.Target{ID: "tab-1", URL: targetURL}, nil
}

func (s *stubTabsClient) OpenWindow(targetURL string) (tabs.Target, error) {
	s.openedURL = targetURL
	return tabs.Target{ID: "tab-window", Type: "page", Title: "OpenAI", URL: targetURL}, nil
}

func (s *stubTabsClient) Activate(targetID string) error {
	if s.activateErr != nil {
		return s.activateErr
	}
	s.activatedTargetID = targetID
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

func (s *stubTabsClient) CaptureSemanticContext(string) (tabs.SemanticContext, error) {
	return s.semanticContext, s.semanticErr
}

func (s *stubTabsClient) CaptureSelection(string) (tabs.SelectionContext, error) {
	return s.selection, s.selectionErr
}

func (s *stubTabsClient) DevTools(string) (tabs.DevToolsTarget, error) {
	return s.devTools, s.devToolsErr
}

func (s *stubTabsClient) DevToolsPanel(string, string) (tabs.DevToolsTarget, error) {
	return s.devToolsPanel, s.devToolsPanelErr
}

func (s *stubTabsClient) EmulateDevice(string, string) (tabs.DeviceEmulationResult, error) {
	return s.deviceResult, s.deviceErr
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

func TestHistoryOpenEndpointRejectsUnsupportedScheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	snapshot := mirror.Snapshot{
		HistoryRows: []mirror.HistoryEntry{
			{URL: "javascript:alert(1)"},
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

	if recorder.Code != http.StatusConflict {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if client.openedURL != "" {
		t.Fatalf("unexpected opened url: %s", client.openedURL)
	}
	if !strings.Contains(recorder.Body.String(), "unsupported url scheme") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
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

func TestDownloadsOpenEndpointRejectsUnsupportedScheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	snapshot := mirror.Snapshot{
		DownloadRows: []mirror.DownloadEntry{
			{TargetPath: "/tmp/file.zip", TabURL: "file:///tmp/file.zip"},
		},
	}
	if err := mirror.Save(paths, snapshot); err != nil {
		t.Fatalf("save mirror failed: %v", err)
	}

	client := &stubTabsClient{}
	restoreDaemonHooks(t, client)

	body := bytes.NewBufferString(`{"index":0}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/downloads/open", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if client.openedURL != "" {
		t.Fatalf("unexpected opened url: %s", client.openedURL)
	}
	if !strings.Contains(recorder.Body.String(), "unsupported url scheme") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
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

func TestBookmarksOpenEndpointRejectsUnsupportedScheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	importRoot := imports.DefaultChromeImportRoot(paths)
	if err := os.MkdirAll(importRoot, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	payload := `{"roots":{"bookmark_bar":{"children":[{"type":"url","name":"Chrome Settings","url":"chrome://settings"}]}}}`
	if err := os.WriteFile(filepath.Join(importRoot, "Bookmarks.json"), []byte(payload), 0o644); err != nil {
		t.Fatalf("write bookmarks failed: %v", err)
	}

	client := &stubTabsClient{}
	restoreDaemonHooks(t, client)

	body := bytes.NewBufferString(`{"index":0}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/bookmarks/open", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if client.openedURL != "" {
		t.Fatalf("unexpected opened url: %s", client.openedURL)
	}
	if !strings.Contains(recorder.Body.String(), "unsupported url scheme") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestTabsOpenEndpointRejectsUnsupportedScheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	client := &stubTabsClient{}
	restoreDaemonHooks(t, client)

	body := bytes.NewBufferString(`{"url":"file:///tmp/secret.txt"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/open", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if client.openedURL != "" {
		t.Fatalf("unexpected opened url: %s", client.openedURL)
	}
	if !strings.Contains(recorder.Body.String(), "unsupported url scheme") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestMirrorScanRejectsOutsideProfileDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	requestBody := `{"profile_dir":"` + filepath.Join(t.TempDir(), "outside-profile") + `"}`
	request := httptest.NewRequest(http.MethodPost, "/v1/mirror/scan", bytes.NewBufferString(requestBody))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestChromeImportRejectsOutsideProfileDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	requestBody := `{"source_profile_dir":"` + filepath.Join(t.TempDir(), "outside-profile") + `"}`
	request := httptest.NewRequest(http.MethodPost, "/v1/import/chrome", bytes.NewBufferString(requestBody))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
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

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	events, err := memory.Load(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 1 || events[0].Kind != memory.EventKindPageCapture {
		t.Fatalf("unexpected memory events: %+v", events)
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

func TestSidebarAskEndpointReturnsStructuredAnswer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body failed: %v", err)
		}
		if !bytes.Contains(body, []byte("Atlas context")) {
			t.Fatalf("tab context missing from request: %s", string(body))
		}
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Atlas answer"}}]}`))
	}))
	defer server.Close()

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
				BaseURL:   server.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-06T12:00:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
	})

	body := bytes.NewBufferString(`{"tab_id":"tab-1","question":"summarize this page"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/ask", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"answer":"Atlas answer"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"context_summary":"title=\"Atlas\"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"trace_id":"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}

	events, err := memory.Load(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 1 || events[0].Kind != memory.EventKindQATurn {
		t.Fatalf("unexpected memory events: %+v", events)
	}
	if events[0].Question != "summarize this page" || events[0].Answer != "Atlas answer" {
		t.Fatalf("unexpected memory event: %+v", events[0])
	}
	if len(events[0].CitedURLs) != 1 || events[0].CitedURLs[0] != "https://chatgpt.com/atlas" {
		t.Fatalf("unexpected memory event: %+v", events[0])
	}
}

func TestSidebarSummarizeEndpointReturnsStructuredSummary(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body failed: %v", err)
		}
		if !bytes.Contains(body, []byte(sidebar.PageSummaryQuestion)) {
			t.Fatalf("summary prompt missing from request: %s", string(body))
		}
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Atlas summary"}}]}`))
	}))
	defer server.Close()

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
				BaseURL:   server.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-06T12:00:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/summarize", bytes.NewBufferString(`{"tab_id":"tab-1"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"summary":"Atlas summary"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"trace_id":"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}

	events, err := memory.Load(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 1 || events[0].Kind != memory.EventKindQATurn {
		t.Fatalf("unexpected memory events: %+v", events)
	}
	if events[0].Question != sidebar.PageSummaryQuestion || events[0].Answer != "Atlas summary" {
		t.Fatalf("unexpected memory event: %+v", events[0])
	}
}

func TestSidebarAskEndpointAllowsProviderOverride(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENROUTER_API_KEY", "router-key")

	openAIServerCalled := false
	openAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		openAIServerCalled = true
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"wrong server"}}]}`))
	}))
	defer openAIServer.Close()

	openRouterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":"openai/gpt-5","choices":[{"message":{"content":"OpenRouter answer"}}]}`))
	}))
	defer openRouterServer.Close()

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
				BaseURL:   openAIServer.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
			{
				ID:        "backup",
				Provider:  "openrouter",
				Model:     "openai/gpt-5",
				BaseURL:   openRouterServer.URL,
				APIKeyEnv: "OPENROUTER_API_KEY",
			},
		},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:    "tab-1",
			Title: "Atlas",
			URL:   "https://chatgpt.com/atlas",
			Text:  "Atlas context",
		},
	})

	body := bytes.NewBufferString(`{"tab_id":"tab-1","question":"summarize this page","provider_id":"backup"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/ask", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"provider":"openrouter"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"answer":"OpenRouter answer"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if openAIServerCalled {
		t.Fatal("expected provider override to bypass default provider")
	}
}

func TestSidebarAskEndpointRejectsUnimplementedBackend(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("ANTHROPIC_API_KEY", "test-key")

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "anthropic",
				Model:     "claude-sonnet-4",
				BaseURL:   "https://api.anthropic.com/v1",
				APIKeyEnv: "ANTHROPIC_API_KEY",
			},
		},
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

func TestSidebarAskEndpointRejectsUnknownProviderID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

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
		},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{})

	body := bytes.NewBufferString(`{"tab_id":"tab-1","question":"summarize this page","provider_id":"missing"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/ask", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`provider id is not configured`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestSidebarAskEndpointSurfacesProviderFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"message":"upstream failed"}}`))
	}))
	defer server.Close()

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
				BaseURL:   server.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:    "tab-1",
			Title: "Atlas",
			URL:   "https://chatgpt.com/atlas",
			Text:  "Atlas context",
		},
	})

	body := bytes.NewBufferString(`{"tab_id":"tab-1","question":"summarize this page"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/ask", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`upstream failed`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"trace_id":"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestSidebarStatusEndpointIncludesRecentErrorState(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"message":"upstream failed"}}`))
	}))
	defer server.Close()

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
				BaseURL:   server.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:    "tab-1",
			Title: "Atlas",
			URL:   "https://chatgpt.com/atlas",
			Text:  "Atlas context",
		},
	})

	body := bytes.NewBufferString(`{"tab_id":"tab-1","question":"summarize this page"}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/ask", body)
	recorder := httptest.NewRecorder()
	NewMux(Status{}).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	statusRequest := httptest.NewRequest(http.MethodGet, "/v1/sidebar/status", nil)
	statusRecorder := httptest.NewRecorder()
	NewMux(Status{}).ServeHTTP(statusRecorder, statusRequest)
	if statusRecorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", statusRecorder.Code, statusRecorder.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(statusRecorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["timeout_ms"].(float64) <= 0 || payload["token_budget"].(float64) <= 0 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload["last_error"] == "" || payload["last_trace_id"] == "" {
		t.Fatalf("unexpected payload: %+v", payload)
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
