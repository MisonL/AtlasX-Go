package main

import (
	"errors"
	"flag"
	"fmt"

	"atlasx/internal/logs"
	"atlasx/internal/platform/macos"
)

func runLogs(args []string) error {
	if len(args) == 0 {
		return errors.New("logs supports subcommands: status")
	}

	switch args[0] {
	case "status":
		return runLogsStatus(args[1:])
	default:
		return fmt.Errorf("unknown logs subcommand %q", args[0])
	}
}

func runLogsStatus(args []string) error {
	fs := flag.NewFlagSet("logs status", flag.ContinueOnError)
	fs.SetOutput(discardCommandOutput{})

	limit := fs.Int("limit", logs.DefaultRecentFilesLimit, "return only the most recent N log files")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *limit < 0 {
		return errors.New("limit must be >= 0")
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	status, err := logs.LoadStatus(paths, *limit)
	if err != nil {
		return err
	}

	fmt.Print(status.Render())
	return nil
}
