package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabOrganizeWindowGroupIntoWindowReturnsResult(t *testing.T) {
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

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/organize-window-group-into-window", bytes.NewBufferString(`{"source_window_id":11,"group_id":"host:chatgpt.com","target_window_id":21}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"source_window_id":11`, `"target_window_id":21`, `"group_id":"host:chatgpt.com"`, `"window_id":21`, `"returned":2`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabOrganizeWindowGroupIntoWindowRejectsInvalidTargetWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/organize-window-group-into-window", bytes.NewBufferString(`{"source_window_id":11,"group_id":"host:chatgpt.com","target_window_id":0}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"error":"target_window_id must be positive"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
