package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"atlasx/internal/blueprint"
	"atlasx/internal/diagnostics"
	"atlasx/internal/launcher"
	"atlasx/internal/platform/macos"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "atlasctl: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("missing command: blueprint, doctor, launch-webapp, status, stop-webapp")
	}

	switch args[0] {
	case "blueprint":
		fmt.Print(blueprint.Render())
		return nil
	case "doctor":
		report, err := diagnostics.Generate()
		if err != nil {
			return err
		}
		fmt.Print(report.Render())
		return nil
	case "launch-webapp":
		return runLaunch(args[1:])
	case "status":
		return runStatus()
	case "stop-webapp":
		return runStop()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runLaunch(args []string) error {
	fs := flag.NewFlagSet("launch-webapp", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	dryRun := fs.Bool("dry-run", false, "print launch plan without starting Chrome")
	sharedProfile := fs.Bool("shared-profile", false, "reuse the user Chrome profile")
	url := fs.String("url", "", "override Atlas web entry URL")

	if err := fs.Parse(args); err != nil {
		return err
	}

	result, err := launcher.Run(launcher.Options{
		DryRun:        *dryRun,
		UseSharedMode: *sharedProfile,
		URLOverride:   *url,
	})
	if err != nil {
		return err
	}

	fmt.Print(result.Render())
	return nil
}

func runStatus() error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	report, err := launcher.Status(paths)
	if err != nil {
		return err
	}
	fmt.Print(report.Render())
	return nil
}

func runStop() error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	report, err := launcher.Stop(paths, 3*time.Second)
	if err != nil {
		return err
	}
	fmt.Print(report.Render())
	return nil
}
