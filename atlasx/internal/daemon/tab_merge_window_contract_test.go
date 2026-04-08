package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabMergeWindowEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowMerge: tabs.WindowMergeResult{
			SourceWindowID: 9,
			TargetWindowID: 7,
			Returned:       1,
			MovedTargets: []tabs.WindowMergeTarget{
				{
					SourceTargetID:    "src-1",
					ActivatedTargetID: "dst-1",
					Target: tabs.Target{
						ID:    "new-1",
						Type:  "page",
						Title: "OpenAI",
						URL:   "https://openai.com",
					},
				},
			},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/merge-window", bytes.NewBufferString(`{"source_window_id":9,"target_window_id":7}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "source_window_id", "target_window_id", "returned", "moved_targets")
}
