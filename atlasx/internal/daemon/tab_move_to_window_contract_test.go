package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabMoveToWindowEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowMove: tabs.WindowMoveResult{
			SourceWindowID:    9,
			TargetWindowID:    7,
			SourceTargetID:    "src-1",
			ActivatedTargetID: "dst-1",
			Target: tabs.Target{
				ID:    "new-1",
				Type:  "page",
				Title: "OpenAI",
				URL:   "https://openai.com",
			},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/move-to-window", bytes.NewBufferString(`{"id":"src-1","window_id":7}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "source_window_id", "target_window_id", "source_target_id", "activated_target_id", "target")
}
