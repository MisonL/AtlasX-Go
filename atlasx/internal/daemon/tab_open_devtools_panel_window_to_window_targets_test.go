package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsPanelWindowToWindowsEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		openDevToolsPanelWindowToWindows: tabs.DevToolsPanelWindowToWindowsResult{
			SourceWindowID: 11,
			Panel:          "network",
			Returned:       1,
			OpenedTargets: []tabs.DevToolsWindowToWindowsTarget{
				{
					SourceTargetID: "src-1",
					Target: tabs.Target{
						ID:    "devtools-open-1",
						Type:  "page",
						Title: "DevTools",
						URL:   "http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1",
					},
				},
			},
		},
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/tabs/open-devtools-panel-window-to-windows",
		bytes.NewBufferString(`{"source_window_id":11,"panel":"network"}`),
	)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"source_window_id":11`, `"panel":"network"`, `"returned":1`, `"opened_targets":[`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected body to contain %q, got %s", fragment, recorder.Body.String())
		}
	}
}

func TestTabsOpenDevToolsPanelWindowToWindowsEndpointRejectsBlankPanel(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/tabs/open-devtools-panel-window-to-windows",
		bytes.NewBufferString(`{"source_window_id":11,"panel":" "}`),
	)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"error":"panel is required"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
