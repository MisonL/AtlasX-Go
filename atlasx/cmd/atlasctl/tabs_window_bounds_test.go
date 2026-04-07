package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsSetWindowBoundsOutputsStructuredBounds(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowBounds: tabs.WindowBounds{
			WindowID: 7,
			State:    "normal",
			Left:     10,
			Top:      20,
			Width:    1280,
			Height:   720,
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "set-window-bounds", "7", "10", "20", "1280", "720"})
	})
	if err != nil {
		t.Fatalf("run tabs set-window-bounds failed: %v", err)
	}
	for _, fragment := range []string{"window_id=7", "left=10", "top=20", "width=1280", "height=720"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsSetWindowBoundsRejectsInvalidWidth(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "set-window-bounds", "7", "10", "20", "bad", "720"})
	})
	if err == nil {
		t.Fatal("expected tabs set-window-bounds to fail")
	}
	if !strings.Contains(err.Error(), `invalid width "bad"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsSetWindowBoundsSurfacesBoundsErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowBoundsErr: errString("width must be positive"),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "set-window-bounds", "7", "10", "20", "0", "720"})
	})
	if err == nil {
		t.Fatal("expected tabs set-window-bounds to fail")
	}
	if !strings.Contains(err.Error(), "width must be positive") {
		t.Fatalf("unexpected error: %v", err)
	}
}
