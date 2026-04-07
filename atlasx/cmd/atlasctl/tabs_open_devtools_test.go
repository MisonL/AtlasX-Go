package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsOpenDevToolsOutputsTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevTools: tabs.Target{
			ID:    "devtools-window-1",
			Type:  "page",
			Title: "DevTools",
			URL:   "http://127.0.0.1/devtools/inspector.html?ws=127.0.0.1/devtools/page/tab-1",
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools failed: %v", err)
	}
	for _, fragment := range []string{"id=devtools-window-1", `title="DevTools"`, "url=http://127.0.0.1/devtools/inspector.html"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOpenDevToolsSurfacesFrontendErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsErr: errString("target does not expose a devtools frontend url"),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs open-devtools to fail")
	}
	if !strings.Contains(err.Error(), "target does not expose a devtools frontend url") {
		t.Fatalf("unexpected error: %v", err)
	}
}
