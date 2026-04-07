package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/tabs"
)

type stubCommandTabsClient struct {
	context    tabs.PageContext
	captureErr error
}

func (s *stubCommandTabsClient) List() ([]tabs.Target, error) {
	return nil, nil
}

func (s *stubCommandTabsClient) Open(string) (tabs.Target, error) {
	return tabs.Target{}, nil
}

func (s *stubCommandTabsClient) Activate(string) error {
	return nil
}

func (s *stubCommandTabsClient) Close(string) error {
	return nil
}

func (s *stubCommandTabsClient) Navigate(string, string) error {
	return nil
}

func (s *stubCommandTabsClient) Capture(string) (tabs.PageContext, error) {
	return s.context, s.captureErr
}

func TestTabsCaptureWritesMemoryEvent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-07T08:00:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "capture", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs capture failed: %v", err)
	}
	if !strings.Contains(output, `captured_at=2026-04-07T08:00:00Z`) {
		t.Fatalf("unexpected output: %s", output)
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	events, err := memory.Load(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("unexpected memory event count: %+v", events)
	}
	if events[0].Kind != memory.EventKindPageCapture {
		t.Fatalf("unexpected memory event: %+v", events[0])
	}
	if events[0].OccurredAt != "2026-04-07T08:00:00Z" || events[0].URL != "https://chatgpt.com/atlas" {
		t.Fatalf("unexpected memory event: %+v", events[0])
	}
}

func TestTabsCaptureFailureDoesNotWriteMemoryEvent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	context := tabs.PageContext{
		ID:           "tab-1",
		Title:        "Atlas",
		URL:          "https://chatgpt.com/atlas",
		CapturedAt:   "2026-04-07T08:00:00Z",
		TextLimit:    4096,
		CaptureError: "cdp error -32000: capture failed",
	}
	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: context,
		captureErr: &tabs.CaptureError{
			Context: context,
			Cause:   errors.New("cdp error -32000: capture failed"),
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "capture", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs capture to fail")
	}
	if !strings.Contains(output, `capture_error="cdp error -32000: capture failed"`) {
		t.Fatalf("unexpected output: %s", output)
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	events, err := loadCommandMemoryEvents(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no memory events, got %+v", events)
	}
}

func TestTabsCaptureFailsWhenMemoryWriteFails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			CapturedAt: "2026-04-07T08:00:00Z",
		},
	})

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(paths.MemoryRoot), 0o755); err != nil {
		t.Fatalf("mkdir memory parent failed: %v", err)
	}
	if err := os.WriteFile(paths.MemoryRoot, []byte("blocking file"), 0o644); err != nil {
		t.Fatalf("write blocking file failed: %v", err)
	}

	_, err = captureStdout(t, func() error {
		return run([]string{"tabs", "capture", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs capture to fail when memory write fails")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func restoreCommandTabsClient(t *testing.T, client commandTabsClient) {
	t.Helper()

	previous := newCommandTabsClient
	newCommandTabsClient = func(paths macos.Paths) (commandTabsClient, error) {
		return client, nil
	}
	t.Cleanup(func() {
		newCommandTabsClient = previous
	})
}

func loadCommandMemoryEvents(paths macos.Paths) ([]memory.Event, error) {
	events, err := memory.Load(paths)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return events, nil
}
