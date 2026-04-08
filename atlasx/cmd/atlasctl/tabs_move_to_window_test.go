package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsMoveToWindowOutputsTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowMove: tabs.WindowMoveResult{
			SourceWindowID:    9,
			TargetWindowID:    7,
			SourceTargetID:    "src-1",
			ActivatedTargetID: "dst-1",
			Target: tabs.Target{
				ID:    "new-1",
				Type:  "page",
				Title: "OpenAI",
				URL:   "https://openai.com",
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "move-to-window", "src-1", "7"})
	})
	if err != nil {
		t.Fatalf("run tabs move-to-window failed: %v", err)
	}
	for _, fragment := range []string{"source_window_id=9", "target_window_id=7", "source_target_id=src-1", "id=new-1"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsMoveToWindowRejectsInvalidTargetWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "move-to-window", "src-1", "bad"})
	})
	if err == nil {
		t.Fatal("expected tabs move-to-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid target window id "bad"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsMoveToWindowSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowMoveErr: errString("window 7 not found"),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "move-to-window", "src-1", "7"})
	})
	if err == nil {
		t.Fatal("expected tabs move-to-window to fail")
	}
	if !strings.Contains(err.Error(), "window 7 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
