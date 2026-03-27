package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"atlasx/internal/blueprint"
	"atlasx/internal/diagnostics"
	"atlasx/internal/launcher"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "atlasctl: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("missing command: blueprint, doctor, launch-webapp")
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
