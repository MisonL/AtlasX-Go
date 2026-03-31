package managedruntime

import (
	"errors"
	"fmt"

	"atlasx/internal/platform/macos"
)

var (
	ErrRuntimeManifestNotFound      = errors.New("managed runtime manifest is not present")
	ErrRuntimeManifestSHA256Missing = errors.New("managed runtime manifest sha256 is empty")
	ErrRuntimeBundleNotFound        = errors.New("managed runtime bundle is not present")
	ErrRuntimeBinaryNotFound        = errors.New("managed runtime binary is not present")
	ErrRuntimeBinaryNotExecutable   = errors.New("managed runtime binary is not executable")
	ErrRuntimeSHA256Mismatch        = errors.New("managed runtime binary sha256 does not match manifest")
)

type VerifyReport struct {
	RuntimeRoot      string `json:"runtime_root"`
	ManifestPath     string `json:"manifest_path"`
	ManifestPresent  bool   `json:"manifest_present"`
	ManifestVersion  string `json:"manifest_version"`
	ManifestChannel  string `json:"manifest_channel"`
	ManifestSHA256   string `json:"manifest_sha256"`
	StagedBundlePath string `json:"staged_bundle_path"`
	BundlePresent    bool   `json:"bundle_present"`
	BinaryPath       string `json:"binary_path"`
	BinaryPresent    bool   `json:"binary_present"`
	BinaryExecutable bool   `json:"binary_executable"`
	ActualSHA256     string `json:"actual_sha256"`
	Verified         bool   `json:"verified"`
}

func Verify(paths macos.Paths) (VerifyReport, error) {
	status, err := Status(paths)
	if err != nil {
		return VerifyReport{}, err
	}

	report := VerifyReport{
		RuntimeRoot:      status.RuntimeRoot,
		ManifestPath:     status.ManifestPath,
		ManifestPresent:  status.ManifestPresent,
		ManifestVersion:  status.ManifestVersion,
		ManifestChannel:  status.ManifestChannel,
		ManifestSHA256:   status.ManifestSHA256,
		StagedBundlePath: status.StagedBundlePath,
		BundlePresent:    status.BundlePresent,
		BinaryPath:       status.BinaryPath,
		BinaryPresent:    status.BinaryPresent,
		BinaryExecutable: status.BinaryExecutable,
	}

	if !report.ManifestPresent {
		return report, ErrRuntimeManifestNotFound
	}
	if !report.BundlePresent {
		return report, fmt.Errorf("%w: %s", ErrRuntimeBundleNotFound, report.StagedBundlePath)
	}
	if !report.BinaryPresent {
		return report, fmt.Errorf("%w: %s", ErrRuntimeBinaryNotFound, report.BinaryPath)
	}
	if !report.BinaryExecutable {
		return report, fmt.Errorf("%w: %s", ErrRuntimeBinaryNotExecutable, report.BinaryPath)
	}
	if report.ManifestSHA256 == "" {
		return report, ErrRuntimeManifestSHA256Missing
	}

	actualSHA256, err := fileSHA256(report.BinaryPath)
	if err != nil {
		return report, err
	}
	report.ActualSHA256 = actualSHA256
	if report.ActualSHA256 != report.ManifestSHA256 {
		return report, fmt.Errorf(
			"%w: manifest=%s actual=%s",
			ErrRuntimeSHA256Mismatch,
			report.ManifestSHA256,
			report.ActualSHA256,
		)
	}

	report.Verified = true
	return report, nil
}

func (r VerifyReport) Render() string {
	return fmt.Sprintf(
		"runtime_root=%s\nmanifest=%s\nmanifest_present=%t\nmanifest_version=%s\nmanifest_channel=%s\nmanifest_sha256=%s\nstaged_bundle=%s\nbundle_present=%t\nbinary=%s\nbinary_present=%t\nbinary_executable=%t\nactual_sha256=%s\nverified=%t\n",
		r.RuntimeRoot,
		r.ManifestPath,
		r.ManifestPresent,
		r.ManifestVersion,
		r.ManifestChannel,
		r.ManifestSHA256,
		r.StagedBundlePath,
		r.BundlePresent,
		r.BinaryPath,
		r.BinaryPresent,
		r.BinaryExecutable,
		r.ActualSHA256,
		r.Verified,
	)
}
