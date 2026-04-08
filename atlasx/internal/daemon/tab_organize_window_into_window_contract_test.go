package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabOrganizeWindowIntoWindowEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
				},
			},
			{
				WindowID: 21,
				Targets: []tabs.Target{
					{ID: "dst-1", Type: "page", Title: "Workspace", URL: "https://workspace.example.com"},
				},
			},
		},
		windowMoveByID: map[string]tabs.WindowMoveResult{
			"tab-1": {
				SourceWindowID:    11,
				TargetWindowID:    21,
				SourceTargetID:    "tab-1",
				ActivatedTargetID: "dst-1",
				Target: tabs.Target{
					ID:    "new-1",
					Type:  "page",
					Title: "Atlas A",
					URL:   "https://chatgpt.com/atlas/a",
				},
			},
			"tab-2": {
				SourceWindowID:    11,
				TargetWindowID:    21,
				SourceTargetID:    "tab-2",
				ActivatedTargetID: "dst-1",
				Target: tabs.Target{
					ID:    "new-2",
					Type:  "page",
					Title: "Atlas B",
					URL:   "https://chatgpt.com/atlas/b",
				},
			},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/organize-window-into-window", bytes.NewBufferString(`{"source_window_id":11,"target_window_id":21}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "source_window_id", "target_window_id", "returned", "groups")
}
