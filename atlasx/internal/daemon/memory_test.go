package daemon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
)

func TestMemoryEndpointReturnsRecentEventsWithLimit(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := memory.AppendPageCapture(paths, memory.PageCaptureInput{
		OccurredAt: "2026-04-07T10:00:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas/1",
	}); err != nil {
		t.Fatalf("append page capture failed: %v", err)
	}
	if err := memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: "2026-04-07T10:01:00Z",
		TabID:      "tab-2",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas/2",
		Question:   "what changed",
		Answer:     "Atlas answer",
		TraceID:    "trace-2",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/memory?limit=1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	if payload["event_count"].(float64) != 2 || payload["returned"].(float64) != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	events, ok := payload["events"].([]any)
	if !ok || len(events) != 1 {
		t.Fatalf("unexpected events: %+v", payload)
	}
	event, ok := events[0].(map[string]any)
	if !ok || event["trace_id"] != "trace-2" {
		t.Fatalf("unexpected event: %+v", events[0])
	}
}

func TestMemoryEndpointRejectsInvalidLimit(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/memory?limit=bad", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "invalid limit") {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestMemorySearchEndpointReturnsRankedSnippets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := memory.AppendPageCapture(paths, memory.PageCaptureInput{
		OccurredAt: "2026-04-07T10:00:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
	}); err != nil {
		t.Fatalf("append page capture failed: %v", err)
	}
	if err := memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: "2026-04-07T10:01:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "how does memory retrieval work",
		Answer:     "Memory retrieval reuses prior Atlas context.",
		TraceID:    "trace-1",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/memory/search?question=memory%20retrieval&tab_id=tab-1&url=https://chatgpt.com/atlas&limit=1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	if payload["returned"].(float64) != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	snippets, ok := payload["snippets"].([]any)
	if !ok || len(snippets) != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	snippet, ok := snippets[0].(string)
	if !ok || !strings.Contains(snippet, "memory retrieval") {
		t.Fatalf("unexpected snippet: %+v", snippets[0])
	}
}

func TestMemorySearchEndpointRejectsEmptyQuestion(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/memory/search", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "question is required") {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}
