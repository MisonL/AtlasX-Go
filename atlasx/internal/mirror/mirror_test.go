package mirror

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestCollectMarksMissingArtifacts(t *testing.T) {
	snapshot, err := Collect(t.TempDir())
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}

	if snapshot.History.Status != statusMissing {
		t.Fatalf("unexpected history status: %s", snapshot.History.Status)
	}
	if snapshot.Downloads.Status != statusMissing {
		t.Fatalf("unexpected downloads status: %s", snapshot.Downloads.Status)
	}
	if snapshot.Bookmarks.Status != statusMissing {
		t.Fatalf("unexpected bookmarks status: %s", snapshot.Bookmarks.Status)
	}
}

func TestCollectParsesBookmarkRoots(t *testing.T) {
	profileDir := t.TempDir()
	bookmarksPath := filepath.Join(profileDir, "Bookmarks")
	payload := `{"roots":{"bookmark_bar":{"children":[{"type":"url","name":"OpenAI","url":"https://openai.com"},{"type":"folder","name":"Docs","children":[{"type":"url","name":"API","url":"https://platform.openai.com"}]}]},"other":{"children":[]}}}`
	if err := os.WriteFile(bookmarksPath, []byte(payload), 0o644); err != nil {
		t.Fatalf("write bookmarks failed: %v", err)
	}

	snapshot, err := Collect(profileDir)
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}

	if snapshot.Bookmarks.Status != statusScanned {
		t.Fatalf("unexpected bookmark status: %s", snapshot.Bookmarks.Status)
	}
	if len(snapshot.Bookmarks.RootSummaries) != 2 {
		t.Fatalf("unexpected root summary count: %d", len(snapshot.Bookmarks.RootSummaries))
	}
}

func TestSaveWritesMirrorFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	snapshot, err := Collect(t.TempDir())
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}
	if err := Save(paths, snapshot); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	data, err := os.ReadFile(paths.MirrorFile)
	if err != nil {
		t.Fatalf("read mirror file failed: %v", err)
	}
	if !strings.Contains(string(data), "\"history\"") {
		t.Fatalf("unexpected mirror payload: %s", string(data))
	}
}
