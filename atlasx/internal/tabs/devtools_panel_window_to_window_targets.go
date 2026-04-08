package tabs

import (
	"fmt"
	"strings"
)

type DevToolsPanelWindowToWindowsResult struct {
	SourceWindowID int                             `json:"source_window_id"`
	Panel          string                          `json:"panel"`
	Returned       int                             `json:"returned"`
	OpenedTargets  []DevToolsWindowToWindowsTarget `json:"opened_targets"`
}

func (c Client) OpenDevToolsPanelWindowToWindows(sourceWindowID int, panel string) (DevToolsPanelWindowToWindowsResult, error) {
	if strings.TrimSpace(panel) == "" {
		return DevToolsPanelWindowToWindowsResult{}, fmt.Errorf("panel is required")
	}

	windows, err := c.Windows()
	if err != nil {
		return DevToolsPanelWindowToWindowsResult{}, err
	}

	var sourceWindow *WindowSummary
	for index := range windows {
		if windows[index].WindowID == sourceWindowID {
			sourceWindow = &windows[index]
			break
		}
	}
	if sourceWindow == nil {
		if err := c.ensureWindowExists(sourceWindowID); err != nil {
			return DevToolsPanelWindowToWindowsResult{}, err
		}
		return DevToolsPanelWindowToWindowsResult{
			SourceWindowID: sourceWindowID,
			Panel:          panel,
			OpenedTargets:  []DevToolsWindowToWindowsTarget{},
		}, nil
	}

	result := DevToolsPanelWindowToWindowsResult{
		SourceWindowID: sourceWindowID,
		Panel:          panel,
		OpenedTargets:  make([]DevToolsWindowToWindowsTarget, 0, len(sourceWindow.Targets)),
	}
	for _, sourceTarget := range sourceWindow.Targets {
		opened, err := c.OpenDevToolsPanelWindow(sourceTarget.ID, panel)
		if err != nil {
			return DevToolsPanelWindowToWindowsResult{}, err
		}
		result.OpenedTargets = append(result.OpenedTargets, DevToolsWindowToWindowsTarget{
			SourceTargetID: sourceTarget.ID,
			Target:         opened,
		})
		result.Returned = len(result.OpenedTargets)
	}
	return result, nil
}
