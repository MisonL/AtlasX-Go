package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsMergeWindowOutputsMovedTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowMerge: tabs.WindowMergeResult{
			SourceWindowID: 9,
			TargetWindowID: 7,
			Returned:       1,
			MovedTargets: []tabs.WindowMergeTarget{
				{
					SourceTargetID:    "src-1",
					ActivatedTargetID: "dst-1",
					Target: tabs.Target{
						ID:    "new-1",
						Type:  "page",
						Title: "OpenAI",
						URL:   "https://openai.com",
					},
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "merge-window", "9", "7"})
	})
	if err != nil {
		t.Fatalf("run tabs merge-window failed: %v", err)
	}
	for _, fragment := range []string{"source_window_id=9", "target_window_id=7", "source_target_id=src-1", "id=new-1"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsMergeWindowRejectsInvalidSourceWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "merge-window", "bad", "7"})
	})
	if err == nil {
		t.Fatal("expected tabs merge-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid source window id "bad"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsMergeWindowSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowMergeErr: errString("window 9 not found"),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "merge-window", "9", "7"})
	})
	if err == nil {
		t.Fatal("expected tabs merge-window to fail")
	}
	if !strings.Contains(err.Error(), "window 9 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
