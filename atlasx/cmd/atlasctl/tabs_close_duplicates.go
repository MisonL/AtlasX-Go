package main

import "fmt"

func runTabsCloseDuplicates(client commandTabsClient) error {
	result, err := client.CloseDuplicates()
	if err != nil {
		return err
	}

	fmt.Printf("returned=%d\n", result.Returned)
	fmt.Printf("groups_returned=%d\n", len(result.Groups))
	for index, group := range result.Groups {
		fmt.Printf("index=%d url=%s kept_target_id=%s returned=%d\n", index, group.URL, group.KeptTargetID, group.Returned)
		for targetIndex, targetID := range group.ClosedTargetIDs {
			fmt.Printf("closed_target_index=%d id=%s\n", targetIndex, targetID)
		}
	}
	for index, targetID := range result.ClosedTargets {
		fmt.Printf("closed_index=%d id=%s\n", index, targetID)
	}
	return nil
}
