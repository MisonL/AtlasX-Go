package diagnostics

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"atlasx/internal/platform/chrome"
	"atlasx/internal/platform/macos"
	"atlasx/internal/profile"
	"atlasx/internal/settings"
)

type Report struct {
	Paths          macos.Paths
	Config         settings.Config
	Chrome         chrome.Detection
	ChromeStatus   string
	IsolatedPath   string
	SharedModeName string
}

func Generate() (Report, error) {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return Report{}, err
	}

	cfg, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return Report{}, err
	}

	selection, err := profile.NewStore(paths.ProfilesRoot).Resolve(profile.ModeIsolated)
	if err != nil {
		return Report{}, err
	}

	detection, detectionErr := chrome.Detect(cfg.ChromeBinary)
	status := "ok"
	if detectionErr != nil {
		status = detectionErr.Error()
		if !errors.Is(detectionErr, chrome.ErrBinaryNotFound) {
			return Report{}, detectionErr
		}
	}

	return Report{
		Paths:          paths,
		Config:         cfg,
		Chrome:         detection,
		ChromeStatus:   status,
		IsolatedPath:   selection.UserDataDir,
		SharedModeName: profile.ModeShared,
	}, nil
}

func (r Report) Render() string {
	lines := []string{
		"AtlasX Doctor",
		fmt.Sprintf("goos=%s", runtime.GOOS),
		fmt.Sprintf("goarch=%s", runtime.GOARCH),
		fmt.Sprintf("support_root=%s", r.Paths.SupportRoot),
		fmt.Sprintf("config_file=%s", r.Paths.ConfigFile),
		fmt.Sprintf("isolated_profile=%s", r.IsolatedPath),
		fmt.Sprintf("shared_profile=%s", r.SharedModeName),
		fmt.Sprintf("chrome_status=%s", r.ChromeStatus),
		fmt.Sprintf("chrome_binary=%s", r.Chrome.BinaryPath),
	}
	if len(r.Chrome.Candidates) > 0 {
		lines = append(lines, "chrome_candidates="+strings.Join(r.Chrome.Candidates, ","))
	}
	return strings.Join(lines, "\n") + "\n"
}
