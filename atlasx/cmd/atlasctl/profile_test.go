package main

import (
	"strings"
	"testing"
)

func TestProfileStatusOutputsDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"profile", "status"})
	})
	if err != nil {
		t.Fatalf("run profile status failed: %v", err)
	}

	assertContainsAll(t, output,
		"profiles_root=",
		"default_profile=isolated",
		"selected_mode=isolated",
		"isolated_present=true",
		"shared_managed=false",
	)
}

func TestProfileStatusRejectsUnknownSubcommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"profile", "inspect"})
	})
	if err == nil {
		t.Fatal("expected profile inspect to fail")
	}
	if !strings.Contains(err.Error(), `unknown profile subcommand "inspect"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
