package managedruntime

import (
	"encoding/json"
	"os"

	"atlasx/internal/platform/macos"
)

type Manifest struct {
	Version     string `json:"version"`
	Channel     string `json:"channel"`
	SHA256      string `json:"sha256"`
	BundlePath  string `json:"bundle_path"`
	BinaryPath  string `json:"binary_path"`
	InstalledAt string `json:"installed_at"`
}

type ManifestStatus struct {
	Path       string `json:"path"`
	Present    bool   `json:"present"`
	Version    string `json:"version"`
	Channel    string `json:"channel"`
	BundlePath string `json:"bundle_path"`
	BinaryPath string `json:"binary_path"`
}

func LoadManifest(paths macos.Paths) (Manifest, error) {
	data, err := os.ReadFile(paths.RuntimeManifestFile)
	if err != nil {
		return Manifest{}, err
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func SaveManifest(paths macos.Paths, manifest Manifest) error {
	if err := macos.EnsureDir(paths.RuntimeRoot); err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(paths.RuntimeManifestFile, append(data, '\n'), 0o644)
}

func ManifestInfo(paths macos.Paths) (ManifestStatus, error) {
	status := ManifestStatus{Path: paths.RuntimeManifestFile}
	manifest, err := LoadManifest(paths)
	if err != nil {
		if os.IsNotExist(err) {
			return status, nil
		}
		return ManifestStatus{}, err
	}

	status.Present = true
	status.Version = manifest.Version
	status.Channel = manifest.Channel
	status.BundlePath = manifest.BundlePath
	status.BinaryPath = manifest.BinaryPath
	return status, nil
}
