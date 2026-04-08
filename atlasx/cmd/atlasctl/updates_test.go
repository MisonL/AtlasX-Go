package main

import (
	"strings"
	"testing"
)

func TestUpdatesStatusCommandRendersSummary(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"updates", "status"})
	})
	if err != nil {
		t.Fatalf("run updates status failed: %v", err)
	}

	assertContainsAll(t, output,
		"runtime_root=",
		"manifest_present=false",
		"staged_ready=false",
		"plan_present=false",
		"plan_pending=false",
	)
}

func TestUpdatesStatusCommandRejectsUnknownSubcommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"updates", "inspect"})
	})
	if err == nil {
		t.Fatal("expected updates inspect to fail")
	}
	if !strings.Contains(err.Error(), `unknown updates subcommand "inspect"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
