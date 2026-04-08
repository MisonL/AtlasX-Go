package tabs

import "fmt"

type WindowMoveResult struct {
	SourceWindowID    int    `json:"source_window_id"`
	TargetWindowID    int    `json:"target_window_id"`
	SourceTargetID    string `json:"source_target_id"`
	ActivatedTargetID string `json:"activated_target_id"`
	Target            Target `json:"target"`
}

func (c Client) MoveToWindow(targetID string, targetWindowID int) (WindowMoveResult, error) {
	windows, err := c.Windows()
	if err != nil {
		return WindowMoveResult{}, err
	}

	var sourceTarget Target
	sourceWindowID := 0
	targetExists := false
	for _, window := range windows {
		if window.WindowID == targetWindowID {
			targetExists = true
		}
		for _, target := range window.Targets {
			if target.ID != targetID {
				continue
			}
			sourceTarget = target
			sourceWindowID = window.WindowID
		}
	}

	if sourceWindowID == 0 {
		return WindowMoveResult{}, fmt.Errorf("page target %s not found", targetID)
	}
	if !targetExists {
		return WindowMoveResult{}, fmt.Errorf("window %d not found", targetWindowID)
	}
	if sourceWindowID == targetWindowID {
		return WindowMoveResult{}, fmt.Errorf("source and target window ids must differ")
	}

	opened, err := c.OpenInWindow(targetWindowID, sourceTarget.URL)
	if err != nil {
		return WindowMoveResult{}, err
	}
	if err := c.Close(sourceTarget.ID); err != nil {
		return WindowMoveResult{}, err
	}

	return WindowMoveResult{
		SourceWindowID:    sourceWindowID,
		TargetWindowID:    targetWindowID,
		SourceTargetID:    sourceTarget.ID,
		ActivatedTargetID: opened.ActivatedTargetID,
		Target:            opened.Target,
	}, nil
}
