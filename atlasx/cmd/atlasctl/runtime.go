package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/macos"
)

func runRuntime(args []string) error {
	if len(args) == 0 {
		return errors.New("missing runtime subcommand: stage, status, clear")
	}

	switch args[0] {
	case "stage":
		return runRuntimeStage(args[1:])
	case "status":
		return runRuntimeStatus()
	case "clear":
		return runRuntimeClear()
	default:
		return fmt.Errorf("unknown runtime subcommand %q", args[0])
	}
}

func runRuntimeStage(args []string) error {
	fs := flag.NewFlagSet("runtime stage", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	bundlePath := fs.String("bundle-path", "", "path to a local Chromium.app bundle")
	version := fs.String("version", "", "managed runtime version")
	channel := fs.String("channel", "local", "managed runtime channel")

	if err := fs.Parse(args); err != nil {
		return err
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	report, err := managedruntime.StageLocal(paths, managedruntime.StageOptions{
		BundlePath: *bundlePath,
		Version:    *version,
		Channel:    *channel,
	})
	if err != nil {
		return err
	}

	fmt.Print(report.Render())
	return nil
}

func runRuntimeStatus() error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	report, err := managedruntime.Status(paths)
	if err != nil {
		return err
	}

	fmt.Print(report.Render())
	return nil
}

func runRuntimeClear() error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	if err := managedruntime.Clear(paths); err != nil {
		return err
	}

	fmt.Printf("cleared_runtime_root=%s\n", paths.RuntimeRoot)
	return nil
}
