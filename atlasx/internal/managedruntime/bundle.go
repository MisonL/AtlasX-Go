package managedruntime

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"atlasx/internal/platform/macos"
)

func ResolveBundleBinaryPath(bundlePath string) (string, error) {
	macosDir := filepath.Join(bundlePath, "Contents", "MacOS")
	entries, err := os.ReadDir(macosDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("bundle missing %s", macosDir)
		}
		return "", err
	}

	executables := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(macosDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return "", err
		}
		if info.Mode()&0o111 == 0 {
			continue
		}
		executables = append(executables, path)
	}

	if len(executables) == 0 {
		return "", fmt.Errorf("bundle has no executable in %s", macosDir)
	}
	if len(executables) > 1 {
		sort.Strings(executables)
		return "", fmt.Errorf("bundle has multiple executables in %s: %v", macosDir, executables)
	}
	return executables[0], nil
}

func DetectManagedBinaryPath(paths macos.Paths) (string, error) {
	manifest, err := LoadManifest(paths)
	if err == nil {
		if manifest.BinaryPath != "" {
			return manifest.BinaryPath, nil
		}
		if manifest.BundlePath != "" {
			return ResolveBundleBinaryPath(manifest.BundlePath)
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	legacyBinaryPath := filepath.Join(paths.RuntimeRoot, stagedBundleName, "Contents", "MacOS", chromiumBinaryName)
	if _, err := os.Stat(legacyBinaryPath); err == nil {
		return legacyBinaryPath, nil
	}
	return "", nil
}
