package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOrganizeWindowOutputsStructuredGroups(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
					{ID: "tab-3", Type: "page", Title: "Elsewhere", URL: "https://example.com/other"},
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-window", "11"})
	})
	if err != nil {
		t.Fatalf("run tabs organize-window failed: %v", err)
	}
	for _, fragment := range []string{
		"source_window_id=11",
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

func TestTabsOrganizeWindowRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "organize-window", "bad"})
	})
	if err == nil {
		t.Fatal("expected tabs organize-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid window id "bad"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
