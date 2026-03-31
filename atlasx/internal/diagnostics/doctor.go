package diagnostics

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"atlasx/internal/launcher"
	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/chrome"
	"atlasx/internal/platform/macos"
	"atlasx/internal/profile"
	"atlasx/internal/settings"
)

type Report struct {
	Paths           macos.Paths
	Config          settings.Config
	Chrome          chrome.Detection
	ChromeStatus    string
	RuntimeManifest managedruntime.ManifestStatus
	IsolatedPath    string
	SharedModeName  string
	Session         launcher.StatusReport
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

	manifestStatus, err := managedruntime.ManifestInfo(paths)
	if err != nil {
		return Report{}, err
	}

	return Report{
		Paths:           paths,
		Config:          cfg,
		Chrome:          detection,
		ChromeStatus:    status,
		RuntimeManifest: manifestStatus,
		IsolatedPath:    selection.UserDataDir,
		SharedModeName:  profile.ModeShared,
		Session:         sessionReport(paths),
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
		fmt.Sprintf("chrome_source=%s", r.Chrome.Source),
		fmt.Sprintf("chrome_binary=%s", r.Chrome.BinaryPath),
		fmt.Sprintf("managed_runtime_manifest_path=%s", r.RuntimeManifest.Path),
		fmt.Sprintf("managed_runtime_manifest_present=%t", r.RuntimeManifest.Present),
		fmt.Sprintf("managed_runtime_manifest_version=%s", r.RuntimeManifest.Version),
		fmt.Sprintf("managed_runtime_manifest_channel=%s", r.RuntimeManifest.Channel),
		fmt.Sprintf("managed_runtime_manifest_sha256=%s", r.RuntimeManifest.SHA256),
		fmt.Sprintf("managed_runtime_manifest_bundle=%s", r.RuntimeManifest.BundlePath),
		fmt.Sprintf("managed_runtime_manifest_binary=%s", r.RuntimeManifest.BinaryPath),
		fmt.Sprintf("managed_session_present=%t", r.Session.Present),
		fmt.Sprintf("managed_session_alive=%t", r.Session.Alive),
		fmt.Sprintf("managed_session_state_file=%s", r.Session.StateFile),
		fmt.Sprintf("managed_session_cdp_status=%s", r.Session.CDP.Status),
		fmt.Sprintf("managed_session_cdp_version_endpoint=%s", r.Session.CDP.VersionEndpoint),
		fmt.Sprintf("managed_session_cdp_browser_ws=%s", r.Session.CDP.BrowserWebSocketURL),
	}
	if len(r.Chrome.Candidates) > 0 {
		lines = append(lines, "chrome_candidates="+strings.Join(r.Chrome.Candidates, ","))
	}
	return strings.Join(lines, "\n") + "\n"
}

func sessionReport(paths macos.Paths) launcher.StatusReport {
	report, err := launcher.Status(paths)
	if err != nil {
		return launcher.StatusReport{StateFile: paths.SessionFile}
	}
	return report
}
