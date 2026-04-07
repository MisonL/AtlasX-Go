package daemon

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabDevToolsEndpointReturnsFrontendURL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		devTools: tabs.DevToolsTarget{
			ID:                  "tab-1",
			Title:               "Atlas",
			URL:                 "https://chatgpt.com/atlas",
			DevToolsFrontendURL: "http://127.0.0.1:9222/devtools/inspector.html?ws=127.0.0.1:9222/devtools/page/1",
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/devtools?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"devtools_frontend_url":"http://127.0.0.1:9222/devtools/inspector.html?ws=127.0.0.1:9222/devtools/page/1"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestTabDevToolsEndpointSurfacesLookupFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		devToolsErr: errors.New("target does not expose a devtools frontend url"),
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/devtools?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}
