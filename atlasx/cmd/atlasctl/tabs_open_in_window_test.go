package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenInWindowOutputsTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowOpen: tabs.WindowOpenResult{
			WindowID:          7,
			ActivatedTargetID: "tab-1",
			Target: tabs.Target{
				ID:    "tab-2",
				Type:  "page",
				Title: "OpenAI",
				URL:   "https://openai.com",
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-in-window", "7", "https://openai.com"})
	})
	if err != nil {
		t.Fatalf("run tabs open-in-window failed: %v", err)
	}
	for _, fragment := range []string{"window_id=7", "activated_target_id=tab-1", "id=tab-2", `title="OpenAI"`} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOpenInWindowRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-in-window", "bad-id", "https://openai.com"})
	})
	if err == nil {
		t.Fatal("expected tabs open-in-window to fail")
	}
	if !strings.Contains(err.Error(), `invalid window id "bad-id"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsOpenInWindowSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowOpenErr: errString("window 7 not found"),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-in-window", "7", "https://openai.com"})
	})
	if err == nil {
		t.Fatal("expected tabs open-in-window to fail")
	}
	if !strings.Contains(err.Error(), "window 7 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
