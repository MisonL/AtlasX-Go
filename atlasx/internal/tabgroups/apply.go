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
	GroupID        string          `json:"group_id"`
	Label          string          `json:"label"`
	WindowID       int             `json:"window_id"`
	Returned       int             `json:"returned"`
	MovedTargets   []AppliedTarget `json:"moved_targets"`
	AlignedTargets []AppliedTarget `json:"aligned_targets"`
}

type ApplyAllResult struct {
	Returned int           `json:"returned"`
	Groups   []ApplyResult `json:"groups"`
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
	return applyGroupToNewWindow(client, group)
}

func ApplyAllToNewWindows(client OrganizerClient) (ApplyAllResult, error) {
	targets, err := client.List()
	if err != nil {
		return ApplyAllResult{}, err
	}

	groups := Suggest(targets)
	result := ApplyAllResult{
		Groups: make([]ApplyResult, 0, len(groups)),
	}
	for _, group := range groups {
		applied, err := applyGroupToNewWindow(client, group)
		if err != nil {
			return ApplyAllResult{}, err
		}
		result.Groups = append(result.Groups, applied)
	}
	result.Returned = len(result.Groups)
	return result, nil
}

func ApplyGroupToWindow(client OrganizerClient, groupID string, windowID int) (ApplyResult, error) {
	targets, err := client.List()
	if err != nil {
		return ApplyResult{}, err
	}

	group, err := findGroup(Suggest(targets), groupID)
	if err != nil {
		return ApplyResult{}, err
	}
	return applyGroupToWindow(client, group, windowID)
}

func applyGroupToNewWindow(client OrganizerClient, group Group) (ApplyResult, error) {
	if len(group.Targets) < 2 {
		return ApplyResult{}, fmt.Errorf("group %s has fewer than 2 page targets", group.ID)
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
		GroupID:        group.ID,
		Label:          group.Label,
		WindowID:       windowID,
		MovedTargets:   make([]AppliedTarget, 0, len(group.Targets)),
		AlignedTargets: []AppliedTarget{},
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

func applyGroupToWindow(client OrganizerClient, group Group, windowID int) (ApplyResult, error) {
	if len(group.Targets) < 2 {
		return ApplyResult{}, fmt.Errorf("group %s has fewer than 2 page targets", group.ID)
	}

	windowTargets, err := sourceWindowIDs(client)
	if err != nil {
		return ApplyResult{}, err
	}
	if _, ok := windowTargets["window:"+fmt.Sprintf("%d", windowID)]; !ok {
		return ApplyResult{}, fmt.Errorf("window %d not found", windowID)
	}

	result := ApplyResult{
		GroupID:        group.ID,
		Label:          group.Label,
		WindowID:       windowID,
		MovedTargets:   make([]AppliedTarget, 0, len(group.Targets)),
		AlignedTargets: make([]AppliedTarget, 0, len(group.Targets)),
	}

	for _, target := range group.Targets {
		sourceWindowID, ok := windowTargets[target.ID]
		if !ok {
			return ApplyResult{}, fmt.Errorf("window for target %s not found", target.ID)
		}
		if sourceWindowID == windowID {
			result.AlignedTargets = append(result.AlignedTargets, AppliedTarget{
				SourceWindowID: sourceWindowID,
				SourceTargetID: target.ID,
				Target:         target,
			})
			continue
		}

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

	result.Returned = len(result.MovedTargets) + len(result.AlignedTargets)
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
	windowTargets, err := sourceWindowIDs(client)
	if err != nil {
		return 0, err
	}
	windowID, ok := windowTargets[targetID]
	if !ok {
		return 0, fmt.Errorf("window for target %s not found", targetID)
	}
	return windowID, nil
}

func sourceWindowIDs(client OrganizerClient) (map[string]int, error) {
	windows, err := client.Windows()
	if err != nil {
		return nil, err
	}
	lookup := make(map[string]int, len(windows))
	for _, window := range windows {
		lookup["window:"+fmt.Sprintf("%d", window.WindowID)] = window.WindowID
		for _, target := range window.Targets {
			lookup[target.ID] = window.WindowID
		}
	}
	return lookup, nil
}
