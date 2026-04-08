package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOrganizeToWindowsOutputsGroups(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
			{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
			{ID: "tab-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
			{ID: "tab-4", Type: "page", Title: "Build Log - B", URL: "about:blank"},
		},
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "new-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
			{
				WindowID: 12,
				Targets: []tabs.Target{
					{ID: "new-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
				},
			},
		},
		windowMoveNewByID: map[string]tabs.WindowMoveToNewResult{
			"tab-1": {
				SourceWindowID: 9,
				SourceTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-1",
					Type:  "page",
					Title: "Atlas A",
					URL:   "https://chatgpt.com/atlas/a",
				},
			},
			"tab-3": {
				SourceWindowID: 9,
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
				SourceWindowID:    9,
				TargetWindowID:    11,
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
				SourceWindowID:    9,
				TargetWindowID:    12,
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
		return run([]string{"tabs", "organize-to-windows"})
	})
	if err != nil {
		t.Fatalf("run tabs organize-to-windows failed: %v", err)
	}
	for _, fragment := range []string{"returned=2", "group_id=host:chatgpt.com", "group_id=title:build log", "window_id=12"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOrganizeToWindowsReturnsEmptyWithoutGroups(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Solo", URL: "https://example.com"},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-to-windows"})
	})
	if err != nil {
		t.Fatalf("run tabs organize-to-windows failed: %v", err)
	}
	if !strings.Contains(output, "returned=0") {
		t.Fatalf("unexpected output: %s", output)
	}
}
