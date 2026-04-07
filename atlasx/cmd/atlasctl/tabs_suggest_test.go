package main

import (
	"strings"
	"testing"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/tabs"
)

func TestTabsSuggestOutputsStructuredSuggestions(t *testing.T) {
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
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "suggest", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs suggest failed: %v", err)
	}
	for _, fragment := range []string{
		"returned=3",
		"memory_returned=1",
		"suggestion_id=summarize_page",
		"suggestion_id=relate_memory",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsSuggestFailurePrintsCaptureError(t *testing.T) {
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
		return run([]string{"tabs", "suggest", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs suggest to fail")
	}
	if !strings.Contains(output, `capture_error="cdp error -32000: capture failed"`) {
		t.Fatalf("unexpected output: %s", output)
	}
}

type errString string

func (e errString) Error() string {
	return string(e)
}
