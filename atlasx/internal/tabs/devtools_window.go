package tabs

func (c Client) OpenDevToolsWindow(targetID string) (Target, error) {
	devToolsTarget, err := c.DevTools(targetID)
	if err != nil {
		return Target{}, err
	}
	return c.OpenWindow(devToolsTarget.DevToolsFrontendURL)
}

func (c Client) OpenDevToolsInWindow(targetID string, windowID int) (WindowOpenResult, error) {
	devToolsTarget, err := c.DevTools(targetID)
	if err != nil {
		return WindowOpenResult{}, err
	}
	return c.OpenInWindow(windowID, devToolsTarget.DevToolsFrontendURL)
}

func (c Client) OpenDevToolsPanelWindow(targetID string, panel string) (Target, error) {
	devToolsTarget, err := c.DevTools(targetID)
	if err != nil {
		return Target{}, err
	}

	panelURL, err := resolveDevToolsPanelURL(devToolsTarget.DevToolsFrontendURL, panel)
	if err != nil {
		return Target{}, err
	}
	return c.OpenWindow(panelURL)
}

func (c Client) OpenDevToolsPanelInWindow(targetID string, panel string, windowID int) (WindowOpenResult, error) {
	devToolsTarget, err := c.DevTools(targetID)
	if err != nil {
		return WindowOpenResult{}, err
	}

	panelURL, err := resolveDevToolsPanelURL(devToolsTarget.DevToolsFrontendURL, panel)
	if err != nil {
		return WindowOpenResult{}, err
	}
	return c.OpenInWindow(windowID, panelURL)
}
