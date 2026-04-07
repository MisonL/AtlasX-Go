package browserdata

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atlasx/internal/imports"
	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
)

func TestResolveHistoryURL(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	snapshot := mirror.Snapshot{
		HistoryRows: []mirror.HistoryEntry{
			{URL: "https://example.com/history"},
		},
	}
	if err := mirror.Save(paths, snapshot); err != nil {
		t.Fatalf("save mirror failed: %v", err)
	}

	url, err := ResolveHistoryURL(paths, 0)
	if err != nil {
		t.Fatalf("resolve history url failed: %v", err)
	}
	if url != "https://example.com/history" {
		t.Fatalf("unexpected history url: %s", url)
	}
}

func TestResolveDownloadURLRejectsEmptySource(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	snapshot := mirror.Snapshot{
		DownloadRows: []mirror.DownloadEntry{
			{TargetPath: "/tmp/file.zip"},
		},
	}
	if err := mirror.Save(paths, snapshot); err != nil {
		t.Fatalf("save mirror failed: %v", err)
	}

	if _, err := ResolveDownloadURL(paths, 0); err == nil {
		t.Fatal("expected empty tab url failure")
	}
}

func TestResolveHistoryURLRejectsUnsupportedScheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	snapshot := mirror.Snapshot{
		HistoryRows: []mirror.HistoryEntry{
			{URL: "javascript:alert(1)"},
		},
	}
	if err := mirror.Save(paths, snapshot); err != nil {
		t.Fatalf("save mirror failed: %v", err)
	}

	if _, err := ResolveHistoryURL(paths, 0); err == nil {
		t.Fatal("expected unsupported scheme failure")
	} else if !strings.Contains(err.Error(), "unsupported url scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveDownloadURLRejectsUnsupportedScheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	snapshot := mirror.Snapshot{
		DownloadRows: []mirror.DownloadEntry{
			{TargetPath: "/tmp/file.zip", TabURL: "file:///tmp/file.zip"},
		},
	}
	if err := mirror.Save(paths, snapshot); err != nil {
		t.Fatalf("save mirror failed: %v", err)
	}

	if _, err := ResolveDownloadURL(paths, 0); err == nil {
		t.Fatal("expected unsupported scheme failure")
	} else if !strings.Contains(err.Error(), "unsupported url scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveBookmarkURLRejectsOutOfRange(t *testing.T) {
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

	if _, err := ResolveBookmarkURL(paths, 1); err == nil {
		t.Fatal("expected out of range failure")
	}
}

func TestResolveBookmarkURLRejectsUnsupportedScheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	importRoot := imports.DefaultChromeImportRoot(paths)
	if err := os.MkdirAll(importRoot, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	payload := `{"roots":{"bookmark_bar":{"children":[{"type":"url","name":"DevTools","url":"devtools://devtools/bundled/inspector.html"}]}}}`
	if err := os.WriteFile(filepath.Join(importRoot, "Bookmarks.json"), []byte(payload), 0o644); err != nil {
		t.Fatalf("write bookmarks failed: %v", err)
	}

	if _, err := ResolveBookmarkURL(paths, 0); err == nil {
		t.Fatal("expected unsupported scheme failure")
	} else if !strings.Contains(err.Error(), "unsupported url scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
}
