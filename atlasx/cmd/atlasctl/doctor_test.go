package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDoctorCommandOutputsTextByDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"doctor"})
	})
	if err != nil {
		t.Fatalf("run doctor failed: %v", err)
	}

	assertContainsAll(t, output,
		"AtlasX Doctor",
		"support_root=",
		"chrome_status=",
	)
}

func TestDoctorCommandOutputsJSON(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"doctor", "--json"})
	})
	if err != nil {
		t.Fatalf("run doctor --json failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("decode doctor json failed: %v output=%s", err, output)
	}
	for _, key := range []string{"Paths", "Config", "Chrome", "ChromeStatus", "RuntimeManifest", "IsolatedPath", "SharedModeName", "Session"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in payload: %+v", key, payload)
		}
	}
}

func TestDoctorCommandRejectsPositionalArguments(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"doctor", "extra"})
	})
	if err == nil {
		t.Fatal("expected doctor extra to fail")
	}
	if !strings.Contains(err.Error(), "doctor accepts no positional arguments") {
		t.Fatalf("unexpected error: %v", err)
	}
}
