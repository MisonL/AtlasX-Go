package launcher

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"atlasx/internal/platform/chrome"
	"atlasx/internal/platform/macos"
	"atlasx/internal/profile"
	"atlasx/internal/settings"
)

const (
	appModePrefix           = "--app="
	autoRemoteDebuggingPort = "0"
)

type Options struct {
	DryRun        bool
	UseSharedMode bool
	URLOverride   string
}

type Result struct {
	BinaryPath    string
	RuntimeSource string
	Args          []string
	DryRun        bool
	Managed       bool
	Mode          string
	StateFile     string
	UserDataDir   string
	URL           string
}

func Run(opts Options) (Result, error) {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return Result{}, err
	}

	cfg, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return Result{}, err
	}

	mode := cfg.DefaultProfile
	if opts.UseSharedMode {
		mode = profile.ModeShared
	}

	selected, err := profile.NewStore(paths.ProfilesRoot).Resolve(mode)
	if err != nil {
		return Result{}, err
	}

	detection, err := chrome.Detect(cfg.ChromeBinary)
	if err != nil {
		return Result{}, err
	}

	url := firstNonEmpty(opts.URLOverride, cfg.WebAppURL)
	args := BuildArgs(detection.BinaryPath, selected, url)
	result := Result{
		BinaryPath:    detection.BinaryPath,
		RuntimeSource: detection.Source,
		Args:          args,
		DryRun:        opts.DryRun,
		Managed:       selected.UserDataDir != "",
		Mode:          selected.Mode,
		StateFile:     paths.SessionFile,
		UserDataDir:   selected.UserDataDir,
		URL:           url,
	}
	if opts.DryRun {
		return result, nil
	}

	cmd := exec.Command(detection.BinaryPath, args...)
	if result.Managed {
		cmd = buildManagedLaunchCommand(detection.BinaryPath, args)
	} else {
		cmd = buildSharedLaunchCommand(detection.BinaryPath, args)
	}

	if err := cmd.Start(); err != nil {
		return Result{}, err
	}

	if result.Managed {
		err = SaveState(paths, State{
			Mode:          selected.Mode,
			Managed:       true,
			RuntimeSource: detection.Source,
			BinaryPath:    detection.BinaryPath,
			Args:          args,
			URL:           url,
			UserDataDir:   selected.UserDataDir,
			StartedAt:     time.Now().UTC().Format(time.RFC3339),
		})
		if err != nil {
			return Result{}, err
		}

		if err := waitForManagedSession(paths, 5*time.Second); err != nil {
			_ = ClearState(paths)
			return Result{}, err
		}
		if err := waitForManagedCDP(paths, 5*time.Second); err != nil {
			_, _ = Stop(paths, 3*time.Second)
			_ = ClearState(paths)
			return Result{}, err
		}
	}

	return result, nil
}

func BuildArgs(binaryPath string, selected profile.Selection, url string) []string {
	args := []string{
		appModePrefix + url,
		"--no-first-run",
		"--disable-sync",
	}
	if binaryPath == "" {
		return args
	}
	if selected.UserDataDir != "" {
		args = append(args, "--remote-debugging-address="+defaultDevToolsHost)
		args = append(args, "--remote-debugging-port="+autoRemoteDebuggingPort)
		args = append(args, "--user-data-dir="+selected.UserDataDir)
	}
	return args
}

func (r Result) Render() string {
	mode := "launch"
	if r.DryRun {
		mode = "dry-run"
	}
	return fmt.Sprintf(
		"mode=%s\nbinary=%s\nruntime_source=%s\nprofile_mode=%s\nmanaged=%t\nstate_file=%s\nargs=%s\n",
		mode,
		r.BinaryPath,
		r.RuntimeSource,
		r.Mode,
		r.Managed,
		r.StateFile,
		strings.Join(r.Args, " "),
	)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return settings.DefaultWebAppURL
}

func buildManagedLaunchCommand(binaryPath string, args []string) *exec.Cmd {
	openArgs := []string{"-na", chrome.AppBundlePath(binaryPath), "--args"}
	openArgs = append(openArgs, args...)
	return exec.Command("open", openArgs...)
}

func buildSharedLaunchCommand(binaryPath string, args []string) *exec.Cmd {
	return exec.Command(binaryPath, args...)
}

func waitForManagedSession(paths macos.Paths, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		report, err := Status(paths)
		if err != nil {
			return err
		}
		if report.Alive {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("managed browser session did not become observable before timeout")
}

func waitForManagedCDP(paths macos.Paths, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		report, err := Status(paths)
		if err != nil {
			return err
		}
		if report.CDP.Status == cdpStatusOK {
			return nil
		}
		time.Sleep(cdpReadyPollInterval)
	}
	return errors.New("managed browser cdp endpoint did not become observable before timeout")
}
