package daemon

import (
	"encoding/json"
	"net/http"
	"os"

	"atlasx/internal/diagnostics"
	"atlasx/internal/imports"
	"atlasx/internal/mirror"
	"atlasx/internal/settings"
)

const DefaultListenAddr = settings.DefaultListenAddr

type Status struct {
	Ready                 bool   `json:"ready"`
	ChromeStatus          string `json:"chrome_status"`
	SupportRoot           string `json:"support_root"`
	ConfigFile            string `json:"config_file"`
	ManagedSessionLive    bool   `json:"managed_session_live"`
	ManagedSessionCDP     string `json:"managed_session_cdp"`
	ManagedSessionCDPURL  string `json:"managed_session_cdp_url"`
	MirrorFile            string `json:"mirror_file"`
	MirrorPresent         bool   `json:"mirror_present"`
	MirrorProfileDir      string `json:"mirror_profile_dir"`
	MirrorHistoryRows     int    `json:"mirror_history_rows"`
	MirrorDownloadRows    int    `json:"mirror_download_rows"`
	ChromeImportPresent   bool   `json:"chrome_import_present"`
	ChromeImportRoot      string `json:"chrome_import_root"`
	ChromeImportBookmarks bool   `json:"chrome_import_bookmarks"`
	ChromeImportHistory   bool   `json:"chrome_import_history"`
}

func Bootstrap() (Status, error) {
	report, err := diagnostics.Generate()
	if err != nil {
		return Status{}, err
	}

	status := Status{
		Ready:                report.ChromeStatus == "ok",
		ChromeStatus:         report.ChromeStatus,
		SupportRoot:          report.Paths.SupportRoot,
		ConfigFile:           report.Paths.ConfigFile,
		ManagedSessionLive:   report.Session.Alive,
		ManagedSessionCDP:    report.Session.CDP.Status,
		ManagedSessionCDPURL: report.Session.CDP.VersionEndpoint,
		MirrorFile:           report.Paths.MirrorFile,
	}

	if snapshot, err := mirror.Load(report.Paths); err == nil {
		status.MirrorPresent = true
		status.MirrorProfileDir = snapshot.ProfileDir
		status.MirrorHistoryRows = len(snapshot.HistoryRows)
		status.MirrorDownloadRows = len(snapshot.DownloadRows)
	} else if !os.IsNotExist(err) {
		return Status{}, err
	}

	if chromeImport, err := imports.LoadChromeReport(report.Paths); err == nil {
		status.ChromeImportPresent = true
		status.ChromeImportRoot = chromeImport.ImportRoot
		status.ChromeImportBookmarks = chromeImport.BookmarksImported.Exists
		status.ChromeImportHistory = chromeImport.HistorySource.Exists
	} else if !os.IsNotExist(err) {
		return Status{}, err
	}

	return status, nil
}

func NewMux(status Status) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, status)
	})
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, status)
	})
	return mux
}

func (s Status) Render() string {
	payload, _ := json.MarshalIndent(s, "", "  ")
	return string(payload) + "\n"
}

func writeJSON(w http.ResponseWriter, code int, payload Status) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
