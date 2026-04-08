package main

import (
	"errors"
	"fmt"

	"atlasx/internal/platform/macos"
	"atlasx/internal/updates"
)

func runUpdates(args []string) error {
	if len(args) == 0 {
		return errors.New("updates supports subcommands: status")
	}

	switch args[0] {
	case "status":
		if len(args) != 1 {
			return errors.New("updates status accepts no arguments")
		}

		paths, err := macos.DiscoverPaths()
		if err != nil {
			return err
		}

		status, err := updates.LoadStatus(paths)
		if err != nil {
			return err
		}

		fmt.Print(status.Render())
		return nil
	default:
		return fmt.Errorf("unknown updates subcommand %q", args[0])
	}
}
