package main

import (
	"errors"
	"fmt"

	"atlasx/internal/defaultbrowser"
)

var readDefaultBrowserStatus = defaultbrowser.ReadStatus
var setDefaultBrowserBundleID = defaultbrowser.SetBundleID

func runDefaultBrowser(args []string) error {
	if len(args) == 0 {
		return errors.New("missing default-browser subcommand: status, set")
	}

	switch args[0] {
	case "status":
		if len(args) != 1 {
			return errors.New("default-browser status accepts no arguments")
		}
		status, err := readDefaultBrowserStatus()
		if err != nil {
			return err
		}
		fmt.Print(status.Render())
		return nil
	case "set":
		if len(args) != 2 {
			return errors.New("default-browser set requires a bundle id")
		}
		status, err := setDefaultBrowserBundleID(args[1])
		if err != nil {
			return err
		}
		fmt.Print(status.Render())
		return nil
	default:
		return fmt.Errorf("unknown default-browser subcommand %q", args[0])
	}
}
