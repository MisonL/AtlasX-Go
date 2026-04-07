package tabs

func (c Client) OpenDevToolsWindow(targetID string) (Target, error) {
	devToolsTarget, err := c.DevTools(targetID)
	if err != nil {
		return Target{}, err
	}
	return c.OpenWindow(devToolsTarget.DevToolsFrontendURL)
}
