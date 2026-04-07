package main

import (
	"errors"
	"fmt"
)

func runTabsEmulateDevice(client commandTabsClient, args []string) error {
	if len(args) < 2 {
		return errors.New("missing target id or preset for tabs emulate-device")
	}

	result, err := client.EmulateDevice(args[0], args[1])
	if err != nil {
		return err
	}

	fmt.Printf(
		"id=%s title=%q url=%s preset=%s viewport=%s mobile=%t touch=%t\n",
		result.ID,
		result.Title,
		result.URL,
		result.Preset,
		result.Viewport,
		result.Mobile,
		result.Touch,
	)
	return nil
}
