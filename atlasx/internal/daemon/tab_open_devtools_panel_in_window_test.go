package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsPanelInWindowEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		openDevToolsPanelInWindow: tabs.WindowOpenResult{
			WindowID:          7,
			ActivatedTargetID: "tab-7",
			Target: tabs.Target{
				ID:    "devtools-window-tab",
				Type:  "page",
				Title: "DevTools",
				URL:   "http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Ftab-1",
			},
		},
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/tabs/open-devtools-panel-in-window",
		bytes.NewBufferString(`{"id":"tab-1","panel":"network","window_id":7}`),
	)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"window_id":7`, `"activated_target_id":"tab-7"`, `"id":"devtools-window-tab"`, `panel=network`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected body to contain %q, got %s", fragment, recorder.Body.String())
		}
	}
}

func TestTabsOpenDevToolsPanelInWindowEndpointRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/tabs/open-devtools-panel-in-window",
		bytes.NewBufferString(`{"id":"tab-1","panel":"network","window_id":0}`),
	)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"error":"window_id must be positive"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
