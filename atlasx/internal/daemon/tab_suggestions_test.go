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

func TestTabSuggestionsReturnsStructuredSuggestions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: "2026-04-07T12:00:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "how does atlas memory work",
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
			CapturedAt: "2026-04-07T12:01:00Z",
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/suggestions?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	if payload["returned"].(float64) != 3 || payload["memory_returned"].(float64) != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	suggestions, ok := payload["suggestions"].([]any)
	if !ok || len(suggestions) != 3 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestTabSuggestionsFailureReturnsCaptureError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	context := tabs.PageContext{
		ID:           "tab-1",
		Title:        "Atlas",
		URL:          "https://chatgpt.com/atlas",
		CapturedAt:   "2026-04-07T12:05:00Z",
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

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/suggestions?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "capture failed") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

type errStringDaemon string

func (e errStringDaemon) Error() string {
	return string(e)
}
