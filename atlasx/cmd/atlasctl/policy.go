package main

import (
	"errors"
	"fmt"

	"atlasx/internal/platform/macos"
	"atlasx/internal/policy"
)

func runPolicy(args []string) error {
	if len(args) == 0 {
		return errors.New("policy supports subcommands: status")
	}

	switch args[0] {
	case "status":
		if len(args) != 1 {
			return errors.New("policy status accepts no arguments")
		}

		paths, err := macos.DiscoverPaths()
		if err != nil {
			return err
		}

		view, err := policy.LoadView(paths)
		if err != nil {
			return err
		}

		fmt.Print(view.Render())
		return nil
	default:
		return fmt.Errorf("unknown policy subcommand %q", args[0])
	}
}
