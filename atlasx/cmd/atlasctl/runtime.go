package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/macos"
)

var runManagedRuntimeInstall = func(paths macos.Paths) (managedruntime.InstallReport, error) {
	return managedruntime.Install(paths, managedruntime.InstallOptions{})
}

func runRuntime(args []string) error {
	if len(args) == 0 {
		return errors.New("missing runtime subcommand: stage, status, verify, clear, install, plan")
	}

	switch args[0] {
	case "stage":
		return runRuntimeStage(args[1:])
	case "status":
		return runRuntimeStatus()
	case "verify":
		return runRuntimeVerify()
	case "clear":
		return runRuntimeClear()
	case "install":
		return runRuntimeInstall()
	case "plan":
		return runRuntimePlan(args[1:])
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

func runRuntimeVerify() error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	report, err := managedruntime.Verify(paths)
	fmt.Print(report.Render())
	if err != nil {
		return err
	}

	return nil
}

func runRuntimeInstall() error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	report, err := runManagedRuntimeInstall(paths)
	fmt.Print(report.Render())
	if err != nil {
		return err
	}
	return nil
}

func runRuntimePlan(args []string) error {
	if len(args) == 0 {
		return errors.New("missing runtime plan subcommand: create, status, clear")
	}

	switch args[0] {
	case "create":
		return runRuntimePlanCreate(args[1:])
	case "status":
		return runRuntimePlanStatus()
	case "clear":
		return runRuntimePlanClear()
	default:
		return fmt.Errorf("unknown runtime plan subcommand %q", args[0])
	}
}

func runRuntimePlanCreate(args []string) error {
	fs := flag.NewFlagSet("runtime plan create", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	version := fs.String("version", "", "managed runtime version")
	channel := fs.String("channel", "", "managed runtime channel")
	sourceURL := fs.String("url", "", "managed runtime archive url")
	expectedSHA256 := fs.String("sha256", "", "expected archive sha256")
	archivePath := fs.String("archive-path", "", "local archive path to reserve for download")
	stagedBundlePath := fs.String("bundle-path", "", "target staged bundle path after install")

	if err := fs.Parse(args); err != nil {
		return err
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	plan, err := managedruntime.NewInstallPlan(managedruntime.InstallPlanOptions{
		Version:          *version,
		Channel:          *channel,
		SourceURL:        *sourceURL,
		ExpectedSHA256:   *expectedSHA256,
		ArchivePath:      *archivePath,
		StagedBundlePath: *stagedBundlePath,
	})
	if err != nil {
		return err
	}

	if err := managedruntime.SaveInstallPlan(paths, plan); err != nil {
		return err
	}

	status, err := managedruntime.InstallPlanInfo(paths)
	if err != nil {
		return err
	}
	fmt.Print(status.Render())
	return nil
}

func runRuntimePlanStatus() error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	status, err := managedruntime.InstallPlanInfo(paths)
	if err != nil {
		return err
	}
	fmt.Print(status.Render())
	return nil
}

func runRuntimePlanClear() error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	if err := managedruntime.ClearInstallPlan(paths); err != nil {
		return err
	}
	fmt.Printf("cleared_install_plan=%s\n", paths.RuntimeInstallPlanFile)
	return nil
}
