package launcher

import (
	"fmt"
	"os/exec"
	"strings"

	"atlasx/internal/platform/chrome"
	"atlasx/internal/platform/macos"
	"atlasx/internal/profile"
	"atlasx/internal/settings"
)

const appModePrefix = "--app="

type Options struct {
	DryRun        bool
	UseSharedMode bool
	URLOverride   string
}

type Result struct {
	BinaryPath string
	Args       []string
	DryRun     bool
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

	args := BuildArgs(detection.BinaryPath, selected, firstNonEmpty(opts.URLOverride, cfg.WebAppURL))
	result := Result{BinaryPath: detection.BinaryPath, Args: args, DryRun: opts.DryRun}
	if opts.DryRun {
		return result, nil
	}

	cmd := exec.Command(detection.BinaryPath, args...)
	return result, cmd.Start()
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
		args = append(args, "--user-data-dir="+selected.UserDataDir)
	}
	return args
}

func (r Result) Render() string {
	mode := "launch"
	if r.DryRun {
		mode = "dry-run"
	}
	return fmt.Sprintf("mode=%s\nbinary=%s\nargs=%s\n", mode, r.BinaryPath, strings.Join(r.Args, " "))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return settings.DefaultWebAppURL
}
