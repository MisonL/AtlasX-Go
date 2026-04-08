package tabs

type WindowOpenResult struct {
	WindowID          int    `json:"window_id"`
	ActivatedTargetID string `json:"activated_target_id"`
	Target            Target `json:"target"`
}

func (c Client) OpenInWindow(windowID int, targetURL string) (WindowOpenResult, error) {
	activated, err := c.ActivateWindow(windowID)
	if err != nil {
		return WindowOpenResult{}, err
	}

	target, err := c.Open(targetURL)
	if err != nil {
		return WindowOpenResult{}, err
	}

	return WindowOpenResult{
		WindowID:          windowID,
		ActivatedTargetID: activated.ActivatedTargetID,
		Target:            target,
	}, nil
}
