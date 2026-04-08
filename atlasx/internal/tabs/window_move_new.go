package tabs

import "fmt"

type WindowMoveToNewResult struct {
	SourceWindowID int    `json:"source_window_id"`
	SourceTargetID string `json:"source_target_id"`
	Target         Target `json:"target"`
}

func (c Client) MoveToNewWindow(targetID string) (WindowMoveToNewResult, error) {
	windows, err := c.Windows()
	if err != nil {
		return WindowMoveToNewResult{}, err
	}

	var sourceTarget Target
	sourceWindowID := 0
	for _, window := range windows {
		for _, target := range window.Targets {
			if target.ID != targetID {
				continue
			}
			sourceTarget = target
			sourceWindowID = window.WindowID
		}
	}

	if sourceWindowID == 0 {
		return WindowMoveToNewResult{}, fmt.Errorf("page target %s not found", targetID)
	}

	opened, err := c.OpenWindow(sourceTarget.URL)
	if err != nil {
		return WindowMoveToNewResult{}, err
	}
	if err := c.Close(sourceTarget.ID); err != nil {
		return WindowMoveToNewResult{}, err
	}

	return WindowMoveToNewResult{
		SourceWindowID: sourceWindowID,
		SourceTargetID: sourceTarget.ID,
		Target:         opened,
	}, nil
}
