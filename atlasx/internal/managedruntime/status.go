package managedruntime

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"atlasx/internal/platform/macos"
)

var ErrStagedRuntimeNotFound = errors.New("managed runtime is not staged")

type StatusReport struct {
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
}

func Status(paths macos.Paths) (StatusReport, error) {
	report := StatusReport{
		RuntimeRoot:      paths.RuntimeRoot,
		ManifestPath:     paths.RuntimeManifestFile,
		StagedBundlePath: filepath.Join(paths.RuntimeRoot, stagedBundleName),
	}

	manifest, err := LoadManifest(paths)
	if err != nil {
		if !os.IsNotExist(err) {
			return StatusReport{}, err
		}
	} else {
		report.ManifestPresent = true
		report.ManifestVersion = manifest.Version
		report.ManifestChannel = manifest.Channel
		report.ManifestSHA256 = manifest.SHA256
		if manifest.BundlePath != "" {
			report.StagedBundlePath = manifest.BundlePath
		}
		if manifest.BinaryPath != "" {
			report.BinaryPath = manifest.BinaryPath
		}
	}

	if info, err := os.Stat(report.StagedBundlePath); err == nil && info.IsDir() {
		report.BundlePresent = true
	} else if err != nil && !os.IsNotExist(err) {
		return StatusReport{}, err
	}

	if report.BinaryPath == "" {
		report.BinaryPath = filepath.Join(report.StagedBundlePath, "Contents", "MacOS", chromiumBinaryName)
	}
	if info, err := os.Stat(report.BinaryPath); err == nil && !info.IsDir() {
		report.BinaryPresent = true
		report.BinaryExecutable = info.Mode()&0o111 != 0
	} else if err != nil && !os.IsNotExist(err) {
		return StatusReport{}, err
	}

	return report, nil
}

func Clear(paths macos.Paths) error {
	report, err := Status(paths)
	if err != nil {
		return err
	}
	if !report.ManifestPresent && !report.BundlePresent {
		return ErrStagedRuntimeNotFound
	}

	if err := os.RemoveAll(report.StagedBundlePath); err != nil {
		return err
	}
	if err := os.Remove(paths.RuntimeManifestFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (r StatusReport) Render() string {
	return fmt.Sprintf(
		"runtime_root=%s\nmanifest=%s\nmanifest_present=%t\nmanifest_version=%s\nmanifest_channel=%s\nmanifest_sha256=%s\nstaged_bundle=%s\nbundle_present=%t\nbinary=%s\nbinary_present=%t\nbinary_executable=%t\n",
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
	)
}
