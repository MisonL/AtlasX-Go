package tabs

import (
	"errors"
	"fmt"
)

type WindowBounds struct {
	WindowID int    `json:"window_id"`
	State    string `json:"state"`
	Left     int    `json:"left"`
	Top      int    `json:"top"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

func (c Client) SetWindowState(windowID int, state string) (WindowBounds, error) {
	if windowID <= 0 {
		return WindowBounds{}, errors.New("window id must be positive")
	}
	if !isSupportedWindowState(state) {
		return WindowBounds{}, fmt.Errorf("unknown window state %q", state)
	}

	browserWS, err := c.browserWS()
	if err != nil {
		return WindowBounds{}, err
	}

	if err := c.setWindowBounds(browserWS, windowID, map[string]any{"windowState": state}, 60); err != nil {
		return WindowBounds{}, err
	}

	bounds, err := c.getWindowBounds(browserWS, windowID, 61)
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

func isSupportedWindowState(state string) bool {
	switch state {
	case "normal", "minimized", "maximized", "fullscreen":
		return true
	default:
		return false
	}
}

func (c Client) setWindowBounds(browserWS string, windowID int, bounds map[string]any, requestID int) error {
	_, err := runPageCommand(browserWS, cdpCommandRequest{
		ID:     requestID,
		Method: "Browser.setWindowBounds",
		Params: map[string]any{
			"windowId": windowID,
			"bounds":   bounds,
		},
	})
	return err
}
