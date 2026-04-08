package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabOrganizeGroupToWindowReturnsMovedTargets(t *testing.T) {
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
	for _, fragment := range []string{`"group_id":"host:chatgpt.com"`, `"window_id":11`, `"source_target_id":"tab-2"`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabOrganizeGroupToWindowRejectsMissingGroupID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/organize-group-to-window", bytes.NewBufferString(`{}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestTabOrganizeGroupToWindowSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Solo", URL: "https://example.com"},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/organize-group-to-window", bytes.NewBufferString(`{"group_id":"host:chatgpt.com"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "group host:chatgpt.com not found") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
