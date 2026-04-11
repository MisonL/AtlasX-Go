package mirror

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"atlasx/internal/platform/macos"
)

const (
	formatJSON        = "json"
	formatSQLite      = "sqlite"
	statusPending     = "pending"
	statusMissing     = "missing"
	statusScanned     = "scanned"
	defaultProfileDir = "webapp-isolated"
)

type Snapshot struct {
	GeneratedAt  string           `json:"generated_at"`
	ProfileDir   string           `json:"profile_dir"`
	History      Artifact         `json:"history"`
	HistoryRows  []HistoryEntry   `json:"history_rows,omitempty"`
	Bookmarks    BookmarkArtifact `json:"bookmarks"`
	Downloads    Artifact         `json:"downloads"`
	DownloadRows []DownloadEntry  `json:"download_rows,omitempty"`
}

type Artifact struct {
	Kind       string `json:"kind"`
	Format     string `json:"format"`
	SourcePath string `json:"source_path"`
	Exists     bool   `json:"exists"`
	SizeBytes  int64  `json:"size_bytes"`
	ModifiedAt string `json:"modified_at,omitempty"`
	Status     string `json:"status"`
	Notes      string `json:"notes"`
}

type BookmarkArtifact struct {
	Artifact
	RootSummaries []BookmarkRootSummary `json:"root_summaries,omitempty"`
}

type BookmarkRootSummary struct {
	Name        string `json:"name"`
	FolderCount int    `json:"folder_count"`
	URLCount    int    `json:"url_count"`
}

type HistoryEntry struct {
	URL           string `json:"url"`
	Title         string `json:"title"`
	VisitCount    int    `json:"visit_count"`
	LastVisitTime string `json:"last_visit_time"`
}

type DownloadEntry struct {
	TargetPath string `json:"target_path"`
	TabURL     string `json:"tab_url"`
	TotalBytes int64  `json:"total_bytes"`
	State      int    `json:"state"`
	EndTime    string `json:"end_time"`
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

func Collect(profileDir string) (Snapshot, error) {
	snapshot := Snapshot{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		ProfileDir:  profileDir,
		History: Artifact{
			Kind:   "history",
			Format: formatSQLite,
			Status: statusPending,
			Notes:  "History rows are not mirrored yet; this record only captures source metadata.",
		},
		Bookmarks: BookmarkArtifact{
			Artifact: Artifact{
				Kind:   "bookmarks",
				Format: formatJSON,
				Status: statusPending,
				Notes:  "Bookmarks source metadata and root summary are mirrored.",
			},
		},
		Downloads: Artifact{
			Kind:   "downloads",
			Format: formatSQLite,
			Status: statusPending,
			Notes:  "Download rows are not mirrored yet; this record only captures source metadata from the Chrome History database.",
		},
	}

	historyPath := filepath.Join(profileDir, "History")
	bookmarksPath := filepath.Join(profileDir, "Bookmarks")

	historyArtifact, err := scanArtifact(snapshot.History, historyPath)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot.History = historyArtifact
	if snapshot.History.Exists {
		snapshot.HistoryRows, err = readHistoryRows(historyPath, 10)
		if err != nil {
			return Snapshot{}, err
		}
	}

	downloadsArtifact, err := scanArtifact(snapshot.Downloads, historyPath)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot.Downloads = downloadsArtifact
	if snapshot.Downloads.Exists {
		snapshot.DownloadRows, err = readDownloadRows(historyPath, 10)
		if err != nil {
			return Snapshot{}, err
		}
	}

	bookmarksArtifact, err := scanBookmarks(snapshot.Bookmarks, bookmarksPath)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot.Bookmarks = bookmarksArtifact

	return snapshot, nil
}

func Save(paths macos.Paths, snapshot Snapshot) error {
	if err := macos.EnsureDir(paths.MirrorsRoot); err != nil {
		return err
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(paths.MirrorFile, append(data, '\n'), 0o644)
}

func Load(paths macos.Paths) (Snapshot, error) {
	data, err := os.ReadFile(paths.MirrorFile)
	if err != nil {
		return Snapshot{}, err
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func DefaultProfilePath(paths macos.Paths) string {
	return filepath.Join(paths.ProfilesRoot, defaultProfileDir)
}

func (s Snapshot) Render(paths macos.Paths) string {
	lines := []string{
		fmt.Sprintf("mirror_file=%s", paths.MirrorFile),
		fmt.Sprintf("profile_dir=%s", s.ProfileDir),
		renderArtifact("history", s.History),
		renderArtifact("downloads", s.Downloads),
		renderArtifact("bookmarks", s.Bookmarks.Artifact),
		fmt.Sprintf("history_rows=%d", len(s.HistoryRows)),
		fmt.Sprintf("download_rows=%d", len(s.DownloadRows)),
	}

	if len(s.Bookmarks.RootSummaries) > 0 {
		rootParts := make([]string, 0, len(s.Bookmarks.RootSummaries))
		for _, root := range s.Bookmarks.RootSummaries {
			rootParts = append(rootParts, fmt.Sprintf("%s:%df/%du", root.Name, root.FolderCount, root.URLCount))
		}
		lines = append(lines, "bookmarks_roots="+strings.Join(rootParts, ","))
	}

	return strings.Join(lines, "\n") + "\n"
}

func renderArtifact(prefix string, artifact Artifact) string {
	return fmt.Sprintf(
		"%s_status=%s %s_exists=%t %s_format=%s %s_source=%s",
		prefix,
		artifact.Status,
		prefix,
		artifact.Exists,
		prefix,
		artifact.Format,
		prefix,
		artifact.SourcePath,
	)
}

func scanArtifact(artifact Artifact, sourcePath string) (Artifact, error) {
	artifact.SourcePath = sourcePath

	info, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			artifact.Status = statusMissing
			return artifact, nil
		}
		return Artifact{}, err
	}

	artifact.Exists = true
	artifact.SizeBytes = info.Size()
	artifact.ModifiedAt = info.ModTime().UTC().Format(time.RFC3339)
	artifact.Status = statusScanned
	return artifact, nil
}

func scanBookmarks(artifact BookmarkArtifact, sourcePath string) (BookmarkArtifact, error) {
	scanned, err := scanArtifact(artifact.Artifact, sourcePath)
	if err != nil {
		return BookmarkArtifact{}, err
	}
	artifact.Artifact = scanned
	if !artifact.Exists {
		return artifact, nil
	}

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return BookmarkArtifact{}, err
	}

	var payload chromeBookmarks
	if err := json.Unmarshal(data, &payload); err != nil {
		return BookmarkArtifact{}, err
	}

	artifact.RootSummaries = make([]BookmarkRootSummary, 0, len(payload.Roots))
	for name, root := range payload.Roots {
		folders, urls := summarizeBookmarkRoot(root)
		artifact.RootSummaries = append(artifact.RootSummaries, BookmarkRootSummary{
			Name:        name,
			FolderCount: folders,
			URLCount:    urls,
		})
	}
	return artifact, nil
}

func summarizeBookmarkRoot(root bookmarkNode) (folderCount int, urlCount int) {
	for _, child := range root.Children {
		childFolders, childURLs := summarizeBookmarkNode(child)
		folderCount += childFolders
		urlCount += childURLs
	}
	return folderCount, urlCount
}

func summarizeBookmarkNode(node bookmarkNode) (folderCount int, urlCount int) {
	switch node.Type {
	case "url":
		return 0, 1
	case "folder":
		folderCount = 1
		for _, child := range node.Children {
			childFolders, childURLs := summarizeBookmarkNode(child)
			folderCount += childFolders
			urlCount += childURLs
		}
	}
	return folderCount, urlCount
}

func readHistoryRows(sourcePath string, limit int) ([]HistoryEntry, error) {
	const query = "select url, title, visit_count, last_visit_time from urls order by last_visit_time desc limit %d;"
	rows, err := querySQLiteJSON(sourcePath, fmt.Sprintf(query, limit))
	if err != nil {
		return nil, err
	}

	type historyRow struct {
		URL           string `json:"url"`
		Title         string `json:"title"`
		VisitCount    int    `json:"visit_count"`
		LastVisitTime int64  `json:"last_visit_time"`
	}

	var payload []historyRow
	if err := json.Unmarshal(rows, &payload); err != nil {
		return nil, err
	}

	entries := make([]HistoryEntry, 0, len(payload))
	for _, row := range payload {
		entries = append(entries, HistoryEntry{
			URL:           row.URL,
			Title:         row.Title,
			VisitCount:    row.VisitCount,
			LastVisitTime: chromeTimeToRFC3339(row.LastVisitTime),
		})
	}
	return entries, nil
}

func readDownloadRows(sourcePath string, limit int) ([]DownloadEntry, error) {
	const query = "select target_path, tab_url, total_bytes, state, end_time from downloads order by end_time desc limit %d;"
	rows, err := querySQLiteJSON(sourcePath, fmt.Sprintf(query, limit))
	if err != nil {
		return nil, err
	}

	type downloadRow struct {
		TargetPath string `json:"target_path"`
		TabURL     string `json:"tab_url"`
		TotalBytes int64  `json:"total_bytes"`
		State      int    `json:"state"`
		EndTime    int64  `json:"end_time"`
	}

	var payload []downloadRow
	if err := json.Unmarshal(rows, &payload); err != nil {
		return nil, err
	}

	entries := make([]DownloadEntry, 0, len(payload))
	for _, row := range payload {
		entries = append(entries, DownloadEntry{
			TargetPath: row.TargetPath,
			TabURL:     row.TabURL,
			TotalBytes: row.TotalBytes,
			State:      row.State,
			EndTime:    chromeTimeToRFC3339(row.EndTime),
		})
	}
	return entries, nil
}

func querySQLiteJSON(sourcePath string, query string) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "atlasx-sqlite-*")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	copyPath := filepath.Join(tmpDir, "source.sqlite")
	if err := copySQLiteSource(sourcePath, copyPath); err != nil {
		return nil, err
	}

	cmd := exec.Command("sqlite3", "-json", copyPath, query)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return output, nil
}

func copySQLiteSource(sourcePath string, destPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = source.Close()
	}()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = dest.Close()
	}()

	if _, err := io.Copy(dest, source); err != nil {
		return err
	}
	return dest.Close()
}

func chromeTimeToRFC3339(value int64) string {
	if value <= 0 {
		return ""
	}
	const microsecondsBetweenWindowsAndUnixEpoch = int64(11644473600000000)
	unixMicroseconds := value - microsecondsBetweenWindowsAndUnixEpoch
	return time.UnixMicro(unixMicroseconds).UTC().Format(time.RFC3339)
}
