package imports

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"atlasx/internal/platform/macos"
	"howett.net/plist"
)

type SafariReport struct {
	SourceBookmarks   FileArtifact          `json:"source_bookmarks"`
	ImportedBookmarks FileArtifact          `json:"imported_bookmarks"`
	SourceHistory     FileArtifact          `json:"source_history"`
	ImportRoot        string                `json:"import_root"`
	GeneratedAt       string                `json:"generated_at"`
	Bookmarks         []SafariBookmarkEntry `json:"bookmarks,omitempty"`
}

type SafariBookmarkEntry struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type safariBookmarksPayload struct {
	Children []safariBookmarkNode `json:"Children"`
}

type safariBookmarkNode struct {
	Title           string               `json:"Title"`
	URLString       string               `json:"URLString"`
	WebBookmarkType string               `json:"WebBookmarkType"`
	Children        []safariBookmarkNode `json:"Children"`
}

func DefaultSafariBookmarksPath(paths macos.Paths) string {
	return filepath.Join(paths.Home, "Library", "Safari", "Bookmarks.plist")
}

func DefaultSafariHistoryPath(paths macos.Paths) string {
	return filepath.Join(paths.Home, "Library", "Safari", "History.db")
}

func DefaultSafariImportRoot(paths macos.Paths) string {
	return filepath.Join(paths.ImportsRoot, "safari")
}

func ImportSafari(paths macos.Paths) (SafariReport, error) {
	bookmarksSourcePath := DefaultSafariBookmarksPath(paths)
	historySourcePath := DefaultSafariHistoryPath(paths)

	bookmarksSource, err := fileArtifact(bookmarksSourcePath)
	if err != nil {
		return SafariReport{}, err
	}
	if !bookmarksSource.Exists {
		return SafariReport{}, errors.New("safari bookmarks source is missing")
	}

	importRoot := DefaultSafariImportRoot(paths)
	if err := macos.EnsureDir(importRoot); err != nil {
		return SafariReport{}, err
	}

	bookmarksDest := filepath.Join(importRoot, "Bookmarks.json")
	reportPath := filepath.Join(importRoot, "report.json")

	entries, rawJSON, err := exportSafariBookmarks(bookmarksSourcePath)
	if err != nil {
		return SafariReport{}, err
	}
	if err := os.WriteFile(bookmarksDest, append(rawJSON, '\n'), 0o644); err != nil {
		return SafariReport{}, err
	}

	importedBookmarks, err := fileArtifact(bookmarksDest)
	if err != nil {
		return SafariReport{}, err
	}
	historySource, err := fileArtifact(historySourcePath)
	if err != nil {
		return SafariReport{}, err
	}

	report := SafariReport{
		SourceBookmarks:   bookmarksSource,
		ImportedBookmarks: importedBookmarks,
		SourceHistory:     historySource,
		ImportRoot:        importRoot,
		GeneratedAt:       time.Now().UTC().Format(time.RFC3339),
		Bookmarks:         entries,
	}

	if err := saveSafariReport(reportPath, report); err != nil {
		return SafariReport{}, err
	}
	return report, nil
}

func (r SafariReport) Render() string {
	return fmt.Sprintf(
		"import_root=%s\nsafari_bookmarks_source=%s\nsafari_bookmarks_imported=%s\nsafari_history_source=%s\nsafari_bookmarks_count=%d\n",
		r.ImportRoot,
		r.SourceBookmarks.Path,
		r.ImportedBookmarks.Path,
		r.SourceHistory.Path,
		len(r.Bookmarks),
	)
}

func exportSafariBookmarks(sourcePath string) ([]SafariBookmarkEntry, []byte, error) {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, nil, err
	}

	entries, err := parseSafariBookmarksPlist(data)
	if err != nil {
		return nil, nil, err
	}
	output, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return entries, output, nil
}

func parseSafariBookmarksPlist(data []byte) ([]SafariBookmarkEntry, error) {
	var payload safariBookmarksPayload
	if _, err := plist.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	entries := make([]SafariBookmarkEntry, 0)
	for _, child := range payload.Children {
		entries = append(entries, flattenSafariBookmarks(child)...)
	}
	return entries, nil
}

func flattenSafariBookmarks(node safariBookmarkNode) []SafariBookmarkEntry {
	switch node.WebBookmarkType {
	case "WebBookmarkTypeLeaf":
		if node.URLString == "" {
			return nil
		}
		return []SafariBookmarkEntry{{Title: node.Title, URL: node.URLString}}
	default:
		entries := make([]SafariBookmarkEntry, 0)
		for _, child := range node.Children {
			entries = append(entries, flattenSafariBookmarks(child)...)
		}
		return entries
	}
}

func saveSafariReport(path string, report SafariReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
