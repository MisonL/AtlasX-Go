package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsSetTitleEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		titleUpdate: tabs.TitleUpdateResult{
			ID:    "tab-1",
			Title: "Atlas Workbench",
			URL:   "https://openai.com/work",
		},
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/tabs/set-title",
		bytes.NewBufferString(`{"id":"tab-1","title":"Atlas Workbench"}`),
	)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "id", "title", "url")
}
