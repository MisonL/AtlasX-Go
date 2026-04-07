package tabs

import "fmt"

type WindowActivateResult struct {
	WindowID          int    `json:"window_id"`
	ActivatedTargetID string `json:"activated_target_id"`
}

func (c Client) ActivateWindow(windowID int) (WindowActivateResult, error) {
	windows, err := c.Windows()
	if err != nil {
		return WindowActivateResult{}, err
	}

	for _, window := range windows {
		if window.WindowID != windowID {
			continue
		}
		if len(window.Targets) == 0 {
			return WindowActivateResult{}, fmt.Errorf("window %d has no page targets", windowID)
		}

		targetID := window.Targets[0].ID
		if err := c.Activate(targetID); err != nil {
			return WindowActivateResult{}, err
		}
		return WindowActivateResult{
			WindowID:          windowID,
			ActivatedTargetID: targetID,
		}, nil
	}

	return WindowActivateResult{}, fmt.Errorf("window %d not found", windowID)
}
