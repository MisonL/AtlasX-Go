package imports

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestParseSafariBookmarksPlist(t *testing.T) {
	payload := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Children</key>
  <array>
    <dict>
      <key>Title</key><string>OpenAI</string>
      <key>URLString</key><string>https://openai.com</string>
      <key>WebBookmarkType</key><string>WebBookmarkTypeLeaf</string>
    </dict>
    <dict>
      <key>WebBookmarkType</key><string>WebBookmarkTypeList</string>
      <key>Children</key>
      <array>
        <dict>
          <key>Title</key><string>Docs</string>
          <key>URLString</key><string>https://platform.openai.com</string>
          <key>WebBookmarkType</key><string>WebBookmarkTypeLeaf</string>
        </dict>
      </array>
    </dict>
  </array>
</dict>
</plist>`)
	entries, err := parseSafariBookmarksPlist(payload)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("unexpected entry count: %d", len(entries))
	}
}

func TestSaveSafariReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "report.json")
	report := SafariReport{ImportRoot: "/tmp/safari", Bookmarks: []SafariBookmarkEntry{{Title: "OpenAI", URL: "https://openai.com"}}}
	if err := saveSafariReport(path, report); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	var loaded SafariReport
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(loaded.Bookmarks) != 1 {
		t.Fatalf("unexpected bookmarks count: %d", len(loaded.Bookmarks))
	}
}

func TestDefaultSafariPaths(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if DefaultSafariBookmarksPath(paths) == "" || DefaultSafariHistoryPath(paths) == "" {
		t.Fatal("expected safari paths")
	}
}
