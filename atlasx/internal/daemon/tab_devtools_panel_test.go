package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabDevToolsPanelEndpointReturnsFrontendURL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		devToolsPanel: tabs.DevToolsTarget{
			ID:                  "tab-1",
			Title:               "Atlas",
			URL:                 "https://chatgpt.com/atlas",
			DevToolsFrontendURL: "http://127.0.0.1:9222/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2F1",
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/devtools-panel?id=tab-1&panel=network", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"devtools_frontend_url":"http://127.0.0.1:9222/devtools/inspector.html?panel=network\u0026ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2F1"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestTabDevToolsPanelEndpointRejectsMissingPanel(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/devtools-panel?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}
