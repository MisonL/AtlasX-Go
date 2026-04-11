package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabSelectionEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		selection: tabs.SelectionContext{
			ID:                     "tab-1",
			Title:                  "Atlas",
			URL:                    "https://chatgpt.com/atlas",
			SelectionText:          "Atlas selected text",
			CapturedAt:             "2026-04-07T12:00:00Z",
			SelectionPresent:       true,
			SelectionTextTruncated: false,
			SelectionTextLength:    19,
			SelectionTextLimit:     1024,
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/selection?id=tab-1", nil)
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
		"selection_text",
		"captured_at",
		"selection_present",
		"selection_text_truncated",
		"selection_text_length",
		"selection_text_limit",
		"capture_error",
	)
}
