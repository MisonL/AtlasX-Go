package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsPanelWindowIntoWindowOutputsStructuredResult(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsPanelWindowIntoWindow: tabs.DevToolsPanelWindowOpenResult{
			SourceWindowID: 11,
			Panel:          "network",
			TargetWindowID: 21,
			Returned:       2,
			OpenedTargets: []tabs.DevToolsWindowOpenTarget{
				{
					SourceTargetID:    "src-1",
					ActivatedTargetID: "dst-1",
					Target: tabs.Target{
						ID:    "devtools-1",
						Type:  "page",
						Title: "DevTools A",
						URL:   "http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1",
					},
				},
				{
					SourceTargetID:    "src-2",
					ActivatedTargetID: "dst-1",
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
		return run([]string{"tabs", "open-devtools-panel-window-into-window", "11", "network", "21"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools-panel-window-into-window failed: %v", err)
	}
	for _, fragment := range []string{
		"source_window_id=11",
		"panel=network",
		"target_window_id=21",
		"returned=2",
		"source_target_id=src-1",
		"source_target_id=src-2",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOpenDevToolsPanelWindowIntoWindowRejectsMissingArgs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel-window-into-window", "11", "network"})
	})
	if err == nil {
		t.Fatal("expected tabs open-devtools-panel-window-into-window to fail")
	}
	if !strings.Contains(err.Error(), "missing source window id or panel or target window id for tabs open-devtools-panel-window-into-window") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsOpenDevToolsPanelWindowIntoWindowRejectsInvalidTargetWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel-window-into-window", "11", "network", "bad"})
	})
	if err == nil {
		t.Fatal("expected tabs open-devtools-panel-window-into-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid target window id "bad"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
