package mirror

import (
	"os"
	"os/exec"
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

func TestCollectReadsHistoryAndDownloadRows(t *testing.T) {
	if _, err := exec.LookPath("sqlite3"); err != nil {
		t.Skip("sqlite3 is not available")
	}

	profileDir := t.TempDir()
	dbPath := filepath.Join(profileDir, "History")
	setupSQLiteFixture(t, dbPath)

	snapshot, err := Collect(profileDir)
	if err != nil {
		t.Fatalf("collect failed: %v", err)
	}

	if len(snapshot.HistoryRows) != 1 {
		t.Fatalf("unexpected history row count: %d", len(snapshot.HistoryRows))
	}
	if snapshot.HistoryRows[0].URL != "https://openai.com" {
		t.Fatalf("unexpected history url: %s", snapshot.HistoryRows[0].URL)
	}
	if len(snapshot.DownloadRows) != 1 {
		t.Fatalf("unexpected download row count: %d", len(snapshot.DownloadRows))
	}
	if snapshot.DownloadRows[0].TargetPath != "/tmp/file.zip" {
		t.Fatalf("unexpected download target path: %s", snapshot.DownloadRows[0].TargetPath)
	}
}

func setupSQLiteFixture(t *testing.T, dbPath string) {
	t.Helper()

	sql := `
create table urls(id integer primary key autoincrement, url longvarchar, title longvarchar, visit_count integer not null, typed_count integer not null, last_visit_time integer not null, hidden integer not null);
create table downloads(id integer primary key, guid varchar not null, current_path longvarchar not null, target_path longvarchar not null, start_time integer not null, received_bytes integer not null, total_bytes integer not null, state integer not null, danger_type integer not null, interrupt_reason integer not null, hash blob not null, end_time integer not null, opened integer not null, last_access_time integer not null, transient integer not null, referrer varchar not null, site_url varchar not null, embedder_download_data varchar not null, tab_url varchar not null, tab_referrer_url varchar not null, http_method varchar not null, by_ext_id varchar not null, by_ext_name varchar not null, by_web_app_id varchar not null, etag varchar not null, last_modified varchar not null, mime_type varchar not null, original_mime_type varchar not null);
insert into urls(url, title, visit_count, typed_count, last_visit_time, hidden) values('https://openai.com', 'OpenAI', 3, 0, 13217472000000000, 0);
insert into downloads(id, guid, current_path, target_path, start_time, received_bytes, total_bytes, state, danger_type, interrupt_reason, hash, end_time, opened, last_access_time, transient, referrer, site_url, embedder_download_data, tab_url, tab_referrer_url, http_method, by_ext_id, by_ext_name, by_web_app_id, etag, last_modified, mime_type, original_mime_type) values(1, 'guid', '/tmp/file.zip', '/tmp/file.zip', 13217472000000000, 10, 10, 1, 0, 0, X'', 13217472000000000, 1, 13217472000000000, 0, '', 'https://openai.com', '', 'https://openai.com', '', 'GET', '', '', '', '', '', 'application/zip', 'application/zip');
`

	cmd := exec.Command("sqlite3", dbPath)
	cmd.Stdin = strings.NewReader(sql)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("sqlite fixture setup failed: %v: %s", err, string(output))
	}
}
