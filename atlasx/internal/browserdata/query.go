package browserdata

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"atlasx/internal/imports"
	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
)

type BookmarkEntry struct {
	Root string
	Name string
	URL  string
}

type chromeBookmarks struct {
	Roots map[string]bookmarkNode `json:"roots"`
}

type bookmarkNode struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	URL      string         `json:"url"`
	Children []bookmarkNode `json:"children"`
}

func LoadHistory(paths macos.Paths) ([]mirror.HistoryEntry, error) {
	snapshot, err := mirror.Load(paths)
	if err != nil {
		return nil, err
	}
	if len(snapshot.HistoryRows) == 0 {
		return nil, errors.New("history mirror has no rows")
	}
	return snapshot.HistoryRows, nil
}

func LoadDownloads(paths macos.Paths) ([]mirror.DownloadEntry, error) {
	snapshot, err := mirror.Load(paths)
	if err != nil {
		return nil, err
	}
	if len(snapshot.DownloadRows) == 0 {
		return nil, errors.New("download mirror has no rows")
	}
	return snapshot.DownloadRows, nil
}

func LoadBookmarks(paths macos.Paths) ([]BookmarkEntry, error) {
	bookmarksPath := filepath.Join(imports.DefaultChromeImportRoot(paths), "Bookmarks.json")
	data, err := os.ReadFile(bookmarksPath)
	if err != nil {
		return nil, err
	}

	var payload chromeBookmarks
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	entries := make([]BookmarkEntry, 0)
	for rootName, root := range payload.Roots {
		entries = append(entries, flattenBookmarks(rootName, root)...)
	}
	if len(entries) == 0 {
		return nil, errors.New("imported bookmarks have no url entries")
	}
	return entries, nil
}

func flattenBookmarks(rootName string, root bookmarkNode) []BookmarkEntry {
	entries := make([]BookmarkEntry, 0)
	for _, child := range root.Children {
		entries = append(entries, flattenBookmarkNode(rootName, child)...)
	}
	return entries
}

func flattenBookmarkNode(rootName string, node bookmarkNode) []BookmarkEntry {
	switch node.Type {
	case "url":
		return []BookmarkEntry{{
			Root: rootName,
			Name: node.Name,
			URL:  node.URL,
		}}
	case "folder":
		entries := make([]BookmarkEntry, 0)
		for _, child := range node.Children {
			entries = append(entries, flattenBookmarkNode(rootName, child)...)
		}
		return entries
	default:
		return nil
	}
}
