package managedruntime

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"

	"atlasx/internal/platform/macos"
)

var ErrInstallAlreadyRunning = fmt.Errorf("managed runtime install is already running")

var installRunning atomic.Bool

var stageManagedRuntime = StageLocal
var verifyManagedRuntime = Verify

type InstallOptions struct {
	HTTPClient *http.Client
}

type InstallReport struct {
	InstallPlanPath         string       `json:"install_plan_path"`
	ArchivePath             string       `json:"archive_path"`
	ArchivePartPath         string       `json:"archive_part_path"`
	DownloadedArchiveSHA256 string       `json:"downloaded_archive_sha256"`
	ExtractedBundlePath     string       `json:"extracted_bundle_path"`
	StagedBundlePath        string       `json:"staged_bundle_path"`
	BinaryPath              string       `json:"binary_path"`
	ManifestPath            string       `json:"manifest_path"`
	Version                 string       `json:"version"`
	Channel                 string       `json:"channel"`
	CurrentPhase            InstallPhase `json:"current_phase"`
	Verified                bool         `json:"verified"`
}

func (r InstallReport) Render() string {
	return fmt.Sprintf(
		"install_plan=%s\narchive=%s\narchive_part=%s\ndownloaded_archive_sha256=%s\nextracted_bundle=%s\nstaged_bundle=%s\nbinary=%s\nmanifest=%s\nversion=%s\nchannel=%s\ncurrent_phase=%s\nverified=%t\n",
		r.InstallPlanPath,
		r.ArchivePath,
		r.ArchivePartPath,
		r.DownloadedArchiveSHA256,
		r.ExtractedBundlePath,
		r.StagedBundlePath,
		r.BinaryPath,
		r.ManifestPath,
		r.Version,
		r.Channel,
		r.CurrentPhase,
		r.Verified,
	)
}

func Install(paths macos.Paths, opts InstallOptions) (InstallReport, error) {
	if !installRunning.CompareAndSwap(false, true) {
		return InstallReport{}, ErrInstallAlreadyRunning
	}
	defer installRunning.Store(false)

	plan, err := LoadInstallPlan(paths)
	if err != nil {
		return InstallReport{}, err
	}

	report := InstallReport{
		InstallPlanPath: paths.RuntimeInstallPlanFile,
		ArchivePath:     plan.ArchivePath,
		ArchivePartPath: plan.ArchivePath + ".part",
		ManifestPath:    paths.RuntimeManifestFile,
		Version:         plan.Version,
		Channel:         plan.Channel,
	}
	cleanup := installCleanup{archivePath: plan.ArchivePath, archivePartPath: plan.ArchivePath + ".part"}

	client := opts.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	var backup runtimeBackup
	fail := func(cause error) (InstallReport, error) {
		if cleanupErr := cleanup.removeArtifacts(); cleanupErr != nil {
			cause = fmt.Errorf("%v; cleanup failed: %w", cause, cleanupErr)
		}

		if backup.shouldRollback() {
			var rollbackErr error
			plan, rollbackErr = failAndRollbackInstallPlan(paths, plan, cause.Error(), func() error {
				return backup.restore(paths)
			})
			if rollbackErr != nil {
				cause = fmt.Errorf("%v; rollback failed: %w", cause, rollbackErr)
			}
		} else {
			failedPlan, advanceErr := AdvanceInstallPlan(plan, InstallEventFail, cause.Error())
			if advanceErr == nil {
				plan = failedPlan
				if saveErr := SaveInstallPlan(paths, plan); saveErr != nil {
					cause = fmt.Errorf("%v; save failed plan: %w", cause, saveErr)
				}
			} else if saveErr := SaveInstallPlan(paths, plan); saveErr != nil {
				cause = fmt.Errorf("%v; save install plan: %w", cause, saveErr)
			}
		}

		report.CurrentPhase = plan.CurrentPhase
		return report, cause
	}

	if plan, err = advanceAndSaveInstallPlan(paths, plan, InstallEventStartDownload); err != nil {
		return report, err
	}
	report.CurrentPhase = plan.CurrentPhase

	archiveSHA256, err := downloadInstallArchive(client, plan.SourceURL, plan.ArchivePath)
	if err != nil {
		return fail(err)
	}
	report.DownloadedArchiveSHA256 = archiveSHA256

	if plan, err = advanceAndSaveInstallPlan(paths, plan, InstallEventFinishDownload); err != nil {
		return fail(err)
	}
	report.CurrentPhase = plan.CurrentPhase

	if plan, err = advanceAndSaveInstallPlan(paths, plan, InstallEventStartVerify); err != nil {
		return fail(err)
	}
	report.CurrentPhase = plan.CurrentPhase

	if archiveSHA256 != plan.ExpectedSHA256 {
		return fail(fmt.Errorf(
			"managed runtime archive sha256 does not match install plan: expected=%s actual=%s",
			plan.ExpectedSHA256,
			archiveSHA256,
		))
	}

	if plan, err = advanceAndSaveInstallPlan(paths, plan, InstallEventFinishVerify); err != nil {
		return fail(err)
	}
	report.CurrentPhase = plan.CurrentPhase

	extractRoot, err := os.MkdirTemp(paths.RuntimeRoot, "install-extract-")
	if err != nil {
		return fail(err)
	}
	cleanup.extractRoot = extractRoot
	defer func() {
		_ = os.RemoveAll(extractRoot)
	}()

	extractedBundlePath, err := extractBundleArchive(plan.ArchivePath, extractRoot)
	if err != nil {
		return fail(err)
	}
	report.ExtractedBundlePath = extractedBundlePath
	if plan.BundleName != "" && filepath.Base(extractedBundlePath) != plan.BundleName {
		return fail(fmt.Errorf(
			"managed runtime archive bundle does not match install plan: expected=%s actual=%s",
			plan.BundleName,
			filepath.Base(extractedBundlePath),
		))
	}

	if _, err := ResolveBundleBinaryPath(extractedBundlePath); err != nil {
		return fail(err)
	}

	if plan, err = advanceAndSaveInstallPlan(paths, plan, InstallEventStartStage); err != nil {
		return fail(err)
	}
	report.CurrentPhase = plan.CurrentPhase

	if backup, err = backupExistingRuntime(paths); err != nil {
		return fail(err)
	}

	stageReport, err := stageManagedRuntime(paths, StageOptions{
		BundlePath: extractedBundlePath,
		Version:    plan.Version,
		Channel:    plan.Channel,
	})
	if err != nil {
		return fail(err)
	}
	report.StagedBundlePath = stageReport.StagedBundlePath
	report.BinaryPath = stageReport.BinaryPath

	verifyReport, err := verifyManagedRuntime(paths)
	if err != nil {
		return fail(err)
	}
	report.Verified = verifyReport.Verified

	if err := backup.cleanup(); err != nil {
		return fail(err)
	}

	if plan, err = advanceAndSaveInstallPlan(paths, plan, InstallEventFinishStage); err != nil {
		return fail(err)
	}
	report.CurrentPhase = plan.CurrentPhase
	return report, nil
}

func failAndRollbackInstallPlan(paths macos.Paths, plan InstallPlan, reason string, restore func() error) (InstallPlan, error) {
	failedPlan, err := AdvanceInstallPlan(plan, InstallEventFail, reason)
	if err != nil {
		return plan, err
	}
	plan = failedPlan
	if err := SaveInstallPlan(paths, plan); err != nil {
		return plan, err
	}

	rollbackPlan, err := AdvanceInstallPlan(plan, InstallEventStartRollback, "")
	if err != nil {
		return plan, err
	}
	plan = rollbackPlan
	if err := SaveInstallPlan(paths, plan); err != nil {
		return plan, err
	}

	if err := restore(); err != nil {
		return plan, err
	}

	rolledBackPlan, err := AdvanceInstallPlan(plan, InstallEventFinishRollback, "")
	if err != nil {
		return plan, err
	}
	rolledBackPlan.LastError = reason
	if err := SaveInstallPlan(paths, rolledBackPlan); err != nil {
		return plan, err
	}
	return rolledBackPlan, nil
}

func advanceAndSaveInstallPlan(paths macos.Paths, plan InstallPlan, event InstallEvent) (InstallPlan, error) {
	advanced, err := AdvanceInstallPlan(plan, event, "")
	if err != nil {
		return plan, err
	}
	if err := SaveInstallPlan(paths, advanced); err != nil {
		return plan, err
	}
	return advanced, nil
}

func downloadInstallArchive(client *http.Client, sourceURL string, archivePath string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		return "", err
	}

	archivePartPath := archivePath + ".part"
	if err := os.RemoveAll(archivePartPath); err != nil {
		return "", err
	}

	response, err := client.Get(sourceURL)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("managed runtime download returned status %d", response.StatusCode)
	}

	partFile, err := os.OpenFile(archivePartPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(partFile, hash), response.Body); err != nil {
		_ = partFile.Close()
		return "", err
	}
	if err := partFile.Close(); err != nil {
		return "", err
	}

	if err := os.Rename(archivePartPath, archivePath); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func extractBundleArchive(archivePath string, extractRoot string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = reader.Close()
	}()

	for _, file := range reader.File {
		targetPath := filepath.Join(extractRoot, filepath.Clean(file.Name))
		if !pathWithinRoot(extractRoot, targetPath) {
			return "", fmt.Errorf("archive entry escapes extraction root: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.Mode().Perm()); err != nil {
				return "", err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return "", err
		}
		if err := extractZipFile(file, targetPath); err != nil {
			return "", err
		}
	}

	return locateExtractedBundle(extractRoot)
}

func extractZipFile(file *zip.File, targetPath string) error {
	source, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = source.Close()
	}()

	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer func() {
		_ = target.Close()
	}()

	_, err = io.Copy(target, source)
	return err
}

func locateExtractedBundle(extractRoot string) (string, error) {
	bundles := make([]string, 0, 1)
	if err := filepath.WalkDir(extractRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && filepath.Ext(path) == ".app" {
			bundles = append(bundles, path)
			return filepath.SkipDir
		}
		return nil
	}); err != nil {
		return "", err
	}

	if len(bundles) == 0 {
		return "", fmt.Errorf("managed runtime archive does not contain a .app bundle")
	}
	if len(bundles) > 1 {
		return "", fmt.Errorf("managed runtime archive contains multiple bundles: %v", bundles)
	}
	return bundles[0], nil
}

func pathWithinRoot(root string, path string) bool {
	relativePath, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return relativePath != ".." && !filepath.IsAbs(relativePath) && relativePath != ""
}

type runtimeBackup struct {
	bundlePath string
	manifest   string
}

func (b runtimeBackup) shouldRollback() bool {
	return b.bundlePath != "" || b.manifest != ""
}

func backupExistingRuntime(paths macos.Paths) (runtimeBackup, error) {
	report, err := Status(paths)
	if err != nil {
		return runtimeBackup{}, err
	}

	backup := runtimeBackup{
		bundlePath: filepath.Join(paths.RuntimeRoot, ".backup-"+filepath.Base(report.StagedBundlePath)),
		manifest:   filepath.Join(paths.RuntimeRoot, ".backup-manifest.json"),
	}

	if report.BundlePresent {
		if err := os.RemoveAll(backup.bundlePath); err != nil {
			return runtimeBackup{}, err
		}
		if err := os.Rename(report.StagedBundlePath, backup.bundlePath); err != nil {
			return runtimeBackup{}, err
		}
	}
	if report.ManifestPresent {
		if err := os.Remove(backup.manifest); err != nil && !os.IsNotExist(err) {
			return runtimeBackup{}, err
		}
		if err := os.Rename(paths.RuntimeManifestFile, backup.manifest); err != nil {
			return runtimeBackup{}, err
		}
	}
	return backup, nil
}

func (b runtimeBackup) restore(paths macos.Paths) error {
	if !b.shouldRollback() {
		return nil
	}

	if err := os.RemoveAll(filepath.Join(paths.RuntimeRoot, stagedBundleName)); err != nil {
		return err
	}
	if err := os.Remove(paths.RuntimeManifestFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	if _, err := os.Stat(b.bundlePath); err == nil {
		if err := os.Rename(b.bundlePath, filepath.Join(paths.RuntimeRoot, stagedBundleName)); err != nil {
			return err
		}
	}
	if _, err := os.Stat(b.manifest); err == nil {
		if err := os.Rename(b.manifest, paths.RuntimeManifestFile); err != nil {
			return err
		}
	}
	return nil
}

func (b runtimeBackup) cleanup() error {
	if b.bundlePath != "" {
		if err := os.RemoveAll(b.bundlePath); err != nil {
			return err
		}
	}
	if b.manifest != "" {
		if err := os.Remove(b.manifest); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

type installCleanup struct {
	archivePath     string
	archivePartPath string
	extractRoot     string
}

func (c installCleanup) removeArtifacts() error {
	targets := []string{c.archivePartPath, c.archivePath, c.extractRoot}
	for _, target := range targets {
		if target == "" {
			continue
		}
		if err := os.RemoveAll(target); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return pruneEmptyParents(filepath.Dir(c.archivePath))
}

func pruneEmptyParents(path string) error {
	for path != "" && path != "." && path != string(filepath.Separator) {
		entries, err := os.ReadDir(path)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if len(entries) != 0 {
			return nil
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		path = filepath.Dir(path)
	}
	return nil
}
