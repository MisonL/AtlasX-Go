package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabSemanticContextEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		semanticContext: tabs.SemanticContext{
			ID:               "tab-1",
			Title:            "Atlas",
			URL:              "https://chatgpt.com/atlas",
			CapturedAt:       "2026-04-07T13:00:00Z",
			Returned:         0,
			HeadingsReturned: 0,
			LinksReturned:    0,
			FormsReturned:    0,
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/semantic-context?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"id",
		"title",
		"url",
		"captured_at",
		"returned",
		"headings_returned",
		"links_returned",
		"forms_returned",
		"headings",
		"links",
		"forms",
		"capture_error",
	)
}
