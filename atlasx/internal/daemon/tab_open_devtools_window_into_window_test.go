package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsWindowIntoWindowEndpoint(t *testing.T) {
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
	for _, fragment := range []string{`"source_window_id":11`, `"target_window_id":21`, `"opened_targets":[`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected body to contain %q, got %s", fragment, recorder.Body.String())
		}
	}
}

func TestTabsOpenDevToolsWindowIntoWindowEndpointRejectsInvalidSourceWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/tabs/open-devtools-window-into-window",
		bytes.NewBufferString(`{"source_window_id":0,"target_window_id":21}`),
	)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"error":"source_window_id must be positive"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
