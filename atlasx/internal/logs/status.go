package logs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"atlasx/internal/platform/macos"
)

const DefaultRecentFilesLimit = 10

type FileEntry struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	SizeBytes  int64  `json:"size_bytes"`
	ModifiedAt string `json:"modified_at"`
}

type Status struct {
	Root        string      `json:"root"`
	Present     bool        `json:"present"`
	FileCount   int         `json:"file_count"`
	TotalBytes  int64       `json:"total_bytes"`
	LatestFile  string      `json:"latest_file"`
	LatestAt    string      `json:"latest_at"`
	Returned    int         `json:"returned"`
	RecentFiles []FileEntry `json:"recent_files"`
}

type fileSnapshot struct {
	name       string
	path       string
	sizeBytes  int64
	modifiedAt string
	modUnix    int64
}

func LoadStatus(paths macos.Paths, limit int) (Status, error) {
	if limit < 0 {
		return Status{}, fmt.Errorf("limit must be >= 0")
	}

	status := Status{Root: paths.LogsRoot}
	info, err := os.Stat(paths.LogsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return status, nil
		}
		return Status{}, err
	}
	if !info.IsDir() {
		return Status{}, fmt.Errorf("logs root is not a directory: %s", paths.LogsRoot)
	}

	status.Present = true
	files, err := scanFiles(paths.LogsRoot)
	if err != nil {
		return Status{}, err
	}

	status.FileCount = len(files)
	for _, file := range files {
		status.TotalBytes += file.sizeBytes
	}

	if len(files) == 0 {
		return status, nil
	}

	sort.Slice(files, func(i int, j int) bool {
		if files[i].modUnix == files[j].modUnix {
			return files[i].path < files[j].path
		}
		return files[i].modUnix > files[j].modUnix
	})

	status.LatestFile = files[0].path
	status.LatestAt = files[0].modifiedAt

	selected := files
	if limit == 0 {
		limit = DefaultRecentFilesLimit
	}
	if limit < len(selected) {
		selected = selected[:limit]
	}

	status.Returned = len(selected)
	status.RecentFiles = make([]FileEntry, 0, len(selected))
	for _, file := range selected {
		status.RecentFiles = append(status.RecentFiles, FileEntry{
			Name:       file.name,
			Path:       file.path,
			SizeBytes:  file.sizeBytes,
			ModifiedAt: file.modifiedAt,
		})
	}
	return status, nil
}

func scanFiles(root string) ([]fileSnapshot, error) {
	files := make([]fileSnapshot, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		files = append(files, fileSnapshot{
			name:       entry.Name(),
			path:       path,
			sizeBytes:  info.Size(),
			modifiedAt: info.ModTime().UTC().Format("2006-01-02T15:04:05Z"),
			modUnix:    info.ModTime().UTC().Unix(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (s Status) Render() string {
	lines := []string{
		fmt.Sprintf("logs_root=%s", s.Root),
		fmt.Sprintf("present=%t", s.Present),
		fmt.Sprintf("file_count=%d", s.FileCount),
		fmt.Sprintf("total_bytes=%d", s.TotalBytes),
		fmt.Sprintf("latest_file=%s", s.LatestFile),
		fmt.Sprintf("latest_at=%s", s.LatestAt),
		fmt.Sprintf("returned=%d", s.Returned),
	}
	for index, file := range s.RecentFiles {
		lines = append(lines,
			fmt.Sprintf("file[%d].name=%s", index, file.Name),
			fmt.Sprintf("file[%d].path=%s", index, file.Path),
			fmt.Sprintf("file[%d].size_bytes=%d", index, file.SizeBytes),
			fmt.Sprintf("file[%d].modified_at=%s", index, file.ModifiedAt),
		)
	}
	return strings.Join(lines, "\n") + "\n"
}
