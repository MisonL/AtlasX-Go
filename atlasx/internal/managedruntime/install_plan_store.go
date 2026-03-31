package managedruntime

import (
	"encoding/json"
	"os"

	"atlasx/internal/platform/macos"
)

type InstallPlanStatus struct {
	Path             string       `json:"path"`
	Present          bool         `json:"present"`
	Version          string       `json:"version"`
	Channel          string       `json:"channel"`
	SourceURL        string       `json:"source_url"`
	ExpectedSHA256   string       `json:"expected_sha256"`
	ArchivePath      string       `json:"archive_path"`
	StagedBundlePath string       `json:"staged_bundle_path"`
	CurrentPhase     InstallPhase `json:"current_phase"`
	LastError        string       `json:"last_error"`
}

func LoadInstallPlan(paths macos.Paths) (InstallPlan, error) {
	data, err := os.ReadFile(paths.RuntimeInstallPlanFile)
	if err != nil {
		return InstallPlan{}, err
	}

	var plan InstallPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return InstallPlan{}, err
	}
	return plan, nil
}

func SaveInstallPlan(paths macos.Paths, plan InstallPlan) error {
	if err := macos.EnsureDir(paths.RuntimeRoot); err != nil {
		return err
	}

	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(paths.RuntimeInstallPlanFile, append(data, '\n'), 0o644)
}

func InstallPlanInfo(paths macos.Paths) (InstallPlanStatus, error) {
	status := InstallPlanStatus{Path: paths.RuntimeInstallPlanFile}

	plan, err := LoadInstallPlan(paths)
	if err != nil {
		if os.IsNotExist(err) {
			return status, nil
		}
		return InstallPlanStatus{}, err
	}

	status.Present = true
	status.Version = plan.Version
	status.Channel = plan.Channel
	status.SourceURL = plan.SourceURL
	status.ExpectedSHA256 = plan.ExpectedSHA256
	status.ArchivePath = plan.ArchivePath
	status.StagedBundlePath = plan.StagedBundlePath
	status.CurrentPhase = plan.CurrentPhase
	status.LastError = plan.LastError
	return status, nil
}
