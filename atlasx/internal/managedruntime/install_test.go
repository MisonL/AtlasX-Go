package managedruntime

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestInstallDownloadsArchiveStagesRuntimeAndAdvancesPlan(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	archiveBytes := createBundleArchiveBytes(t, createFakeChromiumBundle(t))
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveBytes)
	}))
	defer server.Close()

	plan, err := NewInstallPlan(InstallPlanOptions{
		Version:          "123.0.0",
		Channel:          "stable",
		SourceURL:        server.URL,
		ExpectedSHA256:   sha256Hex(archiveBytes),
		ArchivePath:      filepath.Join(t.TempDir(), "downloads", "chromium.zip"),
		StagedBundlePath: filepath.Join(paths.RuntimeRoot, stagedBundleName),
	})
	if err != nil {
		t.Fatalf("new install plan failed: %v", err)
	}
	if err := SaveInstallPlan(paths, plan); err != nil {
		t.Fatalf("save install plan failed: %v", err)
	}

	report, err := Install(paths, InstallOptions{HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if report.CurrentPhase != InstallPhaseStaged {
		t.Fatalf("unexpected install report: %+v", report)
	}

	if _, err := os.Stat(plan.ArchivePath); err != nil {
		t.Fatalf("expected downloaded archive: %v", err)
	}
	if _, err := os.Stat(plan.ArchivePath + ".part"); !os.IsNotExist(err) {
		t.Fatalf("expected archive part to be removed, got: %v", err)
	}

	status, err := InstallPlanInfo(paths)
	if err != nil {
		t.Fatalf("install plan info failed: %v", err)
	}
	if status.CurrentPhase != InstallPhaseStaged || status.LastError != "" {
		t.Fatalf("unexpected install plan status: %+v", status)
	}

	manifest, err := LoadManifest(paths)
	if err != nil {
		t.Fatalf("load manifest failed: %v", err)
	}
	if manifest.Version != "123.0.0" || manifest.Channel != "stable" {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}

	verifyReport, err := Verify(paths)
	if err != nil {
		t.Fatalf("verify installed runtime failed: %v", err)
	}
	if !verifyReport.Verified {
		t.Fatalf("expected verified runtime: %+v", verifyReport)
	}
}

func TestInstallFailsVerificationAndPreservesExistingRuntime(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if _, err := StageLocal(paths, StageOptions{
		BundlePath: createFakeChromiumBundle(t),
		Version:    "122.0.0",
		Channel:    "local",
	}); err != nil {
		t.Fatalf("stage existing runtime failed: %v", err)
	}
	previousManifest, err := LoadManifest(paths)
	if err != nil {
		t.Fatalf("load previous manifest failed: %v", err)
	}

	archiveBytes := createBundleArchiveBytes(t, createFakeChromiumBundle(t))
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(archiveBytes)
	}))
	defer server.Close()

	plan, err := NewInstallPlan(InstallPlanOptions{
		Version:          "123.0.0",
		Channel:          "stable",
		SourceURL:        server.URL,
		ExpectedSHA256:   "badc0ffee",
		ArchivePath:      filepath.Join(t.TempDir(), "downloads", "chromium.zip"),
		StagedBundlePath: filepath.Join(paths.RuntimeRoot, stagedBundleName),
	})
	if err != nil {
		t.Fatalf("new install plan failed: %v", err)
	}
	if err := SaveInstallPlan(paths, plan); err != nil {
		t.Fatalf("save install plan failed: %v", err)
	}

	if _, err := Install(paths, InstallOptions{HTTPClient: server.Client()}); err == nil {
		t.Fatal("expected install verification failure")
	}

	status, err := InstallPlanInfo(paths)
	if err != nil {
		t.Fatalf("install plan info failed: %v", err)
	}
	if status.CurrentPhase != InstallPhaseFailed || status.LastError == "" {
		t.Fatalf("unexpected failed install status: %+v", status)
	}

	currentManifest, err := LoadManifest(paths)
	if err != nil {
		t.Fatalf("load current manifest failed: %v", err)
	}
	if currentManifest.Version != previousManifest.Version || currentManifest.Channel != previousManifest.Channel {
		t.Fatalf("expected previous runtime to remain active: before=%+v after=%+v", previousManifest, currentManifest)
	}

	verifyReport, err := Verify(paths)
	if err != nil {
		t.Fatalf("verify existing runtime failed: %v", err)
	}
	if !verifyReport.Verified {
		t.Fatalf("expected previous runtime to remain verified: %+v", verifyReport)
	}
}

func createBundleArchiveBytes(t *testing.T, bundlePath string) []byte {
	t.Helper()

	archivePath := filepath.Join(t.TempDir(), "bundle.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create archive failed: %v", err)
	}

	zipWriter := zip.NewWriter(file)
	if err := filepath.Walk(bundlePath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relativePath, err := filepath.Rel(filepath.Dir(bundlePath), path)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relativePath)
		if info.IsDir() {
			header.Name += "/"
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		source, err := os.Open(path)
		if err != nil {
			return err
		}
		defer source.Close()

		_, err = io.Copy(writer, source)
		return err
	}); err != nil {
		t.Fatalf("write archive failed: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("close archive writer failed: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close archive file failed: %v", err)
	}

	data, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read archive failed: %v", err)
	}
	return data
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
