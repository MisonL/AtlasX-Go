package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsCloseWindowOutputsClosedTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowClose: tabs.WindowCloseResult{
			WindowID:      7,
			Returned:      2,
			ClosedTargets: []string{"tab-1", "tab-2"},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "close-window", "7"})
	})
	if err != nil {
		t.Fatalf("run tabs close-window failed: %v", err)
	}
	for _, fragment := range []string{"window_id=7", "returned=2", "id=tab-1", "id=tab-2"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsCloseWindowRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "close-window", "bad-id"})
	})
	if err == nil {
		t.Fatal("expected tabs close-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid window id "bad-id"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsCloseWindowSurfacesMissingWindow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowCloseErr: errString("window 7 not found"),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "close-window", "7"})
	})
	if err == nil {
		t.Fatal("expected tabs close-window to fail")
	}
	if !strings.Contains(err.Error(), "window 7 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
