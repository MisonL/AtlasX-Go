package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsSetTitleOutputsStructuredResult(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		titleUpdate: tabs.TitleUpdateResult{
			ID:    "tab-1",
			Title: "Atlas Workbench",
			URL:   "https://openai.com/work",
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "set-title", "tab-1", "Atlas Workbench"})
	})
	if err != nil {
		t.Fatalf("run tabs set-title failed: %v", err)
	}
	for _, fragment := range []string{
		"id=tab-1",
		`title="Atlas Workbench"`,
		"url=https://openai.com/work",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsSetTitleRejectsMissingArgs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "set-title", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs set-title to fail")
	}
	if !strings.Contains(err.Error(), "missing target id or title for tabs set-title") {
		t.Fatalf("unexpected error: %v", err)
	}
}
