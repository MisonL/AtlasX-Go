package permissions

import (
	"fmt"
	"strings"
)

const SourceCodebaseBoundary = "codebase_boundary"

type Status struct {
	Source                        string   `json:"source"`
	NativeBridgePresent           bool     `json:"native_bridge_present"`
	GrantedStateObservable        bool     `json:"granted_state_observable"`
	AccessibilityProbeSupported   bool     `json:"accessibility_probe_supported"`
	ScreenRecordingProbeSupported bool     `json:"screen_recording_probe_supported"`
	AutomationProbeSupported      bool     `json:"automation_probe_supported"`
	FullDiskAccessProbeSupported  bool     `json:"full_disk_access_probe_supported"`
	PermissionPromptSupported     bool     `json:"permission_prompt_supported"`
	PermissionWriteSupported      bool     `json:"permission_write_supported"`
	PermissionStatePersisted      bool     `json:"permission_state_persisted"`
	OSPolicyFailuresSurface       bool     `json:"os_policy_failures_surface"`
	Notes                         []string `json:"notes"`
}

func LoadStatus() Status {
	return Status{
		Source:                        SourceCodebaseBoundary,
		NativeBridgePresent:           false,
		GrantedStateObservable:        false,
		AccessibilityProbeSupported:   false,
		ScreenRecordingProbeSupported: false,
		AutomationProbeSupported:      false,
		FullDiskAccessProbeSupported:  false,
		PermissionPromptSupported:     false,
		PermissionWriteSupported:      false,
		PermissionStatePersisted:      false,
		OSPolicyFailuresSurface:       true,
		Notes: []string{
			"atlasx does not currently probe real macOS TCC grant state",
			"atlasx does not currently request or write permissions",
			"os policy denials are expected to surface as explicit runtime errors",
		},
	}
}

func (s Status) Render() string {
	lines := []string{
		fmt.Sprintf("source=%s", s.Source),
		fmt.Sprintf("native_bridge_present=%t", s.NativeBridgePresent),
		fmt.Sprintf("granted_state_observable=%t", s.GrantedStateObservable),
		fmt.Sprintf("accessibility_probe_supported=%t", s.AccessibilityProbeSupported),
		fmt.Sprintf("screen_recording_probe_supported=%t", s.ScreenRecordingProbeSupported),
		fmt.Sprintf("automation_probe_supported=%t", s.AutomationProbeSupported),
		fmt.Sprintf("full_disk_access_probe_supported=%t", s.FullDiskAccessProbeSupported),
		fmt.Sprintf("permission_prompt_supported=%t", s.PermissionPromptSupported),
		fmt.Sprintf("permission_write_supported=%t", s.PermissionWriteSupported),
		fmt.Sprintf("permission_state_persisted=%t", s.PermissionStatePersisted),
		fmt.Sprintf("os_policy_failures_surface=%t", s.OSPolicyFailuresSurface),
		fmt.Sprintf("note_count=%d", len(s.Notes)),
	}
	for index, note := range s.Notes {
		lines = append(lines, fmt.Sprintf("note[%d]=%s", index, note))
	}
	return strings.Join(lines, "\n") + "\n"
}
