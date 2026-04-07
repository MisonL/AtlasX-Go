package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		openDevTools: tabs.Target{
			ID:    "devtools-window-1",
			Type:  "page",
			Title: "DevTools",
			URL:   "http://127.0.0.1/devtools/inspector.html?ws=127.0.0.1/devtools/page/tab-1",
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/open-devtools", bytes.NewBufferString(`{"id":"tab-1"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"id":"devtools-window-1"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestTabsOpenDevToolsEndpointSurfacesFrontendErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		openDevToolsErr: errStringDaemon("target does not expose a devtools frontend url"),
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/open-devtools", bytes.NewBufferString(`{"id":"tab-1"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "target does not expose a devtools frontend url") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
