package main

import (
	"strings"
	"testing"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/tabs"
)

func TestTabsAgentPlanOutputsReadOnlyPlan(t *testing.T) {
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

	restoreCommandTabsClient(t, &stubCommandTabsClient{
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

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "agent-plan", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs agent-plan failed: %v", err)
	}

	for _, fragment := range []string{
		"returned=5",
		"read_only=true",
		"executed=false",
		"memory_returned=1",
		"guardrail[0]=preview_only",
		"step_id=suggest-summarize_page",
		"step_id=recommend-related-tab-tab-2",
		"requires_confirmation=true",
		"rollback=read_only_preview_no_actions_executed",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsAgentPlanFailurePrintsCaptureError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	context := tabs.PageContext{
		ID:           "tab-1",
		Title:        "Atlas",
		URL:          "https://chatgpt.com/atlas",
		CapturedAt:   "2026-04-08T12:05:00Z",
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
		return run([]string{"tabs", "agent-plan", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs agent-plan to fail")
	}
	if !strings.Contains(output, `capture_error="cdp error -32000: capture failed"`) {
		t.Fatalf("unexpected output: %s", output)
	}
}
