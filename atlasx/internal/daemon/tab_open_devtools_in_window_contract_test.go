package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsInWindowEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		openDevToolsInWindow: tabs.WindowOpenResult{
			WindowID:          7,
			ActivatedTargetID: "tab-7",
			Target: tabs.Target{
				ID:    "devtools-window-tab",
				Type:  "page",
				Title: "DevTools",
				URL:   "http://127.0.0.1/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Ftab-1",
			},
		},
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/tabs/open-devtools-in-window",
		bytes.NewBufferString(`{"id":"tab-1","window_id":7}`),
	)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "window_id", "activated_target_id", "target")
}
