package memory

import (
	"os"
	"strings"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestAppendAndLoadEvents(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if err := AppendPageCapture(paths, PageCaptureInput{
		OccurredAt: "2026-04-07T00:00:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
	}); err != nil {
		t.Fatalf("append page capture failed: %v", err)
	}
	if err := AppendQATurn(paths, QATurnInput{
		OccurredAt: "2026-04-07T00:01:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "summarize this page",
		Answer:     "Atlas answer",
		CitedURLs:  []string{"https://chatgpt.com/atlas"},
		TraceID:    "trace-1",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}

	events, err := Load(paths)
	if err != nil {
		t.Fatalf("load events failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("unexpected events: %+v", events)
	}
	if events[0].Kind != EventKindPageCapture || events[1].Kind != EventKindQATurn {
		t.Fatalf("unexpected event kinds: %+v", events)
	}
	if events[1].TraceID != "trace-1" || len(events[1].CitedURLs) != 1 {
		t.Fatalf("unexpected qa turn event: %+v", events[1])
	}

	data, err := os.ReadFile(paths.MemoryEventsFile)
	if err != nil {
		t.Fatalf("read events file failed: %v", err)
	}
	if !strings.Contains(string(data), "\"kind\":\"page_capture\"") || !strings.Contains(string(data), "\"kind\":\"qa_turn\"") {
		t.Fatalf("unexpected events file: %s", string(data))
	}
}

func TestLoadSummaryWithoutMemoryIsExplicitlyEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	summary, err := LoadSummary(paths)
	if err != nil {
		t.Fatalf("load summary failed: %v", err)
	}
	if summary.Present {
		t.Fatalf("expected memory absent: %+v", summary)
	}
	if summary.EventCount != 0 || summary.LastEventAt != "" || summary.LastEventKind != "" {
		t.Fatalf("unexpected empty summary: %+v", summary)
	}
	if summary.Root != paths.MemoryRoot || summary.EventsFile != paths.MemoryEventsFile {
		t.Fatalf("unexpected summary paths: %+v", summary)
	}
}
