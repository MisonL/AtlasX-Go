package tabs

import (
	"fmt"
	"strings"
)

type DevToolsWindowToWindowsTarget struct {
	SourceTargetID string `json:"source_target_id"`
	Target         Target `json:"target"`
}

type DevToolsWindowToWindowsResult struct {
	SourceWindowID int                             `json:"source_window_id"`
	Returned       int                             `json:"returned"`
	OpenedTargets  []DevToolsWindowToWindowsTarget `json:"opened_targets"`
}

func (c Client) OpenDevToolsWindowToWindows(sourceWindowID int) (DevToolsWindowToWindowsResult, error) {
	windows, err := c.Windows()
	if err != nil {
		return DevToolsWindowToWindowsResult{}, err
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
			return DevToolsWindowToWindowsResult{}, err
		}
		return DevToolsWindowToWindowsResult{
			SourceWindowID: sourceWindowID,
			OpenedTargets:  []DevToolsWindowToWindowsTarget{},
		}, nil
	}

	result := DevToolsWindowToWindowsResult{
		SourceWindowID: sourceWindowID,
		OpenedTargets:  make([]DevToolsWindowToWindowsTarget, 0, len(sourceWindow.Targets)),
	}
	for _, sourceTarget := range sourceWindow.Targets {
		opened, err := c.OpenDevToolsWindow(sourceTarget.ID)
		if err != nil {
			return DevToolsWindowToWindowsResult{}, err
		}
		result.OpenedTargets = append(result.OpenedTargets, DevToolsWindowToWindowsTarget{
			SourceTargetID: sourceTarget.ID,
			Target:         opened,
		})
	}
	result.Returned = len(result.OpenedTargets)
	return result, nil
}

func (c Client) ensureWindowExists(windowID int) error {
	browserWS, err := c.browserWS()
	if err != nil {
		return err
	}
	if _, err := c.getWindowBounds(browserWS, windowID, 40); err != nil {
		if isWindowNotFoundError(err) {
			return fmt.Errorf("window %d not found", windowID)
		}
		return err
	}
	return nil
}

func isWindowNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "window not found") || strings.Contains(text, "no window with id")
}
