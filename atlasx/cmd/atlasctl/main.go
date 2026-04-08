package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"atlasx/internal/blueprint"
	"atlasx/internal/diagnostics"
	"atlasx/internal/imports"
	"atlasx/internal/launcher"
	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
	"atlasx/internal/sourcepaths"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "atlasctl: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("missing command: blueprint, doctor, launch-webapp, status, profile, policy, settings, default-browser, logs, updates, sidebar, stop-webapp, runtime, mirror-scan, tabs, memory, import-chrome, import-safari, history, downloads, bookmarks")
	}

	switch args[0] {
	case "blueprint":
		fmt.Print(blueprint.Render())
		return nil
	case "doctor":
		return runDoctor(args[1:])
	case "launch-webapp":
		return runLaunch(args[1:])
	case "status":
		return runStatus()
	case "profile":
		return runProfile(args[1:])
	case "policy":
		return runPolicy(args[1:])
	case "settings":
		return runSettings(args[1:])
	case "default-browser":
		return runDefaultBrowser(args[1:])
	case "logs":
		return runLogs(args[1:])
	case "updates":
		return runUpdates(args[1:])
	case "sidebar":
		return runSidebar(args[1:])
	case "stop-webapp":
		return runStop()
	case "runtime":
		return runRuntime(args[1:])
	case "mirror-scan":
		return runMirrorScan(args[1:])
	case "tabs":
		return runTabs(args[1:])
	case "memory":
		return runMemory(args[1:])
	case "import-chrome":
		return runImportChrome(args[1:])
	case "import-safari":
		return runImportSafari()
	case "history":
		return runHistory(args[1:])
	case "downloads":
		return runDownloads(args[1:])
	case "bookmarks":
		return runBookmarks(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	jsonOutput := fs.Bool("json", false, "render doctor output as structured json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("doctor accepts no positional arguments")
	}

	report, err := diagnostics.Generate()
	if err != nil {
		return err
	}
	if *jsonOutput {
		fmt.Print(report.RenderJSON())
		return nil
	}
	fmt.Print(report.Render())
	return nil
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

func runMirrorScan(args []string) error {
	fs := flag.NewFlagSet("mirror-scan", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	profileDir := fs.String("profile-dir", "", "override the browser profile directory to scan")

	if err := fs.Parse(args); err != nil {
		return err
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	targetProfileDir := *profileDir
	if targetProfileDir == "" {
		targetProfileDir = mirror.DefaultProfilePath(paths)
	}
	if err := sourcepaths.ValidateMirrorProfileDir(paths, targetProfileDir); err != nil {
		return err
	}

	snapshot, err := mirror.Scan(paths, targetProfileDir)
	if err != nil {
		return err
	}

	fmt.Print(snapshot.Render(paths))
	return nil
}

func runImportChrome(args []string) error {
	fs := flag.NewFlagSet("import-chrome", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	sourceProfileDir := fs.String("source-profile-dir", "", "override the Chrome profile directory to import")

	if err := fs.Parse(args); err != nil {
		return err
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	targetSourceDir := *sourceProfileDir
	if targetSourceDir == "" {
		targetSourceDir = imports.DefaultChromeProfileDir(paths)
	}
	if err := sourcepaths.ValidateChromeImportSourceDir(paths, targetSourceDir); err != nil {
		return err
	}

	report, err := imports.ImportChrome(paths, targetSourceDir)
	if err != nil {
		return err
	}

	fmt.Print(report.Render())
	return nil
}

func runImportSafari() error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	report, err := imports.ImportSafari(paths)
	if err != nil {
		return err
	}

	fmt.Print(report.Render())
	return nil
}
