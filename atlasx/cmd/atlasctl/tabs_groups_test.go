package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsGroupsOutputsStructuredInferredGroups(t *testing.T) {
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
				WindowID: 22,
				Targets: []tabs.Target{
					{ID: "tab-3", Type: "page", Title: "Atlas C", URL: "https://chatgpt.com/atlas/c"},
					{ID: "tab-4", Type: "page", Title: "Elsewhere", URL: "https://example.com/other"},
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "groups"})
	})
	if err != nil {
		t.Fatalf("run tabs groups failed: %v", err)
	}
	for _, fragment := range []string{
		"inferred=true",
		"returned=1",
		"group_id=host:chatgpt.com",
		`label="chatgpt.com"`,
		"window_returned=2",
		"window_id=11",
		"window_id=22",
		`title="Atlas A"`,
		`title="Atlas C"`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}
