package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"atlasx/internal/blueprint"
	"atlasx/internal/browserdata"
	"atlasx/internal/diagnostics"
	"atlasx/internal/imports"
	"atlasx/internal/launcher"
	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
	"atlasx/internal/tabs"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "atlasctl: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("missing command: blueprint, doctor, launch-webapp, status, stop-webapp, mirror-scan, tabs, import-chrome, import-safari, history, downloads, bookmarks")
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

	snapshot, err := mirror.Collect(targetProfileDir)
	if err != nil {
		return err
	}
	if err := mirror.Save(paths, snapshot); err != nil {
		return err
	}

	fmt.Print(snapshot.Render(paths))
	return nil
}

func runTabs(args []string) error {
	if len(args) == 0 {
		return errors.New("missing tabs subcommand: list, open, activate, close, navigate")
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	client, err := tabs.New(paths)
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
	default:
		return fmt.Errorf("unknown tabs subcommand %q", args[0])
	}
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

func runHistory(args []string) error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	rows, err := browserdata.LoadHistory(paths)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("history supports subcommands: list, open")
	}

	switch args[0] {
	case "list":
		for index, row := range rows {
			fmt.Printf("index=%d last_visit_time=%s visit_count=%d title=%q url=%s\n", index, row.LastVisitTime, row.VisitCount, row.Title, row.URL)
		}
		return nil
	case "open":
		index, err := parseIndexArg(args, "history open")
		if err != nil {
			return err
		}
		if index >= len(rows) {
			return fmt.Errorf("history index %d out of range", index)
		}
		client, err := tabs.New(paths)
		if err != nil {
			return err
		}
		target, err := client.Open(rows[index].URL)
		if err != nil {
			return err
		}
		fmt.Printf("opened_history_index=%d id=%s url=%s\n", index, target.ID, target.URL)
		return nil
	default:
		return fmt.Errorf("unknown history subcommand %q", args[0])
	}
}

func runDownloads(args []string) error {
	if len(args) == 0 || args[0] != "list" {
		return errors.New("downloads supports only subcommand: list")
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	rows, err := browserdata.LoadDownloads(paths)
	if err != nil {
		return err
	}
	for index, row := range rows {
		fmt.Printf("index=%d end_time=%s total_bytes=%d state=%d target_path=%s tab_url=%s\n", index, row.EndTime, row.TotalBytes, row.State, row.TargetPath, row.TabURL)
	}
	return nil
}

func runBookmarks(args []string) error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	rows, err := browserdata.LoadBookmarks(paths)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("bookmarks supports subcommands: list, open")
	}

	switch args[0] {
	case "list":
		for index, row := range rows {
			fmt.Printf("index=%d root=%s name=%q url=%s\n", index, row.Root, row.Name, row.URL)
		}
		return nil
	case "open":
		index, err := parseIndexArg(args, "bookmarks open")
		if err != nil {
			return err
		}
		if index >= len(rows) {
			return fmt.Errorf("bookmark index %d out of range", index)
		}
		client, err := tabs.New(paths)
		if err != nil {
			return err
		}
		target, err := client.Open(rows[index].URL)
		if err != nil {
			return err
		}
		fmt.Printf("opened_bookmark_index=%d id=%s url=%s\n", index, target.ID, target.URL)
		return nil
	default:
		return fmt.Errorf("unknown bookmarks subcommand %q", args[0])
	}
}

func parseIndexArg(args []string, commandName string) (int, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("missing index for %s", commandName)
	}
	index, err := strconv.Atoi(args[1])
	if err != nil {
		return 0, fmt.Errorf("invalid index %q", args[1])
	}
	if index < 0 {
		return 0, fmt.Errorf("index must be >= 0")
	}
	return index, nil
}
