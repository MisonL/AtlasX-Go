package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsPanelInWindowOutputsWindowOpenResult(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsPanelInWindow: tabs.WindowOpenResult{
			WindowID:          7,
			ActivatedTargetID: "tab-7",
			Target: tabs.Target{
				ID:    "devtools-window-tab",
				Type:  "page",
				Title: "DevTools",
				URL:   "http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Ftab-1",
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel-in-window", "tab-1", "network", "7"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools-panel-in-window failed: %v", err)
	}
	for _, fragment := range []string{"window_id=7", "activated_target_id=tab-7", "id=devtools-window-tab", "panel=network"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOpenDevToolsPanelInWindowRejectsMissingWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel-in-window", "tab-1", "network"})
	})
	if err == nil {
		t.Fatal("expected tabs open-devtools-panel-in-window to fail")
	}
	if !strings.Contains(err.Error(), "missing target id or panel or window id for tabs open-devtools-panel-in-window") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsOpenDevToolsPanelInWindowRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel-in-window", "tab-1", "network", "not-int"})
	})
	if err == nil {
		t.Fatal("expected tabs open-devtools-panel-in-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid window id "not-int"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
