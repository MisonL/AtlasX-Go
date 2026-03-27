package imports

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
)

const chromeDefaultProfileDir = "Default"

type ChromeReport struct {
	SourceProfileDir    string          `json:"source_profile_dir"`
	GeneratedAt         string          `json:"generated_at"`
	ImportRoot          string          `json:"import_root"`
	BookmarksSource     FileArtifact    `json:"bookmarks_source"`
	BookmarksImported   FileArtifact    `json:"bookmarks_imported"`
	PreferencesSource   FileArtifact    `json:"preferences_source"`
	PreferencesImported FileArtifact    `json:"preferences_imported"`
	HistorySource       FileArtifact    `json:"history_source"`
	SourceMirror        mirror.Snapshot `json:"source_mirror"`
}

type FileArtifact struct {
	Path       string `json:"path"`
	Exists     bool   `json:"exists"`
	SizeBytes  int64  `json:"size_bytes"`
	ModifiedAt string `json:"modified_at,omitempty"`
}

func DefaultChromeProfileDir(paths macos.Paths) string {
	return filepath.Join(paths.Home, "Library", "Application Support", "Google", "Chrome", chromeDefaultProfileDir)
}

func DefaultChromeImportRoot(paths macos.Paths) string {
	return filepath.Join(paths.ImportsRoot, "chrome", chromeDefaultProfileDir)
}

func ImportChrome(paths macos.Paths, sourceProfileDir string) (ChromeReport, error) {
	snapshot, err := mirror.Collect(sourceProfileDir)
	if err != nil {
		return ChromeReport{}, err
	}
	if !snapshot.Bookmarks.Exists {
		return ChromeReport{}, errors.New("chrome bookmarks source is missing")
	}

	importRoot := DefaultChromeImportRoot(paths)
	if err := macos.EnsureDir(importRoot); err != nil {
		return ChromeReport{}, err
	}

	bookmarksDest := filepath.Join(importRoot, "Bookmarks.json")
	preferencesDest := filepath.Join(importRoot, "Preferences.json")
	reportPath := filepath.Join(importRoot, "report.json")

	bookmarksSource, err := fileArtifact(snapshot.Bookmarks.SourcePath)
	if err != nil {
		return ChromeReport{}, err
	}
	if err := copyFile(snapshot.Bookmarks.SourcePath, bookmarksDest); err != nil {
		return ChromeReport{}, err
	}
	bookmarksImported, err := fileArtifact(bookmarksDest)
	if err != nil {
		return ChromeReport{}, err
	}

	preferencesSourcePath := filepath.Join(sourceProfileDir, "Preferences")
	preferencesSource, err := fileArtifact(preferencesSourcePath)
	if err != nil {
		return ChromeReport{}, err
	}
	preferencesImported := FileArtifact{Path: preferencesDest}
	if preferencesSource.Exists {
		if err := copyFile(preferencesSourcePath, preferencesDest); err != nil {
			return ChromeReport{}, err
		}
		preferencesImported, err = fileArtifact(preferencesDest)
		if err != nil {
			return ChromeReport{}, err
		}
	}

	historySource, err := fileArtifact(snapshot.History.SourcePath)
	if err != nil {
		return ChromeReport{}, err
	}

	report := ChromeReport{
		SourceProfileDir:    sourceProfileDir,
		GeneratedAt:         time.Now().UTC().Format(time.RFC3339),
		ImportRoot:          importRoot,
		BookmarksSource:     bookmarksSource,
		BookmarksImported:   bookmarksImported,
		PreferencesSource:   preferencesSource,
		PreferencesImported: preferencesImported,
		HistorySource:       historySource,
		SourceMirror:        snapshot,
	}

	if err := saveReport(reportPath, report); err != nil {
		return ChromeReport{}, err
	}
	return report, nil
}

func (r ChromeReport) Render() string {
	return fmt.Sprintf(
		"import_root=%s\nsource_profile_dir=%s\nbookmarks_source=%s\nbookmarks_imported=%s\npreferences_source=%s\npreferences_imported=%s\nhistory_source=%s\nbookmarks_roots=%d\n",
		r.ImportRoot,
		r.SourceProfileDir,
		r.BookmarksSource.Path,
		r.BookmarksImported.Path,
		r.PreferencesSource.Path,
		r.PreferencesImported.Path,
		r.HistorySource.Path,
		len(r.SourceMirror.Bookmarks.RootSummaries),
	)
}

func copyFile(sourcePath string, destPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		return err
	}
	return dest.Close()
}

func fileArtifact(path string) (FileArtifact, error) {
	artifact := FileArtifact{Path: path}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return artifact, nil
		}
		return FileArtifact{}, err
	}
	artifact.Exists = true
	artifact.SizeBytes = info.Size()
	artifact.ModifiedAt = info.ModTime().UTC().Format(time.RFC3339)
	return artifact, nil
}

func saveReport(path string, report ChromeReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func LoadChromeReport(paths macos.Paths) (ChromeReport, error) {
	reportPath := filepath.Join(DefaultChromeImportRoot(paths), "report.json")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return ChromeReport{}, err
	}

	var report ChromeReport
	if err := json.Unmarshal(data, &report); err != nil {
		return ChromeReport{}, err
	}
	return report, nil
}
