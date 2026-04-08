package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsPanelEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		openDevToolsPanel: tabs.Target{
			ID:    "devtools-window-1",
			Type:  "page",
			Title: "DevTools",
			URL:   "http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Ftab-1",
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/open-devtools-panel", bytes.NewBufferString(`{"id":"tab-1","panel":"network"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "id", "type", "title", "url")
}
