package tabs

func (c Client) OpenDevToolsWindow(targetID string) (Target, error) {
	devToolsTarget, err := c.DevTools(targetID)
	if err != nil {
		return Target{}, err
	}
	return c.OpenWindow(devToolsTarget.DevToolsFrontendURL)
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
