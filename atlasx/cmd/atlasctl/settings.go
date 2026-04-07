package main

import (
	"errors"
	"fmt"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
)

func runSettings(args []string) error {
	if len(args) > 1 {
		return errors.New("settings accepts at most one subcommand: show")
	}
	if len(args) == 1 && args[0] != "show" {
		return fmt.Errorf("unknown settings subcommand %q", args[0])
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	view, err := settings.LoadView(paths)
	if err != nil {
		return err
	}

	fmt.Print(view.Render())
	return nil
}
