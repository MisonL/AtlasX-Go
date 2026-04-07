package tabs

import "fmt"

type WindowCloseResult struct {
	WindowID      int      `json:"window_id"`
	Returned      int      `json:"returned"`
	ClosedTargets []string `json:"closed_targets"`
}

func (c Client) CloseWindow(windowID int) (WindowCloseResult, error) {
	windows, err := c.Windows()
	if err != nil {
		return WindowCloseResult{}, err
	}

	for _, window := range windows {
		if window.WindowID != windowID {
			continue
		}

		closedTargets := make([]string, 0, len(window.Targets))
		for _, target := range window.Targets {
			if err := c.Close(target.ID); err != nil {
				return WindowCloseResult{}, err
			}
			closedTargets = append(closedTargets, target.ID)
		}
		return WindowCloseResult{
			WindowID:      windowID,
			Returned:      len(closedTargets),
			ClosedTargets: closedTargets,
		}, nil
	}

	return WindowCloseResult{}, fmt.Errorf("window %d not found", windowID)
}
