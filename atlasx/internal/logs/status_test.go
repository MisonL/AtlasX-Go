package logs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"atlasx/internal/platform/macos"
)

func TestLoadStatusReturnsAbsentWhenRootMissing(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	status, err := LoadStatus(paths, 5)
	if err != nil {
		t.Fatalf("load status failed: %v", err)
	}
	if status.Present {
		t.Fatalf("expected present=false: %+v", status)
	}
	if status.FileCount != 0 || status.TotalBytes != 0 || status.Returned != 0 {
		t.Fatalf("unexpected absent status: %+v", status)
	}
}

func TestLoadStatusReturnsRecentFiles(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(paths.LogsRoot, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir logs root failed: %v", err)
	}

	writeLogFile(t, filepath.Join(paths.LogsRoot, "atlas.log"), "atlas", time.Date(2026, 4, 8, 10, 0, 0, 0, time.UTC))
	writeLogFile(t, filepath.Join(paths.LogsRoot, "nested", "daemon.log"), "daemon-data", time.Date(2026, 4, 8, 11, 0, 0, 0, time.UTC))

	status, err := LoadStatus(paths, 1)
	if err != nil {
		t.Fatalf("load status failed: %v", err)
	}
	if !status.Present {
		t.Fatalf("expected present=true: %+v", status)
	}
	if status.FileCount != 2 {
		t.Fatalf("unexpected file count: %d", status.FileCount)
	}
	if status.TotalBytes != int64(len("atlas")+len("daemon-data")) {
		t.Fatalf("unexpected total bytes: %d", status.TotalBytes)
	}
	if status.LatestFile != filepath.Join(paths.LogsRoot, "nested", "daemon.log") {
		t.Fatalf("unexpected latest file: %s", status.LatestFile)
	}
	if status.Returned != 1 || len(status.RecentFiles) != 1 {
		t.Fatalf("unexpected returned files: %+v", status)
	}
	if status.RecentFiles[0].Name != "daemon.log" {
		t.Fatalf("unexpected first file: %+v", status.RecentFiles[0])
	}
}

func TestLoadStatusRejectsNegativeLimit(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	_, err = LoadStatus(paths, -1)
	if err == nil {
		t.Fatal("expected negative limit to fail")
	}
}

func writeLogFile(t *testing.T, path string, content string, modTime time.Time) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write log file failed: %v", err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("chtimes failed: %v", err)
	}
}
