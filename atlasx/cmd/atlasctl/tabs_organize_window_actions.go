package main

import (
	"errors"
	"fmt"
	"strconv"

	"atlasx/internal/tabgroups"
)

func runTabsOrganizeIntoWindow(client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing window id for tabs organize-into-window")
	}
	windowID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid window id %q", args[0])
	}
	result, err := tabgroups.ApplyAllToWindow(client, windowID)
	if err != nil {
		return err
	}
	printApplyAllResult(result)
	return nil
}

func runTabsOrganizeWindowToWindows(client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing source window id for tabs organize-window-to-windows")
	}
	sourceWindowID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid source window id %q", args[0])
	}
	result, err := tabgroups.ApplyWindowToNewWindows(client, sourceWindowID)
	if err != nil {
		return err
	}
	fmt.Printf("source_window_id=%d returned=%d\n", result.SourceWindowID, result.Returned)
	for groupIndex, group := range result.Groups {
		printApplyGroupResult(groupIndex, group)
	}
	return nil
}

func runTabsOrganizeWindowIntoWindow(client commandTabsClient, args []string) error {
	if len(args) < 2 {
		return errors.New("missing source window id or target window id for tabs organize-window-into-window")
	}
	sourceWindowID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid source window id %q", args[0])
	}
	targetWindowID, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid target window id %q", args[1])
	}
	result, err := tabgroups.ApplyWindowToWindow(client, sourceWindowID, targetWindowID)
	if err != nil {
		return err
	}
	fmt.Printf("source_window_id=%d target_window_id=%d returned=%d\n", result.SourceWindowID, result.TargetWindowID, result.Returned)
	for groupIndex, group := range result.Groups {
		printApplyGroupResult(groupIndex, group)
	}
	return nil
}

func runTabsOrganizeWindowGroupToWindow(client commandTabsClient, args []string) error {
	if len(args) < 2 {
		return errors.New("missing source window id or group id for tabs organize-window-group-to-window")
	}
	sourceWindowID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid source window id %q", args[0])
	}
	result, err := tabgroups.ApplyWindowGroupToNewWindow(client, sourceWindowID, args[1])
	if err != nil {
		return err
	}
	fmt.Printf("source_window_id=%d group_id=%s label=%q window_id=%d returned=%d\n", result.SourceWindowID, result.GroupID, result.Label, result.WindowID, result.Returned)
	for index, moved := range result.MovedTargets {
		fmt.Printf(
			"moved_index=%d source_window_id=%d source_target_id=%s activated_target_id=%s id=%s type=%s title=%q url=%s\n",
			index,
			moved.SourceWindowID,
			moved.SourceTargetID,
			moved.ActivatedTargetID,
			moved.Target.ID,
			moved.Target.Type,
			moved.Target.Title,
			moved.Target.URL,
		)
	}
	return nil
}

func runTabsOrganizeWindowGroupIntoWindow(client commandTabsClient, args []string) error {
	if len(args) < 3 {
		return errors.New("missing source window id, group id, or target window id for tabs organize-window-group-into-window")
	}
	sourceWindowID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid source window id %q", args[0])
	}
	targetWindowID, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid target window id %q", args[2])
	}
	result, err := tabgroups.ApplyWindowGroupToWindow(client, sourceWindowID, args[1], targetWindowID)
	if err != nil {
		return err
	}
	fmt.Printf(
		"source_window_id=%d target_window_id=%d group_id=%s label=%q window_id=%d returned=%d\n",
		result.SourceWindowID,
		result.TargetWindowID,
		result.GroupID,
		result.Label,
		result.WindowID,
		result.Returned,
	)
	for index, moved := range result.MovedTargets {
		fmt.Printf(
			"moved_index=%d source_window_id=%d source_target_id=%s activated_target_id=%s id=%s type=%s title=%q url=%s\n",
			index,
			moved.SourceWindowID,
			moved.SourceTargetID,
			moved.ActivatedTargetID,
			moved.Target.ID,
			moved.Target.Type,
			moved.Target.Title,
			moved.Target.URL,
		)
	}
	for index, aligned := range result.AlignedTargets {
		fmt.Printf(
			"aligned_index=%d source_window_id=%d source_target_id=%s id=%s type=%s title=%q url=%s\n",
			index,
			aligned.SourceWindowID,
			aligned.SourceTargetID,
			aligned.Target.ID,
			aligned.Target.Type,
			aligned.Target.Title,
			aligned.Target.URL,
		)
	}
	return nil
}

func printApplyAllResult(result tabgroups.ApplyAllResult) {
	fmt.Printf("returned=%d\n", result.Returned)
	for groupIndex, group := range result.Groups {
		printApplyGroupResult(groupIndex, group)
	}
}

func printApplyGroupResult(groupIndex int, group tabgroups.ApplyResult) {
	fmt.Printf("group_index=%d group_id=%s label=%q window_id=%d returned=%d\n", groupIndex, group.GroupID, group.Label, group.WindowID, group.Returned)
	for index, moved := range group.MovedTargets {
		fmt.Printf(
			"moved_index=%d source_window_id=%d source_target_id=%s activated_target_id=%s id=%s type=%s title=%q url=%s\n",
			index,
			moved.SourceWindowID,
			moved.SourceTargetID,
			moved.ActivatedTargetID,
			moved.Target.ID,
			moved.Target.Type,
			moved.Target.Title,
			moved.Target.URL,
		)
	}
	for index, aligned := range group.AlignedTargets {
		fmt.Printf(
			"aligned_index=%d source_window_id=%d source_target_id=%s id=%s type=%s title=%q url=%s\n",
			index,
			aligned.SourceWindowID,
			aligned.SourceTargetID,
			aligned.Target.ID,
			aligned.Target.Type,
			aligned.Target.Title,
			aligned.Target.URL,
		)
	}
}
