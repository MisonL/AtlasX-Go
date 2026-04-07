package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsActivateWindowOutputsActivatedTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowActivate: tabs.WindowActivateResult{
			WindowID:          7,
			ActivatedTargetID: "tab-1",
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "activate-window", "7"})
	})
	if err != nil {
		t.Fatalf("run tabs activate-window failed: %v", err)
	}
	for _, fragment := range []string{"window_id=7", "activated_target_id=tab-1"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsActivateWindowRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "activate-window", "bad-id"})
	})
	if err == nil {
		t.Fatal("expected tabs activate-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid window id "bad-id"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsActivateWindowSurfacesMissingWindow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowActivateErr: errString("window 7 not found"),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "activate-window", "7"})
	})
	if err == nil {
		t.Fatal("expected tabs activate-window to fail")
	}
	if !strings.Contains(err.Error(), "window 7 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
