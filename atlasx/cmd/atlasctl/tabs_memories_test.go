//go:build darwin

package main

import (
	"strings"
	"testing"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/tabs"
)

func TestTabsMemoriesOutputsRelevantSnippets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: "2026-04-07T14:00:00Z",
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
			CapturedAt: "2026-04-07T14:01:00Z",
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "memories", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs memories failed: %v", err)
	}
	for _, fragment := range []string{
		"returned=1",
		`snippet="qa_turn`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsMemoriesReturnsEmptyWhenNoMemoryExists(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas page",
			CapturedAt: "2026-04-07T14:01:00Z",
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "memories", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs memories failed: %v", err)
	}
	if !strings.Contains(output, "returned=0") {
		t.Fatalf("unexpected output: %s", output)
	}
	if strings.Contains(output, `snippet="`) {
		t.Fatalf("expected no snippets, got %s", output)
	}
}

func TestTabsMemoriesHidesSnippetsWhenPageVisibilityDisabled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: "2026-04-07T14:00:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "what is atlas memory",
		Answer:     "Atlas memory answer",
		TraceID:    "trace-1",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
		MemoryPageVisibility: settings.Bool(false),
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas memory retrieval page",
			CapturedAt: "2026-04-07T14:01:00Z",
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "memories", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs memories failed: %v", err)
	}
	if !strings.Contains(output, "returned=0") {
		t.Fatalf("unexpected output: %s", output)
	}
	if strings.Contains(output, `snippet="`) {
		t.Fatalf("expected hidden snippets, got %s", output)
	}
}

func TestTabsMemoriesFailurePrintsCaptureError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	context := tabs.PageContext{
		ID:           "tab-1",
		Title:        "Atlas",
		URL:          "https://chatgpt.com/atlas",
		CapturedAt:   "2026-04-07T14:01:00Z",
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
		return run([]string{"tabs", "memories", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs memories to fail")
	}
	if !strings.Contains(output, `capture_error="cdp error -32000: capture failed"`) {
		t.Fatalf("unexpected output: %s", output)
	}
}
