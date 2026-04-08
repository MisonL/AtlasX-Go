package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabOrganizeIntoWindowReturnsGroups(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
			{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
			{ID: "tab-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
			{ID: "tab-4", Type: "page", Title: "Build Log - B", URL: "about:blank"},
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
			{
				WindowID: 7,
				Targets: []tabs.Target{
					{ID: "tab-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
					{ID: "tab-4", Type: "page", Title: "Build Log - B", URL: "about:blank"},
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
			"tab-3": {
				SourceWindowID:    7,
				TargetWindowID:    11,
				SourceTargetID:    "tab-3",
				ActivatedTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-3",
					Type:  "page",
					Title: "Build Log - A",
					URL:   "about:blank",
				},
			},
			"tab-4": {
				SourceWindowID:    7,
				TargetWindowID:    11,
				SourceTargetID:    "tab-4",
				ActivatedTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-4",
					Type:  "page",
					Title: "Build Log - B",
					URL:   "about:blank",
				},
			},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/organize-into-window", bytes.NewBufferString(`{"window_id":11}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"returned":2`, `"group_id":"host:chatgpt.com"`, `"group_id":"title:build log"`, `"aligned_targets":[`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabOrganizeIntoWindowRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/organize-into-window", bytes.NewBufferString(`{"window_id":0}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"error":"window_id must be positive"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
