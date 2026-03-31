package main

import (
	"errors"
	"fmt"
	"strconv"

	"atlasx/internal/browserdata"
	"atlasx/internal/platform/macos"
	"atlasx/internal/tabs"
)

func runHistory(args []string) error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	rows, err := browserdata.LoadHistory(paths)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("history supports subcommands: list, open")
	}

	switch args[0] {
	case "list":
		for index, row := range rows {
			fmt.Printf("index=%d last_visit_time=%s visit_count=%d title=%q url=%s\n", index, row.LastVisitTime, row.VisitCount, row.Title, row.URL)
		}
		return nil
	case "open":
		return runIndexedOpen(paths, args, "history open", browserdata.ResolveHistoryURL, "opened_history_index")
	default:
		return fmt.Errorf("unknown history subcommand %q", args[0])
	}
}

func runDownloads(args []string) error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	rows, err := browserdata.LoadDownloads(paths)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("downloads supports subcommands: list, open")
	}

	switch args[0] {
	case "list":
		for index, row := range rows {
			fmt.Printf("index=%d end_time=%s total_bytes=%d state=%d target_path=%s tab_url=%s\n", index, row.EndTime, row.TotalBytes, row.State, row.TargetPath, row.TabURL)
		}
		return nil
	case "open":
		return runIndexedOpen(paths, args, "downloads open", browserdata.ResolveDownloadURL, "opened_download_index")
	default:
		return fmt.Errorf("unknown downloads subcommand %q", args[0])
	}
}

func runBookmarks(args []string) error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	rows, err := browserdata.LoadBookmarks(paths)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("bookmarks supports subcommands: list, open")
	}

	switch args[0] {
	case "list":
		for index, row := range rows {
			fmt.Printf("index=%d root=%s name=%q url=%s\n", index, row.Root, row.Name, row.URL)
		}
		return nil
	case "open":
		return runIndexedOpen(paths, args, "bookmarks open", browserdata.ResolveBookmarkURL, "opened_bookmark_index")
	default:
		return fmt.Errorf("unknown bookmarks subcommand %q", args[0])
	}
}

type indexedURLResolver func(paths macos.Paths, index int) (string, error)

func runIndexedOpen(paths macos.Paths, args []string, commandName string, resolver indexedURLResolver, successKey string) error {
	index, err := parseIndexArg(args, commandName)
	if err != nil {
		return err
	}

	targetURL, err := resolver(paths, index)
	if err != nil {
		return err
	}

	client, err := tabs.New(paths)
	if err != nil {
		return err
	}

	target, err := client.Open(targetURL)
	if err != nil {
		return err
	}

	fmt.Printf("%s=%d id=%s url=%s\n", successKey, index, target.ID, target.URL)
	return nil
}

func parseIndexArg(args []string, commandName string) (int, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("missing index for %s", commandName)
	}
	index, err := strconv.Atoi(args[1])
	if err != nil {
		return 0, fmt.Errorf("invalid index %q", args[1])
	}
	if index < 0 {
		return 0, fmt.Errorf("index must be >= 0")
	}
	return index, nil
}
