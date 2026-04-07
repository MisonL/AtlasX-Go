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
	"atlasx/internal/memory"
	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
	"atlasx/internal/sourcepaths"
	"atlasx/internal/tabs"
)

var newCommandTabsClient = func(paths macos.Paths) (commandTabsClient, error) {
	return tabs.New(paths)
}

type commandTabsClient interface {
	List() ([]tabs.Target, error)
	Open(string) (tabs.Target, error)
	Activate(string) error
	Close(string) error
	Navigate(string, string) error
	Capture(string) (tabs.PageContext, error)
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "atlasctl: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("missing command: blueprint, doctor, launch-webapp, status, stop-webapp, runtime, mirror-scan, tabs, import-chrome, import-safari, history, downloads, bookmarks")
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
	case "runtime":
		return runRuntime(args[1:])
	case "mirror-scan":
		return runMirrorScan(args[1:])
	case "tabs":
		return runTabs(args[1:])
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

func runTabs(args []string) error {
	if len(args) == 0 {
		return errors.New("missing tabs subcommand: list, open, activate, close, navigate, capture")
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	client, err := newCommandTabsClient(paths)
	if err != nil {
		return err
	}

	switch args[0] {
	case "list":
		targets, err := client.List()
		if err != nil {
			return err
		}
		for _, target := range tabs.PageTargets(targets) {
			fmt.Printf("id=%s type=%s title=%q url=%s\n", target.ID, target.Type, target.Title, target.URL)
		}
		return nil
	case "open":
		if len(args) < 2 {
			return errors.New("missing url for tabs open")
		}
		target, err := client.Open(args[1])
		if err != nil {
			return err
		}
		fmt.Printf("id=%s type=%s title=%q url=%s\n", target.ID, target.Type, target.Title, target.URL)
		return nil
	case "activate":
		if len(args) < 2 {
			return errors.New("missing target id for tabs activate")
		}
		if err := client.Activate(args[1]); err != nil {
			return err
		}
		fmt.Printf("activated=%s\n", args[1])
		return nil
	case "close":
		if len(args) < 2 {
			return errors.New("missing target id for tabs close")
		}
		if err := client.Close(args[1]); err != nil {
			return err
		}
		fmt.Printf("closed=%s\n", args[1])
		return nil
	case "navigate":
		if len(args) < 3 {
			return errors.New("missing target id or url for tabs navigate")
		}
		if err := client.Navigate(args[1], args[2]); err != nil {
			return err
		}
		fmt.Printf("navigated=%s url=%s\n", args[1], args[2])
		return nil
	case "capture":
		if len(args) < 2 {
			return errors.New("missing target id for tabs capture")
		}
		context, err := client.Capture(args[1])
		if err != nil {
			printPageContext(context)
			return err
		}
		if err := memory.AppendPageCapture(paths, memory.PageCaptureInput{
			OccurredAt: context.CapturedAt,
			TabID:      context.ID,
			Title:      context.Title,
			URL:        context.URL,
		}); err != nil {
			return err
		}
		printPageContext(context)
		return nil
	default:
		return fmt.Errorf("unknown tabs subcommand %q", args[0])
	}
}

func printPageContext(context tabs.PageContext) {
	fmt.Printf(
		"id=%s title=%q url=%s captured_at=%s text_length=%d text_limit=%d text_truncated=%t capture_error=%q\n",
		context.ID,
		context.Title,
		context.URL,
		context.CapturedAt,
		context.TextLength,
		context.TextLimit,
		context.TextTruncated,
		context.CaptureError,
	)
	fmt.Printf("text=%q\n", context.Text)
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
