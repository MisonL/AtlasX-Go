package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsMoveToNewWindowOutputsTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowMoveNew: tabs.WindowMoveToNewResult{
			SourceWindowID: 9,
			SourceTargetID: "src-1",
			Target: tabs.Target{
				ID:    "new-1",
				Type:  "page",
				Title: "OpenAI",
				URL:   "https://openai.com",
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "move-to-new-window", "src-1"})
	})
	if err != nil {
		t.Fatalf("run tabs move-to-new-window failed: %v", err)
	}
	for _, fragment := range []string{"source_window_id=9", "source_target_id=src-1", "id=new-1"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsMoveToNewWindowRejectsMissingTargetID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "move-to-new-window"})
	})
	if err == nil {
		t.Fatal("expected tabs move-to-new-window to fail")
	}
	if !strings.Contains(err.Error(), "missing target id for tabs move-to-new-window") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsMoveToNewWindowSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowMoveNewErr: errString("page target src-1 not found"),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "move-to-new-window", "src-1"})
	})
	if err == nil {
		t.Fatal("expected tabs move-to-new-window to fail")
	}
	if !strings.Contains(err.Error(), "page target src-1 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
