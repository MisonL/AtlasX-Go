package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsInWindowOutputsWindowOpenResult(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsInWindow: tabs.WindowOpenResult{
			WindowID:          7,
			ActivatedTargetID: "tab-7",
			Target: tabs.Target{
				ID:    "devtools-window-tab",
				Type:  "page",
				Title: "DevTools",
				URL:   "http://127.0.0.1/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Ftab-1",
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-in-window", "tab-1", "7"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools-in-window failed: %v", err)
	}
	for _, fragment := range []string{"window_id=7", "activated_target_id=tab-7", "id=devtools-window-tab", `title="DevTools"`} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOpenDevToolsInWindowRejectsMissingWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-in-window", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs open-devtools-in-window to fail")
	}
	if !strings.Contains(err.Error(), "missing target id or window id for tabs open-devtools-in-window") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsOpenDevToolsInWindowRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-in-window", "tab-1", "not-int"})
	})
	if err == nil {
		t.Fatal("expected tabs open-devtools-in-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid window id "not-int"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
