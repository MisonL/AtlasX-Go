package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabSearchEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		searchTargets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas Docs", URL: "https://openai.com/docs/atlas"},
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/search?q=atlas", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "returned", "targets")
}
