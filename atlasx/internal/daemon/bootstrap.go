package daemon

import (
	"fmt"
	"net/http"
	"os"

	"atlasx/internal/browserdata"
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

func NewMux(_ Status) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", serveStatus)
	mux.HandleFunc("/v1/status", serveStatus)
	mux.HandleFunc("/v1/history", func(w http.ResponseWriter, _ *http.Request) {
		serveBrowserData(w, browserdata.LoadHistory)
	})
	mux.HandleFunc("/v1/downloads", func(w http.ResponseWriter, _ *http.Request) {
		serveBrowserData(w, browserdata.LoadDownloads)
	})
	mux.HandleFunc("/v1/bookmarks", func(w http.ResponseWriter, _ *http.Request) {
		serveBrowserData(w, browserdata.LoadBookmarks)
	})
	mux.HandleFunc("/v1/history/open", func(w http.ResponseWriter, r *http.Request) {
		serveBrowserOpenAction(w, r, browserdata.ResolveHistoryURL, "opened_history_index")
	})
	mux.HandleFunc("/v1/downloads/open", func(w http.ResponseWriter, r *http.Request) {
		serveBrowserOpenAction(w, r, browserdata.ResolveDownloadURL, "opened_download_index")
	})
	mux.HandleFunc("/v1/bookmarks/open", func(w http.ResponseWriter, r *http.Request) {
		serveBrowserOpenAction(w, r, browserdata.ResolveBookmarkURL, "opened_bookmark_index")
	})
	mux.HandleFunc("/v1/tabs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
			return
		}
		serveTabsList(w)
	})
	mux.HandleFunc("/v1/tabs/open", func(w http.ResponseWriter, r *http.Request) {
		serveTabAction(w, r, func(client tabClient, request tabActionRequest) (any, error) {
			return client.Open(request.URL)
		})
	})
	mux.HandleFunc("/v1/tabs/activate", func(w http.ResponseWriter, r *http.Request) {
		serveTabAction(w, r, func(client tabClient, request tabActionRequest) (any, error) {
			return map[string]string{"activated": request.ID}, client.Activate(request.ID)
		})
	})
	mux.HandleFunc("/v1/tabs/close", func(w http.ResponseWriter, r *http.Request) {
		serveTabAction(w, r, func(client tabClient, request tabActionRequest) (any, error) {
			return map[string]string{"closed": request.ID}, client.Close(request.ID)
		})
	})
	mux.HandleFunc("/v1/tabs/navigate", func(w http.ResponseWriter, r *http.Request) {
		serveTabAction(w, r, func(client tabClient, request tabActionRequest) (any, error) {
			return map[string]string{"navigated": request.ID, "url": request.URL}, client.Navigate(request.ID, request.URL)
		})
	})
	mux.HandleFunc("/v1/mirror/scan", func(w http.ResponseWriter, r *http.Request) {
		serveMirrorScan(w, r)
	})
	mux.HandleFunc("/v1/import/chrome", func(w http.ResponseWriter, r *http.Request) {
		serveChromeImport(w, r)
	})
	mux.HandleFunc("/v1/import/safari", func(w http.ResponseWriter, r *http.Request) {
		serveSafariImport(w, r)
	})
	return mux
}
