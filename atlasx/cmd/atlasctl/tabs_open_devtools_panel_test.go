package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsPanelOutputsTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsPanel: tabs.Target{
			ID:    "devtools-window-1",
			Type:  "page",
			Title: "DevTools",
			URL:   "http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Ftab-1",
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel", "tab-1", "network"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools-panel failed: %v", err)
	}
	for _, fragment := range []string{"id=devtools-window-1", `title="DevTools"`, "panel=network"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOpenDevToolsPanelRejectsMissingPanel(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs open-devtools-panel to fail")
	}
	if !strings.Contains(err.Error(), "missing target id or panel for tabs open-devtools-panel") {
		t.Fatalf("unexpected error: %v", err)
	}
}
