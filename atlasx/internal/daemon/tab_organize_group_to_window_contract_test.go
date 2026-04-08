package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabOrganizeGroupToWindowEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
			{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
		},
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "new-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
		},
		windowMoveNew: tabs.WindowMoveToNewResult{
			SourceWindowID: 9,
			SourceTargetID: "tab-1",
			Target: tabs.Target{
				ID:    "new-1",
				Type:  "page",
				Title: "Atlas A",
				URL:   "https://chatgpt.com/atlas/a",
			},
		},
		windowMove: tabs.WindowMoveResult{
			SourceWindowID:    9,
			TargetWindowID:    11,
			SourceTargetID:    "tab-2",
			ActivatedTargetID: "new-1",
			Target: tabs.Target{
				ID:    "new-2",
				Type:  "page",
				Title: "Atlas B",
				URL:   "https://chatgpt.com/atlas/b",
			},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/organize-group-to-window", bytes.NewBufferString(`{"group_id":"host:chatgpt.com"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "group_id", "label", "window_id", "returned", "moved_targets")
}
