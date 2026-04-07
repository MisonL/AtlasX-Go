package main

import (
	"strings"
	"testing"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
)

func TestMemoryListOutputsSummaryAndEvents(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths := seedMemoryEventsForCommandTest(t)

	output, err := captureStdout(t, func() error {
		return run([]string{"memory", "list"})
	})
	if err != nil {
		t.Fatalf("run memory list failed: %v", err)
	}

	for _, fragment := range []string{
		"memory_root=" + paths.MemoryRoot,
		"events_file=" + paths.MemoryEventsFile,
		"present=true",
		"event_count=2",
		"returned=2",
		"kind=page_capture",
		"kind=qa_turn",
		`question="what is atlas"`,
		`answer="Atlas answer"`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestMemoryListLimitReturnsOnlyRecentEvents(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	seedMemoryEventsForCommandTest(t)

	output, err := captureStdout(t, func() error {
		return run([]string{"memory", "list", "--limit", "1"})
	})
	if err != nil {
		t.Fatalf("run memory list failed: %v", err)
	}
	if !strings.Contains(output, "returned=1") {
		t.Fatalf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "trace_id=trace-1") {
		t.Fatalf("unexpected output: %s", output)
	}
	if strings.Contains(output, "kind=page_capture") {
		t.Fatalf("expected output to exclude older events, got %s", output)
	}
}

func TestMemoryListWithoutEventsIsExplicitlyEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"memory", "list"})
	})
	if err != nil {
		t.Fatalf("run memory list failed: %v", err)
	}
	for _, fragment := range []string{
		"present=false",
		"event_count=0",
		"returned=0",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestMemorySearchOutputsRankedSnippets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	seedMemoryEventsForCommandTest(t)

	output, err := captureStdout(t, func() error {
		return run([]string{"memory", "search", "--tab-id", "tab-1", "--url", "https://chatgpt.com/atlas", "--limit", "1", "what", "is", "atlas"})
	})
	if err != nil {
		t.Fatalf("run memory search failed: %v", err)
	}
	for _, fragment := range []string{
		`question="what is atlas"`,
		"returned=1",
		`snippet="qa_turn`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestMemorySearchWithoutEventsReturnsEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"memory", "search", "what", "is", "atlas"})
	})
	if err != nil {
		t.Fatalf("run memory search failed: %v", err)
	}
	if !strings.Contains(output, "returned=0") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func seedMemoryEventsForCommandTest(t *testing.T) macos.Paths {
	t.Helper()

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := memory.AppendPageCapture(paths, memory.PageCaptureInput{
		OccurredAt: "2026-04-07T12:00:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
	}); err != nil {
		t.Fatalf("append page capture failed: %v", err)
	}
	if err := memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: "2026-04-07T12:01:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "what is atlas",
		Answer:     "Atlas answer",
		CitedURLs:  []string{"https://chatgpt.com/atlas"},
		TraceID:    "trace-1",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}
	return paths
}
