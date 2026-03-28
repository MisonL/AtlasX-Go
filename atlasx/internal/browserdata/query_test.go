package browserdata

import (
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/imports"
	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
)

func TestLoadHistoryAndDownloads(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	snapshot := mirror.Snapshot{
		HistoryRows:  []mirror.HistoryEntry{{URL: "https://openai.com"}},
		DownloadRows: []mirror.DownloadEntry{{TargetPath: "/tmp/file.zip"}},
	}
	if err := mirror.Save(paths, snapshot); err != nil {
		t.Fatalf("save mirror failed: %v", err)
	}

	history, err := LoadHistory(paths)
	if err != nil {
		t.Fatalf("load history failed: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("unexpected history count: %d", len(history))
	}

	downloads, err := LoadDownloads(paths)
	if err != nil {
		t.Fatalf("load downloads failed: %v", err)
	}
	if len(downloads) != 1 {
		t.Fatalf("unexpected download count: %d", len(downloads))
	}
}

func TestLoadBookmarks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	importRoot := imports.DefaultChromeImportRoot(paths)
	if err := os.MkdirAll(importRoot, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	payload := `{"roots":{"bookmark_bar":{"children":[{"type":"url","name":"OpenAI","url":"https://openai.com"}]}}}`
	if err := os.WriteFile(filepath.Join(importRoot, "Bookmarks.json"), []byte(payload), 0o644); err != nil {
		t.Fatalf("write bookmarks failed: %v", err)
	}

	bookmarks, err := LoadBookmarks(paths)
	if err != nil {
		t.Fatalf("load bookmarks failed: %v", err)
	}
	if len(bookmarks) != 1 {
		t.Fatalf("unexpected bookmark count: %d", len(bookmarks))
	}
	if bookmarks[0].URL != "https://openai.com" {
		t.Fatalf("unexpected bookmark url: %s", bookmarks[0].URL)
	}
}
