package main

import (
	"strings"
	"testing"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/tabs"
)

func TestTabsRecommendContextOutputsStructuredRecommendations(t *testing.T) {
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

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas memory retrieval page",
			CapturedAt: "2026-04-07T12:01:00Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
			{ID: "tab-2", Type: "page", Title: "Atlas docs", URL: "https://chatgpt.com/docs"},
			{ID: "tab-3", Type: "page", Title: "Elsewhere", URL: "https://example.com/other"},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "recommend-context", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs recommend-context failed: %v", err)
	}

	for _, fragment := range []string{
		"returned=2",
		"memory_returned=1",
		"recommendation_id=related-tab-tab-2",
		"recommendation_id=memory-relevant-qa-turn",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsRecommendContextReturnsEmptyRecommendations(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "about:blank",
			Text:       "Blank page",
			CapturedAt: "2026-04-07T12:01:00Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "about:blank"},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "recommend-context", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs recommend-context failed: %v", err)
	}
	if !strings.Contains(output, "returned=0") || !strings.Contains(output, "memory_returned=0") {
		t.Fatalf("unexpected output: %s", output)
	}
	if strings.Contains(output, "recommendation_id=") {
		t.Fatalf("expected no recommendations, got %s", output)
	}
}

func TestTabsRecommendContextFailurePrintsCaptureError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	context := tabs.PageContext{
		ID:           "tab-1",
		Title:        "Atlas",
		URL:          "https://chatgpt.com/atlas",
		CapturedAt:   "2026-04-07T12:05:00Z",
		TextLimit:    4096,
		CaptureError: "cdp error -32000: capture failed",
	}
	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: context,
		captureErr: &tabs.CaptureError{
			Context: context,
			Cause:   errString("cdp error -32000: capture failed"),
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "recommend-context", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs recommend-context to fail")
	}
	if !strings.Contains(output, `capture_error="cdp error -32000: capture failed"`) {
		t.Fatalf("unexpected output: %s", output)
	}
}
