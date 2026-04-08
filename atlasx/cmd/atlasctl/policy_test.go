package main

import (
	"strings"
	"testing"
)

func TestPolicyStatusOutputsGuardrails(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"policy", "status"})
	})
	if err != nil {
		t.Fatalf("run policy status failed: %v", err)
	}

	assertContainsAll(t, output,
		"default_listen_addr=127.0.0.1:17537",
		"loopback_only_default=true",
		"remote_control_flag=--allow-remote-control",
		"shared_profile_managed=false",
		"sidebar_secrets_persisted=false",
	)
}

func TestPolicyStatusRejectsUnknownSubcommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"policy", "inspect"})
	})
	if err == nil {
		t.Fatal("expected policy inspect to fail")
	}
	if !strings.Contains(err.Error(), `unknown policy subcommand "inspect"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
