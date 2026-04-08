package main

import (
	"strings"
	"testing"

	"atlasx/internal/permissions"
)

func TestPermissionsStatusOutputsBoundaryFacts(t *testing.T) {
	previous := loadPermissionsStatus
	loadPermissionsStatus = func() permissions.Status {
		return permissions.Status{
			Source:                      permissions.SourceCodebaseBoundary,
			GrantedStateObservable:      false,
			AccessibilityProbeSupported: false,
			PermissionPromptSupported:   false,
			PermissionWriteSupported:    false,
			OSPolicyFailuresSurface:     true,
			Notes:                       []string{"note"},
		}
	}
	t.Cleanup(func() {
		loadPermissionsStatus = previous
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"permissions", "status"})
	})
	if err != nil {
		t.Fatalf("run permissions status failed: %v", err)
	}

	assertContainsAll(t, output,
		"source=codebase_boundary",
		"granted_state_observable=false",
		"accessibility_probe_supported=false",
		"permission_prompt_supported=false",
		"permission_write_supported=false",
		"os_policy_failures_surface=true",
	)
}

func TestPermissionsStatusRejectsUnknownSubcommand(t *testing.T) {
	_, err := captureStdout(t, func() error {
		return run([]string{"permissions", "inspect"})
	})
	if err == nil {
		t.Fatal("expected permissions inspect to fail")
	}
	if !strings.Contains(err.Error(), `unknown permissions subcommand "inspect"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
