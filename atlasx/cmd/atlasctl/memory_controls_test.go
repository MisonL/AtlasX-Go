package main

import (
	"strings"
	"testing"
)

func TestMemoryControlsOutputsPersistFlag(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"memory", "controls"})
	})
	if err != nil {
		t.Fatalf("run memory controls failed: %v", err)
	}
	if !strings.Contains(output, "persist_enabled=true") {
		t.Fatalf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "page_visibility_enabled=true") {
		t.Fatalf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "hidden_host_count=0") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestMemorySetPersistUpdatesControls(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"memory", "set-persist", "disabled"})
	})
	if err != nil {
		t.Fatalf("run memory set-persist failed: %v", err)
	}
	if !strings.Contains(output, "persist_enabled=false") {
		t.Fatalf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "page_visibility_enabled=true") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestMemorySetPageVisibilityUpdatesControls(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"memory", "set-page-visibility", "hidden"})
	})
	if err != nil {
		t.Fatalf("run memory set-page-visibility failed: %v", err)
	}
	if !strings.Contains(output, "page_visibility_enabled=false") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestMemorySetSiteVisibilityUpdatesControls(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"memory", "set-site-visibility", "https://ChatGPT.com/atlas", "hidden"})
	})
	if err != nil {
		t.Fatalf("run memory set-site-visibility failed: %v", err)
	}
	if !strings.Contains(output, "hidden_host_count=1") {
		t.Fatalf("unexpected output: %s", output)
	}
	if !strings.Contains(output, "hidden_host[0]=chatgpt.com") {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestMemorySetSiteVisibilityVisibleClearsControls(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"memory", "set-site-visibility", "chatgpt.com", "hidden"})
	})
	if err != nil {
		t.Fatalf("run memory set-site-visibility hidden failed: %v", err)
	}

	output, err := captureStdout(t, func() error {
		return run([]string{"memory", "set-site-visibility", "chatgpt.com", "visible"})
	})
	if err != nil {
		t.Fatalf("run memory set-site-visibility visible failed: %v", err)
	}
	if !strings.Contains(output, "hidden_host_count=0") {
		t.Fatalf("unexpected output: %s", output)
	}
	if strings.Contains(output, "hidden_host[0]=") {
		t.Fatalf("expected hidden hosts cleared, got output: %s", output)
	}
}

func TestMemorySetPersistRejectsInvalidValue(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"memory", "set-persist", "maybe"})
	})
	if err == nil {
		t.Fatal("expected memory set-persist to fail")
	}
	if !strings.Contains(err.Error(), `invalid persist value "maybe"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemorySetPageVisibilityRejectsInvalidValue(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"memory", "set-page-visibility", "maybe"})
	})
	if err == nil {
		t.Fatal("expected memory set-page-visibility to fail")
	}
	if !strings.Contains(err.Error(), `invalid page visibility value "maybe"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMemorySetSiteVisibilityRejectsInvalidHost(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"memory", "set-site-visibility", ":// bad host", "hidden"})
	})
	if err == nil {
		t.Fatal("expected memory set-site-visibility to fail")
	}
	if !strings.Contains(err.Error(), `invalid site host ":// bad host"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
