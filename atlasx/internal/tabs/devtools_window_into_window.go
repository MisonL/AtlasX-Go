package tabs

import "fmt"

type DevToolsWindowOpenTarget struct {
	SourceTargetID    string `json:"source_target_id"`
	ActivatedTargetID string `json:"activated_target_id"`
	Target            Target `json:"target"`
}

type DevToolsWindowOpenResult struct {
	SourceWindowID int                        `json:"source_window_id"`
	TargetWindowID int                        `json:"target_window_id"`
	Returned       int                        `json:"returned"`
	OpenedTargets  []DevToolsWindowOpenTarget `json:"opened_targets"`
}

func (c Client) OpenDevToolsWindowIntoWindow(sourceWindowID int, targetWindowID int) (DevToolsWindowOpenResult, error) {
	if sourceWindowID == targetWindowID {
		return DevToolsWindowOpenResult{}, fmt.Errorf("source and target window ids must differ")
	}

	windows, err := c.Windows()
	if err != nil {
		return DevToolsWindowOpenResult{}, err
	}

	var sourceWindow *WindowSummary
	targetExists := false
	for index := range windows {
		switch windows[index].WindowID {
		case sourceWindowID:
			sourceWindow = &windows[index]
		case targetWindowID:
			targetExists = true
		}
	}
	if sourceWindow == nil {
		return DevToolsWindowOpenResult{}, fmt.Errorf("window %d not found", sourceWindowID)
	}
	if !targetExists {
		return DevToolsWindowOpenResult{}, fmt.Errorf("window %d not found", targetWindowID)
	}

	result := DevToolsWindowOpenResult{
		SourceWindowID: sourceWindowID,
		TargetWindowID: targetWindowID,
		OpenedTargets:  make([]DevToolsWindowOpenTarget, 0, len(sourceWindow.Targets)),
	}
	for _, sourceTarget := range sourceWindow.Targets {
		opened, err := c.OpenDevToolsInWindow(sourceTarget.ID, targetWindowID)
		if err != nil {
			return DevToolsWindowOpenResult{}, err
		}
		result.OpenedTargets = append(result.OpenedTargets, DevToolsWindowOpenTarget{
			SourceTargetID:    sourceTarget.ID,
			ActivatedTargetID: opened.ActivatedTargetID,
			Target:            opened.Target,
		})
		result.Returned = len(result.OpenedTargets)
	}

	return result, nil
}
