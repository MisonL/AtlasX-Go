package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOrganizeGroupToWindowOutputsMovedTargets(t *testing.T) {
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
					{ID: "new-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
				},
			},
		},
		windowMoveNew: tabs.WindowMoveToNewResult{
			SourceWindowID: 9,
			SourceTargetID: "tab-1",
			Target: tabs.Target{
				ID:    "new-1",
				Type:  "page",
				Title: "Atlas A",
				URL:   "https://chatgpt.com/atlas/a",
			},
		},
		windowMove: tabs.WindowMoveResult{
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
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-group-to-window", "host:chatgpt.com"})
	})
	if err != nil {
		t.Fatalf("run tabs organize-group-to-window failed: %v", err)
	}
	for _, fragment := range []string{"group_id=host:chatgpt.com", "window_id=11", "moved_index=1", "source_target_id=tab-2", "id=new-2"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOrganizeGroupToWindowRejectsMissingGroupID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-group-to-window"})
	})
	if err == nil {
		t.Fatal("expected tabs organize-group-to-window to fail")
	}
	if !strings.Contains(err.Error(), "missing group id for tabs organize-group-to-window") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsOrganizeGroupToWindowSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Solo", URL: "https://example.com"},
		},
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-group-to-window", "host:chatgpt.com"})
	})
	if err == nil {
		t.Fatal("expected tabs organize-group-to-window to fail")
	}
	if !strings.Contains(err.Error(), "group host:chatgpt.com not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
