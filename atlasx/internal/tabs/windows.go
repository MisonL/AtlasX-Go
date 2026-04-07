package tabs

import (
	"encoding/json"
	"sort"
)

type WindowSummary struct {
	WindowID int      `json:"window_id"`
	State    string   `json:"state"`
	Left     int      `json:"left"`
	Top      int      `json:"top"`
	Width    int      `json:"width"`
	Height   int      `json:"height"`
	Returned int      `json:"returned"`
	Targets  []Target `json:"targets"`
}

type windowForTargetResult struct {
	WindowID int          `json:"windowId"`
	Bounds   windowBounds `json:"bounds"`
}

type windowBounds struct {
	Left        int    `json:"left"`
	Top         int    `json:"top"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	WindowState string `json:"windowState"`
}

func (c Client) Windows() ([]WindowSummary, error) {
	targets, err := c.List()
	if err != nil {
		return nil, err
	}

	pages := PageTargets(targets)
	if len(pages) == 0 {
		return []WindowSummary{}, nil
	}

	browserWS, err := c.browserWS()
	if err != nil {
		return nil, err
	}

	grouped := make(map[int]*WindowSummary, len(pages))
	for index, target := range pages {
		window, err := c.windowForTarget(browserWS, target, 40+index)
		if err != nil {
			return nil, err
		}

		summary, ok := grouped[window.WindowID]
		if !ok {
			summary = &WindowSummary{
				WindowID: window.WindowID,
				State:    window.Bounds.WindowState,
				Left:     window.Bounds.Left,
				Top:      window.Bounds.Top,
				Width:    window.Bounds.Width,
				Height:   window.Bounds.Height,
				Targets:  make([]Target, 0, 1),
			}
			grouped[window.WindowID] = summary
		}
		summary.Targets = append(summary.Targets, target)
		summary.Returned = len(summary.Targets)
	}

	windowIDs := make([]int, 0, len(grouped))
	for windowID := range grouped {
		windowIDs = append(windowIDs, windowID)
	}
	sort.Ints(windowIDs)

	windows := make([]WindowSummary, 0, len(windowIDs))
	for _, windowID := range windowIDs {
		windows = append(windows, *grouped[windowID])
	}
	return windows, nil
}

func (c Client) windowForTarget(browserWS string, target Target, requestID int) (windowForTargetResult, error) {
	response, err := runPageCommand(browserWS, cdpCommandRequest{
		ID:     requestID,
		Method: "Browser.getWindowForTarget",
		Params: map[string]any{
			"targetId": target.ID,
		},
	})
	if err != nil {
		return windowForTargetResult{}, err
	}

	var payload windowForTargetResult
	if err := json.Unmarshal(response.Result, &payload); err != nil {
		return windowForTargetResult{}, err
	}
	return payload, nil
}
