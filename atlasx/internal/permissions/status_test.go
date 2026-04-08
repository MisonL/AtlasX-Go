package permissions

import (
	"strings"
	"testing"
)

func TestLoadStatusReturnsBoundaryFacts(t *testing.T) {
	status := LoadStatus()

	if status.Source != SourceCodebaseBoundary {
		t.Fatalf("unexpected source: %+v", status)
	}
	if status.NativeBridgePresent {
		t.Fatalf("expected no native bridge: %+v", status)
	}
	if status.GrantedStateObservable {
		t.Fatalf("expected granted state to be unobservable: %+v", status)
	}
	if status.AccessibilityProbeSupported || status.ScreenRecordingProbeSupported ||
		status.AutomationProbeSupported || status.FullDiskAccessProbeSupported {
		t.Fatalf("expected permission probes to be unsupported: %+v", status)
	}
	if status.PermissionPromptSupported || status.PermissionWriteSupported || status.PermissionStatePersisted {
		t.Fatalf("expected permission prompt/write path to be unsupported: %+v", status)
	}
	if !status.OSPolicyFailuresSurface {
		t.Fatalf("expected os policy failures to surface: %+v", status)
	}
	if len(status.Notes) != 3 {
		t.Fatalf("unexpected notes: %+v", status)
	}
}

func TestStatusRenderIncludesBoundaryFacts(t *testing.T) {
	rendered := LoadStatus().Render()

	for _, fragment := range []string{
		"source=codebase_boundary",
		"granted_state_observable=false",
		"accessibility_probe_supported=false",
		"permission_prompt_supported=false",
		"permission_write_supported=false",
		"os_policy_failures_surface=true",
	} {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("expected render to contain %q, got %s", fragment, rendered)
		}
	}
}
