package main

import "fmt"

func runTabsWindows(client commandTabsClient) error {
	windows, err := client.Windows()
	if err != nil {
		return err
	}

	fmt.Printf("returned=%d\n", len(windows))
	for index, window := range windows {
		fmt.Printf(
			"index=%d window_id=%d state=%s left=%d top=%d width=%d height=%d returned=%d\n",
			index,
			window.WindowID,
			window.State,
			window.Left,
			window.Top,
			window.Width,
			window.Height,
			window.Returned,
		)
		for targetIndex, target := range window.Targets {
			fmt.Printf("target_index=%d id=%s title=%q url=%s\n", targetIndex, target.ID, target.Title, target.URL)
		}
	}
	return nil
}
