package tabgroups

import (
	"fmt"

	"atlasx/internal/tabs"
)

type OrganizerClient interface {
	List() ([]tabs.Target, error)
	Windows() ([]tabs.WindowSummary, error)
	MoveToWindow(string, int) (tabs.WindowMoveResult, error)
	MoveToNewWindow(string) (tabs.WindowMoveToNewResult, error)
}

type AppliedTarget struct {
	SourceWindowID    int         `json:"source_window_id"`
	SourceTargetID    string      `json:"source_target_id"`
	ActivatedTargetID string      `json:"activated_target_id"`
	Target            tabs.Target `json:"target"`
}

type ApplyResult struct {
	GroupID      string          `json:"group_id"`
	Label        string          `json:"label"`
	WindowID     int             `json:"window_id"`
	Returned     int             `json:"returned"`
	MovedTargets []AppliedTarget `json:"moved_targets"`
}

func ApplyToNewWindow(client OrganizerClient, groupID string) (ApplyResult, error) {
	targets, err := client.List()
	if err != nil {
		return ApplyResult{}, err
	}

	group, err := findGroup(Suggest(targets), groupID)
	if err != nil {
		return ApplyResult{}, err
	}
	if len(group.Targets) < 2 {
		return ApplyResult{}, fmt.Errorf("group %s has fewer than 2 page targets", groupID)
	}

	firstMove, err := client.MoveToNewWindow(group.Targets[0].ID)
	if err != nil {
		return ApplyResult{}, err
	}

	windowID, err := windowIDForTarget(client, firstMove.Target.ID)
	if err != nil {
		return ApplyResult{}, err
	}

	result := ApplyResult{
		GroupID:      group.ID,
		Label:        group.Label,
		WindowID:     windowID,
		MovedTargets: make([]AppliedTarget, 0, len(group.Targets)),
	}
	result.MovedTargets = append(result.MovedTargets, AppliedTarget{
		SourceWindowID: firstMove.SourceWindowID,
		SourceTargetID: firstMove.SourceTargetID,
		Target:         firstMove.Target,
	})

	for _, target := range group.Targets[1:] {
		moved, err := client.MoveToWindow(target.ID, windowID)
		if err != nil {
			return ApplyResult{}, err
		}
		result.MovedTargets = append(result.MovedTargets, AppliedTarget{
			SourceWindowID:    moved.SourceWindowID,
			SourceTargetID:    moved.SourceTargetID,
			ActivatedTargetID: moved.ActivatedTargetID,
			Target:            moved.Target,
		})
	}

	result.Returned = len(result.MovedTargets)
	return result, nil
}

func findGroup(groups []Group, groupID string) (Group, error) {
	for _, group := range groups {
		if group.ID == groupID {
			return group, nil
		}
	}
	return Group{}, fmt.Errorf("group %s not found", groupID)
}

func windowIDForTarget(client OrganizerClient, targetID string) (int, error) {
	windows, err := client.Windows()
	if err != nil {
		return 0, err
	}
	for _, window := range windows {
		for _, target := range window.Targets {
			if target.ID == targetID {
				return window.WindowID, nil
			}
		}
	}
	return 0, fmt.Errorf("window for target %s not found", targetID)
}
