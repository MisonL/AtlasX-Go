package updates

import (
	"fmt"
	"strings"

	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/macos"
)

type Status struct {
	RuntimeRoot      string                      `json:"runtime_root"`
	ManifestPresent  bool                        `json:"manifest_present"`
	StagedVersion    string                      `json:"staged_version"`
	StagedChannel    string                      `json:"staged_channel"`
	StagedBundlePath string                      `json:"staged_bundle_path"`
	StagedBinaryPath string                      `json:"staged_binary_path"`
	StagedReady      bool                        `json:"staged_ready"`
	PlanPresent      bool                        `json:"plan_present"`
	PlanVersion      string                      `json:"plan_version"`
	PlanChannel      string                      `json:"plan_channel"`
	PlanBundleName   string                      `json:"plan_bundle_name"`
	PlanSourceURL    string                      `json:"plan_source_url"`
	PlanArchivePath  string                      `json:"plan_archive_path"`
	PlanPhase        managedruntime.InstallPhase `json:"plan_phase"`
	PlanLastError    string                      `json:"plan_last_error"`
	PlanPending      bool                        `json:"plan_pending"`
	PlanInFlight     bool                        `json:"plan_in_flight"`
}

func LoadStatus(paths macos.Paths) (Status, error) {
	report, err := managedruntime.Status(paths)
	if err != nil {
		return Status{}, err
	}

	status := Status{
		RuntimeRoot:      report.RuntimeRoot,
		ManifestPresent:  report.ManifestPresent,
		StagedVersion:    report.ManifestVersion,
		StagedChannel:    report.ManifestChannel,
		StagedBundlePath: report.StagedBundlePath,
		StagedBinaryPath: report.BinaryPath,
		PlanPresent:      report.InstallPlanPresent,
		PlanVersion:      report.InstallPlanVersion,
		PlanChannel:      report.InstallPlanChannel,
		PlanBundleName:   report.InstallPlanBundleName,
		PlanSourceURL:    report.InstallPlanSourceURL,
		PlanArchivePath:  report.InstallPlanArchivePath,
		PlanPhase:        report.InstallPlanPhase,
		PlanLastError:    report.InstallPlanLastError,
	}
	status.StagedReady = report.ManifestPresent && report.BundlePresent && report.BinaryPresent && report.BinaryExecutable
	status.PlanPending = computePlanPending(report.InstallPlanPresent, report.InstallPlanPhase)
	status.PlanInFlight = computePlanInFlight(report.InstallPlanPresent, report.InstallPlanPhase)
	return status, nil
}

func computePlanPending(present bool, phase managedruntime.InstallPhase) bool {
	if !present {
		return false
	}
	switch phase {
	case "",
		managedruntime.InstallPhaseStaged,
		managedruntime.InstallPhaseRolledBack:
		return false
	default:
		return true
	}
}

func computePlanInFlight(present bool, phase managedruntime.InstallPhase) bool {
	if !present {
		return false
	}
	switch phase {
	case managedruntime.InstallPhaseDownloading,
		managedruntime.InstallPhaseDownloaded,
		managedruntime.InstallPhaseVerifying,
		managedruntime.InstallPhaseVerified,
		managedruntime.InstallPhaseStaging,
		managedruntime.InstallPhaseRollback:
		return true
	default:
		return false
	}
}

func (s Status) Render() string {
	return strings.Join([]string{
		fmt.Sprintf("runtime_root=%s", s.RuntimeRoot),
		fmt.Sprintf("manifest_present=%t", s.ManifestPresent),
		fmt.Sprintf("staged_version=%s", s.StagedVersion),
		fmt.Sprintf("staged_channel=%s", s.StagedChannel),
		fmt.Sprintf("staged_bundle_path=%s", s.StagedBundlePath),
		fmt.Sprintf("staged_binary_path=%s", s.StagedBinaryPath),
		fmt.Sprintf("staged_ready=%t", s.StagedReady),
		fmt.Sprintf("plan_present=%t", s.PlanPresent),
		fmt.Sprintf("plan_version=%s", s.PlanVersion),
		fmt.Sprintf("plan_channel=%s", s.PlanChannel),
		fmt.Sprintf("plan_bundle_name=%s", s.PlanBundleName),
		fmt.Sprintf("plan_source_url=%s", s.PlanSourceURL),
		fmt.Sprintf("plan_archive_path=%s", s.PlanArchivePath),
		fmt.Sprintf("plan_phase=%s", s.PlanPhase),
		fmt.Sprintf("plan_last_error=%s", s.PlanLastError),
		fmt.Sprintf("plan_pending=%t", s.PlanPending),
		fmt.Sprintf("plan_in_flight=%t", s.PlanInFlight),
	}, "\n") + "\n"
}
