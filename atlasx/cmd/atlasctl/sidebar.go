package main

import (
	"errors"
	"fmt"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
)

func runSidebar(args []string) error {
	if len(args) == 0 {
		return errors.New("missing sidebar subcommand: status")
	}

	switch args[0] {
	case "status":
		return runSidebarStatus(args[1:])
	default:
		return fmt.Errorf("unknown sidebar subcommand %q", args[0])
	}
}

func runSidebarStatus(args []string) error {
	if len(args) != 0 {
		return errors.New("sidebar status does not accept extra arguments")
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	config, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return err
	}

	status, err := sidebar.FromSettings(config).StatusWithRuntime(paths)
	if err != nil {
		return err
	}

	fmt.Print(status.Render())
	return nil
}
