package managedruntime

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"atlasx/internal/platform/macos"
)

const (
	stagedBundleName   = "Chromium.app"
	chromiumBinaryName = "Chromium"
)

type StageOptions struct {
	BundlePath string
	Version    string
	Channel    string
}

type StageReport struct {
	SourceBundlePath string `json:"source_bundle_path"`
	StagedBundlePath string `json:"staged_bundle_path"`
	BinaryPath       string `json:"binary_path"`
	ManifestPath     string `json:"manifest_path"`
	Version          string `json:"version"`
	Channel          string `json:"channel"`
	SHA256           string `json:"sha256"`
}

func StageLocal(paths macos.Paths, opts StageOptions) (StageReport, error) {
	if opts.BundlePath == "" {
		return StageReport{}, fmt.Errorf("bundle_path is required")
	}
	if opts.Version == "" {
		return StageReport{}, fmt.Errorf("version is required")
	}
	if opts.Channel == "" {
		opts.Channel = "local"
	}

	sourceBundlePath, err := filepath.Abs(opts.BundlePath)
	if err != nil {
		return StageReport{}, err
	}
	sourceBinaryPath, err := ResolveBundleBinaryPath(sourceBundlePath)
	if err != nil {
		return StageReport{}, err
	}

	if err := macos.EnsureDir(paths.RuntimeRoot); err != nil {
		return StageReport{}, err
	}

	stagedBundlePath := filepath.Join(paths.RuntimeRoot, stagedBundleName)
	if err := os.RemoveAll(stagedBundlePath); err != nil {
		return StageReport{}, err
	}
	if err := copyTree(sourceBundlePath, stagedBundlePath); err != nil {
		return StageReport{}, err
	}

	stagedBinaryPath := filepath.Join(stagedBundlePath, "Contents", "MacOS", filepath.Base(sourceBinaryPath))
	sha256sum, err := fileSHA256(stagedBinaryPath)
	if err != nil {
		return StageReport{}, err
	}

	if err := SaveManifest(paths, Manifest{
		Version:     opts.Version,
		Channel:     opts.Channel,
		SHA256:      sha256sum,
		BundlePath:  stagedBundlePath,
		BinaryPath:  stagedBinaryPath,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		return StageReport{}, err
	}

	return StageReport{
		SourceBundlePath: sourceBundlePath,
		StagedBundlePath: stagedBundlePath,
		BinaryPath:       stagedBinaryPath,
		ManifestPath:     paths.RuntimeManifestFile,
		Version:          opts.Version,
		Channel:          opts.Channel,
		SHA256:           sha256sum,
	}, nil
}

func (r StageReport) Render() string {
	return fmt.Sprintf(
		"source_bundle=%s\nstaged_bundle=%s\nbinary=%s\nmanifest=%s\nversion=%s\nchannel=%s\nsha256=%s\n",
		r.SourceBundlePath,
		r.StagedBundlePath,
		r.BinaryPath,
		r.ManifestPath,
		r.Version,
		r.Channel,
		r.SHA256,
	)
}

func copyTree(sourceRoot string, targetRoot string) error {
	return filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetRoot, relativePath)

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return os.MkdirAll(targetPath, info.Mode().Perm())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(linkTarget, targetPath)
		}
		return copyFile(path, targetPath, info.Mode().Perm())
	})
}

func copyFile(sourcePath string, targetPath string, mode fs.FileMode) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() {
		_ = targetFile.Close()
	}()

	_, err = io.Copy(targetFile, sourceFile)
	return err
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
