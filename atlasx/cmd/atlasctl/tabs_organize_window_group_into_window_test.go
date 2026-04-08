package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOrganizeWindowGroupIntoWindowOutputsResult(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
				},
			},
			{
				WindowID: 21,
				Targets: []tabs.Target{
					{ID: "dst-1", Type: "page", Title: "Workspace", URL: "https://workspace.example.com"},
				},
			},
		},
		windowMoveByID: map[string]tabs.WindowMoveResult{
			"tab-1": {
				SourceWindowID:    11,
				TargetWindowID:    21,
				SourceTargetID:    "tab-1",
				ActivatedTargetID: "dst-1",
				Target: tabs.Target{
					ID:    "new-1",
					Type:  "page",
					Title: "Atlas A",
					URL:   "https://chatgpt.com/atlas/a",
				},
			},
			"tab-2": {
				SourceWindowID:    11,
				TargetWindowID:    21,
				SourceTargetID:    "tab-2",
				ActivatedTargetID: "dst-1",
				Target: tabs.Target{
					ID:    "new-2",
					Type:  "page",
					Title: "Atlas B",
					URL:   "https://chatgpt.com/atlas/b",
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-window-group-into-window", "11", "host:chatgpt.com", "21"})
	})
	if err != nil {
		t.Fatalf("run tabs organize-window-group-into-window failed: %v", err)
	}
	for _, fragment := range []string{"source_window_id=11", "target_window_id=21", "group_id=host:chatgpt.com", "window_id=21", "returned=2"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOrganizeWindowGroupIntoWindowRejectsInvalidTargetWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-window-group-into-window", "11", "host:chatgpt.com", "bad"})
	})
	if err == nil {
		t.Fatal("expected tabs organize-window-group-into-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid target window id "bad"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
