package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsEmulateDeviceOutputsStructuredResult(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		deviceResult: tabs.DeviceEmulationResult{
			ID:       "tab-1",
			Title:    "Atlas",
			URL:      "https://chatgpt.com/atlas",
			Preset:   "iphone-13",
			Viewport: "390x844@3",
			Mobile:   true,
			Touch:    true,
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "emulate-device", "tab-1", "iphone-13"})
	})
	if err != nil {
		t.Fatalf("run tabs emulate-device failed: %v", err)
	}
	for _, fragment := range []string{
		"preset=iphone-13",
		"viewport=390x844@3",
		"mobile=true",
		"touch=true",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsEmulateDeviceSupportsOffPreset(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		deviceResult: tabs.DeviceEmulationResult{
			ID:       "tab-1",
			Title:    "Atlas",
			URL:      "https://chatgpt.com/atlas",
			Preset:   "off",
			Viewport: "off",
			Mobile:   false,
			Touch:    false,
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "emulate-device", "tab-1", "off"})
	})
	if err != nil {
		t.Fatalf("run tabs emulate-device failed: %v", err)
	}
	if !strings.Contains(output, "preset=off") || !strings.Contains(output, "mobile=false") || !strings.Contains(output, "touch=false") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestTabsEmulateDeviceSurfacesPresetError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		deviceErr: errString(`unknown device preset "unknown"`),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "emulate-device", "tab-1", "unknown"})
	})
	if err == nil {
		t.Fatal("expected tabs emulate-device to fail")
	}
	if !strings.Contains(err.Error(), `unknown device preset "unknown"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
