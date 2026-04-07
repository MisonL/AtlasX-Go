package tabs

import "errors"

func (c Client) SetWindowBounds(windowID int, left int, top int, width int, height int) (WindowBounds, error) {
	if windowID <= 0 {
		return WindowBounds{}, errors.New("window id must be positive")
	}
	if width <= 0 {
		return WindowBounds{}, errors.New("width must be positive")
	}
	if height <= 0 {
		return WindowBounds{}, errors.New("height must be positive")
	}

	browserWS, err := c.browserWS()
	if err != nil {
		return WindowBounds{}, err
	}

	if err := c.setWindowBounds(browserWS, windowID, map[string]any{
		"left":   left,
		"top":    top,
		"width":  width,
		"height": height,
	}, 70); err != nil {
		return WindowBounds{}, err
	}

	bounds, err := c.getWindowBounds(browserWS, windowID, 71)
	if err != nil {
		return WindowBounds{}, err
	}
	return WindowBounds{
		WindowID: windowID,
		State:    bounds.WindowState,
		Left:     bounds.Left,
		Top:      bounds.Top,
		Width:    bounds.Width,
		Height:   bounds.Height,
	}, nil
}
