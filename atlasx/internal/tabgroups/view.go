package tabgroups

import (
	"sort"

	"atlasx/internal/tabs"
)

type GroupsClient interface {
	Windows() ([]tabs.WindowSummary, error)
}

type GroupWindow struct {
	WindowID int `json:"window_id"`
	Returned int `json:"returned"`
}

type GroupView struct {
	ID             string        `json:"id"`
	Label          string        `json:"label"`
	Reason         string        `json:"reason"`
	Inferred       bool          `json:"inferred"`
	Returned       int           `json:"returned"`
	WindowReturned int           `json:"window_returned"`
	WindowIDs      []int         `json:"window_ids"`
	Windows        []GroupWindow `json:"windows"`
	Targets        []tabs.Target `json:"targets"`
}

type ViewResult struct {
	Inferred bool        `json:"inferred"`
	Returned int         `json:"returned"`
	Groups   []GroupView `json:"groups"`
}

func Inspect(client GroupsClient) (ViewResult, error) {
	windows, err := client.Windows()
	if err != nil {
		return ViewResult{}, err
	}
	return BuildView(windows), nil
}

func BuildView(windows []tabs.WindowSummary) ViewResult {
	targetWindowIDs := make(map[string]int)
	flattened := make([]tabs.Target, 0)
	for _, window := range windows {
		for _, target := range tabs.PageTargets(window.Targets) {
			flattened = append(flattened, target)
			targetWindowIDs[target.ID] = window.WindowID
		}
	}

	suggested := Suggest(flattened)
	groups := make([]GroupView, 0, len(suggested))
	for _, group := range suggested {
		windowCounts := make(map[int]int)
		for _, target := range group.Targets {
			windowID, ok := targetWindowIDs[target.ID]
			if !ok {
				continue
			}
			windowCounts[windowID]++
		}

		windowIDs := make([]int, 0, len(windowCounts))
		for windowID := range windowCounts {
			windowIDs = append(windowIDs, windowID)
		}
		sort.Ints(windowIDs)

		groupWindows := make([]GroupWindow, 0, len(windowIDs))
		for _, windowID := range windowIDs {
			groupWindows = append(groupWindows, GroupWindow{
				WindowID: windowID,
				Returned: windowCounts[windowID],
			})
		}

		groups = append(groups, GroupView{
			ID:             group.ID,
			Label:          group.Label,
			Reason:         group.Reason,
			Inferred:       true,
			Returned:       group.Returned,
			WindowReturned: len(windowIDs),
			WindowIDs:      append([]int(nil), windowIDs...),
			Windows:        groupWindows,
			Targets:        append([]tabs.Target(nil), group.Targets...),
		})
	}

	return ViewResult{
		Inferred: true,
		Returned: len(groups),
		Groups:   groups,
	}
}
