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
	targets         []tabs.Target
	listErr         error
	context         tabs.PageContext
	captureErr      error
	semanticContext tabs.SemanticContext
	semanticErr     error
	selection       tabs.SelectionContext
	selectionErr    error
	devTools        tabs.DevToolsTarget
	devToolsErr     error
}

func (s *stubCommandTabsClient) List() ([]tabs.Target, error) {
	return s.targets, s.listErr
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

func (s *stubCommandTabsClient) CaptureSemanticContext(string) (tabs.SemanticContext, error) {
	return s.semanticContext, s.semanticErr
}

func (s *stubCommandTabsClient) CaptureSelection(string) (tabs.SelectionContext, error) {
	return s.selection, s.selectionErr
}

func (s *stubCommandTabsClient) DevTools(string) (tabs.DevToolsTarget, error) {
	return s.devTools, s.devToolsErr
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

func TestTabsSelectionOutputsSelectionContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		selection: tabs.SelectionContext{
			ID:                     "tab-1",
			Title:                  "Atlas",
			URL:                    "https://chatgpt.com/atlas",
			SelectionText:          "Atlas selected text",
			CapturedAt:             "2026-04-07T09:00:00Z",
			SelectionPresent:       true,
			SelectionTextLength:    20,
			SelectionTextLimit:     1024,
			SelectionTextTruncated: false,
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "selection", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs selection failed: %v", err)
	}
	if !strings.Contains(output, `selection_text="Atlas selected text"`) {
		t.Fatalf("unexpected output: %s", output)
	}
	if !strings.Contains(output, `selection_present=true`) {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestTabsSelectionFailurePrintsCaptureError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	selection := tabs.SelectionContext{
		ID:                 "tab-1",
		Title:              "Atlas",
		URL:                "https://chatgpt.com/atlas",
		CapturedAt:         "2026-04-07T09:00:00Z",
		SelectionTextLimit: 1024,
		CaptureError:       "cdp error -32000: selection failed",
	}
	restoreCommandTabsClient(t, &stubCommandTabsClient{
		selection: selection,
		selectionErr: &tabs.SelectionCaptureError{
			Context: selection,
			Cause:   errors.New("cdp error -32000: selection failed"),
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "selection", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs selection to fail")
	}
	if !strings.Contains(output, `capture_error="cdp error -32000: selection failed"`) {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestTabsDevToolsOutputsFrontendURL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		devTools: tabs.DevToolsTarget{
			ID:                  "tab-1",
			Title:               "Atlas",
			URL:                 "https://chatgpt.com/atlas",
			DevToolsFrontendURL: "http://127.0.0.1:9222/devtools/inspector.html?ws=127.0.0.1:9222/devtools/page/1",
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "devtools", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs devtools failed: %v", err)
	}
	if !strings.Contains(output, "devtools_frontend_url=http://127.0.0.1:9222/devtools/inspector.html?ws=127.0.0.1:9222/devtools/page/1") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestTabsOrganizeOutputsStructuredGroups(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
			{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
			{ID: "tab-3", Type: "page", Title: "Elsewhere", URL: "https://example.com/other"},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize"})
	})
	if err != nil {
		t.Fatalf("run tabs organize failed: %v", err)
	}
	for _, fragment := range []string{
		"returned=1",
		"group_id=host:chatgpt.com",
		`label="chatgpt.com"`,
		`title="Atlas A"`,
		`title="Atlas B"`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
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
