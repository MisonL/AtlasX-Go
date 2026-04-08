package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOrganizeGroupIntoWindowOutputsMovedAndAlignedTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
			{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
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
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-group-into-window", "host:chatgpt.com", "11"})
	})
	if err != nil {
		t.Fatalf("run tabs organize-group-into-window failed: %v", err)
	}
	for _, fragment := range []string{"group_id=host:chatgpt.com", "window_id=11", "moved_index=0", "aligned_index=0"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOrganizeGroupIntoWindowRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-group-into-window", "host:chatgpt.com", "bad"})
	})
	if err == nil {
		t.Fatal("expected tabs organize-group-into-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid window id "bad"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
