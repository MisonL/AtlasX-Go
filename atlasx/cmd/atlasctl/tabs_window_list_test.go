package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsWindowsOutputsStructuredWindows(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 7,
				State:    "normal",
				Left:     20,
				Top:      30,
				Width:    1440,
				Height:   900,
				Returned: 1,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "windows"})
	})
	if err != nil {
		t.Fatalf("run tabs windows failed: %v", err)
	}
	for _, fragment := range []string{
		"returned=1",
		"window_id=7",
		"state=normal",
		`title="Atlas"`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsWindowsSurfacesBrowserErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windowsErr: errString("browser websocket debugger url is not available"),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "windows"})
	})
	if err == nil {
		t.Fatal("expected tabs windows to fail")
	}
	if !strings.Contains(err.Error(), "browser websocket debugger url is not available") {
		t.Fatalf("unexpected error: %v", err)
	}
}
