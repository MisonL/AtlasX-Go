package main

import (
	"errors"
	"fmt"
	"strconv"

	"atlasx/internal/platform/macos"
	"atlasx/internal/tabgroups"
)

func runTabsOrganize(paths macos.Paths, client commandTabsClient) error {
	targets, err := client.List()
	if err != nil {
		return err
	}

	groups := tabgroups.Suggest(targets)
	fmt.Printf("returned=%d\n", len(groups))
	for index, group := range groups {
		fmt.Printf("index=%d group_id=%s label=%q returned=%d\n", index, group.ID, group.Label, group.Returned)
		fmt.Printf("reason=%q\n", group.Reason)
		for targetIndex, target := range group.Targets {
			fmt.Printf("target_index=%d id=%s title=%q url=%s\n", targetIndex, target.ID, target.Title, target.URL)
		}
	}
	return nil
}

func runTabsGroups(client commandTabsClient) error {
	result, err := tabgroups.Inspect(client)
	if err != nil {
		return err
	}

	fmt.Printf("inferred=%t returned=%d\n", result.Inferred, result.Returned)
	for index, group := range result.Groups {
		fmt.Printf(
			"index=%d group_id=%s label=%q returned=%d window_returned=%d inferred=%t\n",
			index,
			group.ID,
			group.Label,
			group.Returned,
			group.WindowReturned,
			group.Inferred,
		)
		fmt.Printf("reason=%q\n", group.Reason)
		for windowIndex, window := range group.Windows {
			fmt.Printf("window_index=%d window_id=%d returned=%d\n", windowIndex, window.WindowID, window.Returned)
		}
		for targetIndex, target := range group.Targets {
			fmt.Printf("target_index=%d id=%s title=%q url=%s\n", targetIndex, target.ID, target.Title, target.URL)
		}
	}
	return nil
}

func runTabsOrganizeWindow(paths macos.Paths, client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing window id for tabs organize-window")
	}
	windowID, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid window id %q", args[0])
	}

	result, err := tabgroups.SuggestWindow(client, windowID)
	if err != nil {
		return err
	}

	fmt.Printf("source_window_id=%d returned=%d\n", result.SourceWindowID, result.Returned)
	for index, group := range result.Groups {
		fmt.Printf("index=%d group_id=%s label=%q returned=%d\n", index, group.ID, group.Label, group.Returned)
		fmt.Printf("reason=%q\n", group.Reason)
		for targetIndex, target := range group.Targets {
			fmt.Printf("target_index=%d id=%s title=%q url=%s\n", targetIndex, target.ID, target.Title, target.URL)
		}
	}
	return nil
}
