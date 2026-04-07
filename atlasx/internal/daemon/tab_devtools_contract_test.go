package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabDevToolsEndpointContract(t *testing.T) {
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

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "id", "title", "url", "devtools_frontend_url")
}
