package tabs

type DevToolsPanelWindowOpenResult struct {
	SourceWindowID int                        `json:"source_window_id"`
	Panel          string                     `json:"panel"`
	TargetWindowID int                        `json:"target_window_id"`
	Returned       int                        `json:"returned"`
	OpenedTargets  []DevToolsWindowOpenTarget `json:"opened_targets"`
}

func (c Client) OpenDevToolsPanelWindowIntoWindow(sourceWindowID int, panel string, targetWindowID int) (DevToolsPanelWindowOpenResult, error) {
	sourceWindow, err := c.resolveWindowPair(sourceWindowID, targetWindowID)
	if err != nil {
		return DevToolsPanelWindowOpenResult{}, err
	}

	result := DevToolsPanelWindowOpenResult{
		SourceWindowID: sourceWindowID,
		Panel:          panel,
		TargetWindowID: targetWindowID,
		OpenedTargets:  make([]DevToolsWindowOpenTarget, 0, len(sourceWindow.Targets)),
	}
	for _, sourceTarget := range sourceWindow.Targets {
		opened, err := c.OpenDevToolsPanelInWindow(sourceTarget.ID, panel, targetWindowID)
		if err != nil {
			return DevToolsPanelWindowOpenResult{}, err
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
