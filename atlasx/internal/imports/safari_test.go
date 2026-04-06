package imports

import (
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestImportSafariWritesSuccessStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	bookmarksPath := DefaultSafariBookmarksPath(paths)
	if err := os.MkdirAll(filepath.Dir(bookmarksPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	payload := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Children</key>
  <array>
    <dict>
      <key>WebBookmarkType</key><string>WebBookmarkTypeLeaf</string>
      <key>Title</key><string>OpenAI</string>
      <key>URLString</key><string>https://openai.com</string>
    </dict>
  </array>
</dict>
</plist>`)
	if err := os.WriteFile(bookmarksPath, payload, 0o644); err != nil {
		t.Fatalf("write safari bookmarks failed: %v", err)
	}

	if _, err := ImportSafari(paths); err != nil {
		t.Fatalf("import safari failed: %v", err)
	}

	status, err := LoadSafariImportStatus(paths)
	if err != nil {
		t.Fatalf("load safari import status failed: %v", err)
	}
	if status.Result != importResultSucceeded || status.Source != bookmarksPath {
		t.Fatalf("unexpected safari import status: %+v", status)
	}
}

func TestImportSafariWritesFailureStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if _, err := ImportSafari(paths); err == nil {
		t.Fatal("expected safari import failure")
	}

	status, err := LoadSafariImportStatus(paths)
	if err != nil {
		t.Fatalf("load safari import status failed: %v", err)
	}
	if status.Result != importResultFailed || status.Error == "" {
		t.Fatalf("unexpected safari import status: %+v", status)
	}
}
