package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOrganizeIntoWindowOutputsGroups(t *testing.T) {
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
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
			{
				WindowID: 9,
				Targets: []tabs.Target{
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
				},
			},
			{
				WindowID: 7,
				Targets: []tabs.Target{
					{ID: "tab-3", Type: "page", Title: "Build Log - A", URL: "about:blank"},
					{ID: "tab-4", Type: "page", Title: "Build Log - B", URL: "about:blank"},
				},
			},
		},
		windowMoveByID: map[string]tabs.WindowMoveResult{
			"tab-2": {
				SourceWindowID:    9,
				TargetWindowID:    11,
				SourceTargetID:    "tab-2",
				ActivatedTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-2",
					Type:  "page",
					Title: "Atlas B",
					URL:   "https://chatgpt.com/atlas/b",
				},
			},
			"tab-3": {
				SourceWindowID:    7,
				TargetWindowID:    11,
				SourceTargetID:    "tab-3",
				ActivatedTargetID: "tab-1",
				Target: tabs.Target{
					ID:    "new-3",
					Type:  "page",
					Title: "Build Log - A",
					URL:   "about:blank",
				},
			},
			"tab-4": {
				SourceWindowID:    7,
				TargetWindowID:    11,
				SourceTargetID:    "tab-4",
				ActivatedTargetID: "tab-1",
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
		return run([]string{"tabs", "organize-into-window", "11"})
	})
	if err != nil {
		t.Fatalf("run tabs organize-into-window failed: %v", err)
	}
	for _, fragment := range []string{"returned=2", "group_id=host:chatgpt.com", "group_id=title:build log", "aligned_index=0", "window_id=11"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOrganizeIntoWindowRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-into-window", "bad"})
	})
	if err == nil {
		t.Fatal("expected tabs organize-into-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid window id "bad"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
