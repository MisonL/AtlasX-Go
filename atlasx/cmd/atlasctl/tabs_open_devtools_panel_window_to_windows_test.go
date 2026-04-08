package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsPanelWindowToWindowsOutputsStructuredResult(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsPanelWindowToWindows: tabs.DevToolsPanelWindowToWindowsResult{
			SourceWindowID: 11,
			Panel:          "network",
			Returned:       2,
			OpenedTargets: []tabs.DevToolsWindowToWindowsTarget{
				{
					SourceTargetID: "src-1",
					Target: tabs.Target{
						ID:    "devtools-1",
						Type:  "page",
						Title: "DevTools A",
						URL:   "http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1",
					},
				},
				{
					SourceTargetID: "src-2",
					Target: tabs.Target{
						ID:    "devtools-2",
						Type:  "page",
						Title: "DevTools B",
						URL:   "http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-2",
					},
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel-window-to-windows", "11", "network"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools-panel-window-to-windows failed: %v", err)
	}
	for _, fragment := range []string{
		"source_window_id=11",
		"panel=network",
		"returned=2",
		"source_target_id=src-1",
		"source_target_id=src-2",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOpenDevToolsPanelWindowToWindowsRejectsMissingArgs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel-window-to-windows", "11"})
	})
	if err == nil {
		t.Fatal("expected tabs open-devtools-panel-window-to-windows to fail")
	}
	if !strings.Contains(err.Error(), "missing source window id or panel for tabs open-devtools-panel-window-to-windows") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsOpenDevToolsPanelWindowToWindowsRejectsInvalidSourceWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel-window-to-windows", "bad", "network"})
	})
	if err == nil {
		t.Fatal("expected tabs open-devtools-panel-window-to-windows to fail")
	}
	if !strings.Contains(err.Error(), `invalid source window id "bad"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
