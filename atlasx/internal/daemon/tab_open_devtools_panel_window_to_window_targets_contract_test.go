package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsPanelWindowToWindowsEndpointContract(t *testing.T) {
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

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "source_window_id", "panel", "returned", "opened_targets")
}
