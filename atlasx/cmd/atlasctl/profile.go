package main

import (
	"errors"
	"fmt"

	"atlasx/internal/platform/macos"
	"atlasx/internal/profile"
)

func runProfile(args []string) error {
	if len(args) == 0 {
		return errors.New("profile supports subcommands: status")
	}

	switch args[0] {
	case "status":
		if len(args) != 1 {
			return errors.New("profile status accepts no arguments")
		}

		paths, err := macos.DiscoverPaths()
		if err != nil {
			return err
		}

		view, err := profile.LoadView(paths)
		if err != nil {
			return err
		}
		fmt.Print(view.Render())
		return nil
	default:
		return fmt.Errorf("unknown profile subcommand %q", args[0])
	}
}
