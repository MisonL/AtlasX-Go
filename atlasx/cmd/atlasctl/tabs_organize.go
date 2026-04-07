package main

import (
	"fmt"

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
