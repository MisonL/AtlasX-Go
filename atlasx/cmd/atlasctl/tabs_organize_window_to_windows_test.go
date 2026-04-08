package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOrganizeWindowToWindowsOutputsGroups(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
					{ID: "tab-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
					{ID: "tab-4", Type: "page", Title: "Build Log - B", URL: "about:blank"},
				},
			},
			{
				WindowID: 21,
				Targets: []tabs.Target{
					{ID: "new-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
			{
				WindowID: 22,
				Targets: []tabs.Target{
					{ID: "new-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
				},
			},
		},
		windowMoveNewByID: map[string]tabs.WindowMoveToNewResult{
			"tab-1": {
				SourceWindowID: 11,
				SourceTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-1",
					Type:  "page",
					Title: "Atlas A",
					URL:   "https://chatgpt.com/atlas/a",
				},
			},
			"tab-3": {
				SourceWindowID: 11,
				SourceTargetID: "tab-3",
				Target: tabs.Target{
					ID:    "new-3",
					Type:  "page",
					Title: "Build Log - A",
					URL:   "about:blank",
				},
			},
		},
		windowMoveByID: map[string]tabs.WindowMoveResult{
			"tab-2": {
				SourceWindowID:    11,
				TargetWindowID:    21,
				SourceTargetID:    "tab-2",
				ActivatedTargetID: "new-1",
				Target: tabs.Target{
					ID:    "new-2",
					Type:  "page",
					Title: "Atlas B",
					URL:   "https://chatgpt.com/atlas/b",
				},
			},
			"tab-4": {
				SourceWindowID:    11,
				TargetWindowID:    22,
				SourceTargetID:    "tab-4",
				ActivatedTargetID: "new-3",
				Target: tabs.Target{
					ID:    "new-4",
					Type:  "page",
					Title: "Build Log - B",
					URL:   "about:blank",
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-window-to-windows", "11"})
	})
	if err != nil {
		t.Fatalf("run tabs organize-window-to-windows failed: %v", err)
	}
	for _, fragment := range []string{"source_window_id=11", "returned=2", "group_id=host:chatgpt.com", "group_id=title:build log"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOrganizeWindowToWindowsRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-window-to-windows", "bad"})
	})
	if err == nil {
		t.Fatal("expected tabs organize-window-to-windows to fail")
	}
	if !strings.Contains(err.Error(), `invalid source window id "bad"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
