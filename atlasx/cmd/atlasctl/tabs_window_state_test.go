package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsSetWindowStateOutputsStructuredBounds(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowState: tabs.WindowBounds{
			WindowID: 7,
			State:    "maximized",
			Left:     20,
			Top:      30,
			Width:    1440,
			Height:   900,
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "set-window-state", "7", "maximized"})
	})
	if err != nil {
		t.Fatalf("run tabs set-window-state failed: %v", err)
	}
	for _, fragment := range []string{"window_id=7", "state=maximized", "width=1440", "height=900"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsSetWindowStateRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "set-window-state", "bad-id", "normal"})
	})
	if err == nil {
		t.Fatal("expected tabs set-window-state to fail")
	}
	if !strings.Contains(err.Error(), `invalid window id "bad-id"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsSetWindowStateSurfacesStateErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowStateErr: errString(`unknown window state "unknown"`),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "set-window-state", "7", "unknown"})
	})
	if err == nil {
		t.Fatal("expected tabs set-window-state to fail")
	}
	if !strings.Contains(err.Error(), `unknown window state "unknown"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
