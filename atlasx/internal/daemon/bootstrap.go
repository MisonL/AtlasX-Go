package daemon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"atlasx/internal/browserdata"
	"atlasx/internal/diagnostics"
	"atlasx/internal/imports"
	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/tabs"
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
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		status, err := Bootstrap()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, status)
	})
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, _ *http.Request) {
		status, err := Bootstrap()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, status)
	})
	mux.HandleFunc("/v1/history", func(w http.ResponseWriter, _ *http.Request) {
		serveBrowserData(w, browserdata.LoadHistory)
	})
	mux.HandleFunc("/v1/downloads", func(w http.ResponseWriter, _ *http.Request) {
		serveBrowserData(w, browserdata.LoadDownloads)
	})
	mux.HandleFunc("/v1/bookmarks", func(w http.ResponseWriter, _ *http.Request) {
		serveBrowserData(w, browserdata.LoadBookmarks)
	})
	mux.HandleFunc("/v1/tabs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
			return
		}
		serveTabsList(w)
	})
	mux.HandleFunc("/v1/tabs/open", func(w http.ResponseWriter, r *http.Request) {
		serveTabAction(w, r, func(client tabs.Client, request tabActionRequest) (any, error) {
			return client.Open(request.URL)
		})
	})
	mux.HandleFunc("/v1/tabs/activate", func(w http.ResponseWriter, r *http.Request) {
		serveTabAction(w, r, func(client tabs.Client, request tabActionRequest) (any, error) {
			return map[string]string{"activated": request.ID}, client.Activate(request.ID)
		})
	})
	mux.HandleFunc("/v1/tabs/close", func(w http.ResponseWriter, r *http.Request) {
		serveTabAction(w, r, func(client tabs.Client, request tabActionRequest) (any, error) {
			return map[string]string{"closed": request.ID}, client.Close(request.ID)
		})
	})
	mux.HandleFunc("/v1/tabs/navigate", func(w http.ResponseWriter, r *http.Request) {
		serveTabAction(w, r, func(client tabs.Client, request tabActionRequest) (any, error) {
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

func (s Status) Render() string {
	payload, _ := json.MarshalIndent(s, "", "  ")
	return string(payload) + "\n"
}

func writeJSON(w http.ResponseWriter, code int, payload Status) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}

type dataLoader[T any] func(macos.Paths) ([]T, error)
type tabActionRequest struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}
type mirrorScanRequest struct {
	ProfileDir string `json:"profile_dir"`
}
type chromeImportRequest struct {
	SourceProfileDir string `json:"source_profile_dir"`
}

type tabAction func(client tabs.Client, request tabActionRequest) (any, error)

func serveBrowserData[T any](w http.ResponseWriter, loader dataLoader[T]) {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	payload, err := loader(paths)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusConflict, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

func serveTabsList(w http.ResponseWriter) {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	client, err := tabs.New(paths)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}

	targets, err := client.List()
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(tabs.PageTargets(targets))
}

func serveTabAction(w http.ResponseWriter, r *http.Request, action tabAction) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request tabActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	client, err := tabs.New(paths)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}

	payload, err := action(client, request)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(payload)
}

func serveMirrorScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request mirrorScanRequest
	if err := decodeOptionalJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	profileDir := request.ProfileDir
	if profileDir == "" {
		profileDir = mirror.DefaultProfilePath(paths)
	}

	snapshot, err := mirror.Collect(profileDir)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if err := mirror.Save(paths, snapshot); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(snapshot)
}

func serveChromeImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request chromeImportRequest
	if err := decodeOptionalJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	sourceProfileDir := request.SourceProfileDir
	if sourceProfileDir == "" {
		sourceProfileDir = imports.DefaultChromeProfileDir(paths)
	}

	report, err := imports.ImportChrome(paths, sourceProfileDir)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(report)
}

func serveSafariImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	report, err := imports.ImportSafari(paths)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(report)
}

func decodeOptionalJSON(r *http.Request, target any) error {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
