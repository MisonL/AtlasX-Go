package tabs

import "fmt"

type WindowMergeTarget struct {
	SourceTargetID    string `json:"source_target_id"`
	ActivatedTargetID string `json:"activated_target_id"`
	Target            Target `json:"target"`
}

type WindowMergeResult struct {
	SourceWindowID int                 `json:"source_window_id"`
	TargetWindowID int                 `json:"target_window_id"`
	Returned       int                 `json:"returned"`
	MovedTargets   []WindowMergeTarget `json:"moved_targets"`
}

func (c Client) MergeWindow(sourceWindowID int, targetWindowID int) (WindowMergeResult, error) {
	if sourceWindowID == targetWindowID {
		return WindowMergeResult{}, fmt.Errorf("source and target window ids must differ")
	}

	windows, err := c.Windows()
	if err != nil {
		return WindowMergeResult{}, err
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
		return WindowMergeResult{}, fmt.Errorf("window %d not found", sourceWindowID)
	}
	if !targetExists {
		return WindowMergeResult{}, fmt.Errorf("window %d not found", targetWindowID)
	}

	result := WindowMergeResult{
		SourceWindowID: sourceWindowID,
		TargetWindowID: targetWindowID,
		MovedTargets:   make([]WindowMergeTarget, 0, len(sourceWindow.Targets)),
	}
	for _, sourceTarget := range sourceWindow.Targets {
		opened, err := c.OpenInWindow(targetWindowID, sourceTarget.URL)
		if err != nil {
			return WindowMergeResult{}, err
		}
		if err := c.Close(sourceTarget.ID); err != nil {
			return WindowMergeResult{}, err
		}
		result.MovedTargets = append(result.MovedTargets, WindowMergeTarget{
			SourceTargetID:    sourceTarget.ID,
			ActivatedTargetID: opened.ActivatedTargetID,
			Target:            opened.Target,
		})
		result.Returned = len(result.MovedTargets)
	}

	return result, nil
}
