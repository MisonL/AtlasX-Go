package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabOrganizeGroupIntoWindowReturnsMovedAndAlignedTargets(t *testing.T) {
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
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
			{
				WindowID: 9,
				Targets: []tabs.Target{
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
				},
			},
		},
		windowMoveByID: map[string]tabs.WindowMoveResult{
			"tab-2": {
				SourceWindowID:    9,
				TargetWindowID:    11,
				SourceTargetID:    "tab-2",
				ActivatedTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-2",
					Type:  "page",
					Title: "Atlas B",
					URL:   "https://chatgpt.com/atlas/b",
				},
			},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/organize-group-into-window", bytes.NewBufferString(`{"group_id":"host:chatgpt.com","window_id":11}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"group_id":"host:chatgpt.com"`, `"window_id":11`, `"aligned_targets"`, `"moved_targets"`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabOrganizeGroupIntoWindowRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/organize-group-into-window", bytes.NewBufferString(`{"group_id":"host:chatgpt.com","window_id":0}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}
