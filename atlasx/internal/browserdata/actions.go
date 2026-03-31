package browserdata

import (
	"fmt"

	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
)

func ResolveHistoryURL(paths macos.Paths, index int) (string, error) {
	rows, err := LoadHistory(paths)
	if err != nil {
		return "", err
	}
	return resolveIndexedURL(rows, index, func(row mirror.HistoryEntry) string {
		return row.URL
	}, "history")
}

func ResolveDownloadURL(paths macos.Paths, index int) (string, error) {
	rows, err := LoadDownloads(paths)
	if err != nil {
		return "", err
	}
	url, err := resolveIndexedURL(rows, index, func(row mirror.DownloadEntry) string {
		return row.TabURL
	}, "download")
	if err != nil {
		return "", err
	}
	if url == "" {
		return "", fmt.Errorf("download index %d has empty tab url", index)
	}
	return url, nil
}

func ResolveBookmarkURL(paths macos.Paths, index int) (string, error) {
	rows, err := LoadBookmarks(paths)
	if err != nil {
		return "", err
	}
	return resolveIndexedURL(rows, index, func(row BookmarkEntry) string {
		return row.URL
	}, "bookmark")
}

func resolveIndexedURL[T any](rows []T, index int, getter func(T) string, label string) (string, error) {
	if index < 0 {
		return "", fmt.Errorf("index must be >= 0")
	}
	if index >= len(rows) {
		return "", fmt.Errorf("%s index %d out of range", label, index)
	}
	return getter(rows[index]), nil
}
