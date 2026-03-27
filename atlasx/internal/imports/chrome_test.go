package imports

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestImportChromeCopiesBookmarksAndPreferences(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	sourceProfileDir := filepath.Join(t.TempDir(), "Default")
	if err := os.MkdirAll(sourceProfileDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	bookmarksPayload := `{"roots":{"bookmark_bar":{"children":[{"type":"url","name":"OpenAI","url":"https://openai.com"}]}}}`
	if err := os.WriteFile(filepath.Join(sourceProfileDir, "Bookmarks"), []byte(bookmarksPayload), 0o644); err != nil {
		t.Fatalf("write bookmarks failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceProfileDir, "Preferences"), []byte(`{"browser":{"show_home_button":true}}`), 0o644); err != nil {
		t.Fatalf("write preferences failed: %v", err)
	}

	report, err := ImportChrome(paths, sourceProfileDir)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if !report.BookmarksImported.Exists {
		t.Fatal("expected imported bookmarks")
	}
	if !report.PreferencesImported.Exists {
		t.Fatal("expected imported preferences")
	}
}

func TestImportChromeFailsWithoutBookmarks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	sourceProfileDir := filepath.Join(t.TempDir(), "Default")
	if err := os.MkdirAll(sourceProfileDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	_, err = ImportChrome(paths, sourceProfileDir)
	if err == nil {
		t.Fatal("expected import failure")
	}
	if !strings.Contains(err.Error(), "bookmarks source is missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}
