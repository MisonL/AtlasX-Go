package main

import (
	"errors"
	"fmt"

	"atlasx/internal/permissions"
)

var loadPermissionsStatus = permissions.LoadStatus

func runPermissions(args []string) error {
	if len(args) == 0 {
		return errors.New("permissions supports subcommands: status")
	}

	switch args[0] {
	case "status":
		if len(args) != 1 {
			return errors.New("permissions status accepts no arguments")
		}
		fmt.Print(loadPermissionsStatus().Render())
		return nil
	default:
		return fmt.Errorf("unknown permissions subcommand %q", args[0])
	}
}
