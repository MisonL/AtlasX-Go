package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsWindowIntoWindowEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		openDevToolsWindowIntoWindow: tabs.DevToolsWindowOpenResult{
			SourceWindowID: 11,
			TargetWindowID: 21,
			Returned:       1,
			OpenedTargets: []tabs.DevToolsWindowOpenTarget{
				{
					SourceTargetID:    "src-1",
					ActivatedTargetID: "dst-1",
					Target: tabs.Target{
						ID:    "devtools-open-1",
						Type:  "page",
						Title: "DevTools",
						URL:   "http://127.0.0.1/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1",
					},
				},
			},
		},
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/tabs/open-devtools-window-into-window",
		bytes.NewBufferString(`{"source_window_id":11,"target_window_id":21}`),
	)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "source_window_id", "target_window_id", "returned", "opened_targets")
}
