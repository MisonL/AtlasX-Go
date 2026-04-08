package daemon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/tabs"
)

func TestTabAgentPlanReturnsStructuredPreview(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: "2026-04-08T12:00:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "what is atlas memory",
		Answer:     "Atlas memory answer",
		TraceID:    "trace-1",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas memory retrieval page",
			CapturedAt: "2026-04-08T12:01:00Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
			{ID: "tab-2", Type: "page", Title: "Atlas docs", URL: "https://chatgpt.com/docs"},
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/agent-plan?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	for _, fragment := range []string{
		`"read_only":true`,
		`"executed":false`,
		`"rollback":"read_only_preview_no_actions_executed"`,
		`"requires_confirmation":true`,
		`"id":"suggest-summarize_page"`,
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected body to contain %q, got %s", fragment, body)
		}
	}
}

func TestTabAgentPlanFailureReturnsCaptureError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	context := tabs.PageContext{
		ID:           "tab-1",
		Title:        "Atlas",
		URL:          "https://chatgpt.com/atlas",
		CapturedAt:   "2026-04-08T12:05:00Z",
		TextLimit:    4096,
		CaptureError: "cdp error -32000: capture failed",
	}
	restoreDaemonHooks(t, &stubTabsClient{
		context: context,
		captureErr: &tabs.CaptureError{
			Context: context,
			Cause:   errStringDaemon("cdp error -32000: capture failed"),
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/agent-plan?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "capture failed") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
