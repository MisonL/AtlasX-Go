package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsDevToolsPanelOutputsFrontendURL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		devToolsPanel: tabs.DevToolsTarget{
			ID:                  "tab-1",
			Title:               "Atlas",
			URL:                 "https://chatgpt.com/atlas",
			DevToolsFrontendURL: "http://127.0.0.1:9222/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2F1",
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "devtools-panel", "tab-1", "network"})
	})
	if err != nil {
		t.Fatalf("run tabs devtools-panel failed: %v", err)
	}
	for _, fragment := range []string{"id=tab-1", "devtools_frontend_url=http://127.0.0.1:9222/devtools/inspector.html?panel=network"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsDevToolsPanelRejectsMissingPanel(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "devtools-panel", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs devtools-panel to fail")
	}
	if !strings.Contains(err.Error(), "missing target id or panel for tabs devtools-panel") {
		t.Fatalf("unexpected error: %v", err)
	}
}
