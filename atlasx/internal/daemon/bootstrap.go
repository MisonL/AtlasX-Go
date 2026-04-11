package daemon

import (
	"net/http"
	"os"

	"atlasx/internal/browserdata"
	"atlasx/internal/diagnostics"
	"atlasx/internal/imports"
	"atlasx/internal/logs"
	"atlasx/internal/managedruntime"
	"atlasx/internal/memory"
	"atlasx/internal/mirror"
	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
	"atlasx/internal/updates"
)

const DefaultListenAddr = settings.DefaultListenAddr

type Status struct {
	Ready                      bool                        `json:"ready"`
	ChromeStatus               string                      `json:"chrome_status"`
	ChromeSource               string                      `json:"chrome_source"`
	SupportRoot                string                      `json:"support_root"`
	ConfigFile                 string                      `json:"config_file"`
	ManagedSessionLive         bool                        `json:"managed_session_live"`
	ManagedSessionStale        bool                        `json:"managed_session_stale"`
	ManagedSessionStateCleaned bool                        `json:"managed_session_state_cleaned"`
	ManagedSessionCDP          string                      `json:"managed_session_cdp"`
	ManagedSessionCDPURL       string                      `json:"managed_session_cdp_url"`
	MirrorFile                 string                      `json:"mirror_file"`
	MirrorPresent              bool                        `json:"mirror_present"`
	MirrorProfileDir           string                      `json:"mirror_profile_dir"`
	MirrorHistoryRows          int                         `json:"mirror_history_rows"`
	MirrorDownloadRows         int                         `json:"mirror_download_rows"`
	MirrorLastScanAt           string                      `json:"mirror_last_scan_at"`
	MirrorLastScanSource       string                      `json:"mirror_last_scan_source"`
	MirrorLastScanResult       string                      `json:"mirror_last_scan_result"`
	MirrorLastScanError        string                      `json:"mirror_last_scan_error"`
	ChromeImportPresent        bool                        `json:"chrome_import_present"`
	ChromeImportRoot           string                      `json:"chrome_import_root"`
	ChromeImportBookmarks      bool                        `json:"chrome_import_bookmarks"`
	ChromeImportHistory        bool                        `json:"chrome_import_history"`
	ChromeImportLastAt         string                      `json:"chrome_import_last_at"`
	ChromeImportLastSource     string                      `json:"chrome_import_last_source"`
	ChromeImportLastResult     string                      `json:"chrome_import_last_result"`
	ChromeImportLastError      string                      `json:"chrome_import_last_error"`
	SafariImportLastAt         string                      `json:"safari_import_last_at"`
	SafariImportLastSource     string                      `json:"safari_import_last_source"`
	SafariImportLastResult     string                      `json:"safari_import_last_result"`
	SafariImportLastError      string                      `json:"safari_import_last_error"`
	LogsPresent                bool                        `json:"logs_present"`
	LogsFileCount              int                         `json:"logs_file_count"`
	LogsLatestFile             string                      `json:"logs_latest_file"`
	LogsLatestAt               string                      `json:"logs_latest_at"`
	MemoryRoot                 string                      `json:"memory_root"`
	MemoryEventsFile           string                      `json:"memory_events_file"`
	MemoryPresent              bool                        `json:"memory_present"`
	MemoryEventCount           int                         `json:"memory_event_count"`
	MemoryLastEventAt          string                      `json:"memory_last_event_at"`
	MemoryLastEventKind        string                      `json:"memory_last_event_kind"`
	RuntimeManifestPath        string                      `json:"runtime_manifest_path"`
	RuntimeManifestPresent     bool                        `json:"runtime_manifest_present"`
	RuntimeManifestVersion     string                      `json:"runtime_manifest_version"`
	RuntimeManifestChannel     string                      `json:"runtime_manifest_channel"`
	RuntimeManifestSHA256      string                      `json:"runtime_manifest_sha256"`
	RuntimeManifestBundlePath  string                      `json:"runtime_manifest_bundle_path"`
	RuntimeManifestBinaryPath  string                      `json:"runtime_manifest_binary_path"`
	RuntimeBundlePresent       bool                        `json:"runtime_bundle_present"`
	RuntimeBinaryPresent       bool                        `json:"runtime_binary_present"`
	RuntimeBinaryExecutable    bool                        `json:"runtime_binary_executable"`
	UpdatesPlanPresent         bool                        `json:"updates_plan_present"`
	UpdatesPlanPending         bool                        `json:"updates_plan_pending"`
	UpdatesPlanInFlight        bool                        `json:"updates_plan_in_flight"`
	UpdatesPlanPhase           managedruntime.InstallPhase `json:"updates_plan_phase"`
	SidebarQAConfigured        bool                        `json:"sidebar_qa_configured"`
	SidebarQAReady             bool                        `json:"sidebar_qa_ready"`
	SidebarQADefaultProvider   string                      `json:"sidebar_qa_default_provider"`
	SidebarQAProvider          string                      `json:"sidebar_qa_provider"`
	SidebarQAModel             string                      `json:"sidebar_qa_model"`
	SidebarQAProviders         []sidebar.ProviderStatus    `json:"sidebar_qa_providers"`
	SidebarQATimeoutMS         int                         `json:"sidebar_qa_timeout_ms"`
	SidebarQARetryAttempts     int                         `json:"sidebar_qa_retry_attempts"`
	SidebarQATokenBudget       int                         `json:"sidebar_qa_token_budget"`
	SidebarQALastTraceID       string                      `json:"sidebar_qa_last_trace_id"`
	SidebarQALastError         string                      `json:"sidebar_qa_last_error"`
	SidebarQALastErrorAt       string                      `json:"sidebar_qa_last_error_at"`
}

func Bootstrap() (Status, error) {
	report, err := diagnostics.Generate()
	if err != nil {
		return Status{}, err
	}

	status := Status{
		Ready:                      report.ChromeStatus == "ok",
		ChromeStatus:               report.ChromeStatus,
		ChromeSource:               report.Chrome.Source,
		SupportRoot:                report.Paths.SupportRoot,
		ConfigFile:                 report.Paths.ConfigFile,
		ManagedSessionLive:         report.Session.Ready,
		ManagedSessionStale:        report.Session.Stale,
		ManagedSessionStateCleaned: report.Session.StateCleaned,
		ManagedSessionCDP:          report.Session.CDP.Status,
		ManagedSessionCDPURL:       report.Session.CDP.VersionEndpoint,
		MirrorFile:                 report.Paths.MirrorFile,
		MemoryRoot:                 report.Paths.MemoryRoot,
		MemoryEventsFile:           report.Paths.MemoryEventsFile,
		RuntimeManifestPath:        report.RuntimeManifest.Path,
	}

	if snapshot, err := mirror.Load(report.Paths); err == nil {
		status.MirrorPresent = true
		status.MirrorProfileDir = snapshot.ProfileDir
		status.MirrorHistoryRows = len(snapshot.HistoryRows)
		status.MirrorDownloadRows = len(snapshot.DownloadRows)
	} else if !os.IsNotExist(err) {
		return Status{}, err
	}
	if scanStatus, err := mirror.LoadScanStatus(report.Paths); err == nil {
		status.MirrorLastScanAt = scanStatus.GeneratedAt
		status.MirrorLastScanSource = scanStatus.ProfileDir
		status.MirrorLastScanResult = scanStatus.Result
		status.MirrorLastScanError = scanStatus.Error
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
	if chromeImportStatus, err := imports.LoadChromeImportStatus(report.Paths); err == nil {
		status.ChromeImportLastAt = chromeImportStatus.GeneratedAt
		status.ChromeImportLastSource = chromeImportStatus.Source
		status.ChromeImportLastResult = chromeImportStatus.Result
		status.ChromeImportLastError = chromeImportStatus.Error
	} else if !os.IsNotExist(err) {
		return Status{}, err
	}
	if safariImportStatus, err := imports.LoadSafariImportStatus(report.Paths); err == nil {
		status.SafariImportLastAt = safariImportStatus.GeneratedAt
		status.SafariImportLastSource = safariImportStatus.Source
		status.SafariImportLastResult = safariImportStatus.Result
		status.SafariImportLastError = safariImportStatus.Error
	} else if !os.IsNotExist(err) {
		return Status{}, err
	}
	memorySummary, err := memory.LoadSummary(report.Paths)
	if err != nil {
		return Status{}, err
	}
	status.MemoryPresent = memorySummary.Present
	status.MemoryEventCount = memorySummary.EventCount
	status.MemoryLastEventAt = memorySummary.LastEventAt
	status.MemoryLastEventKind = memorySummary.LastEventKind
	logsStatus, err := logs.LoadStatus(report.Paths, 1)
	if err != nil {
		return Status{}, err
	}
	status.LogsPresent = logsStatus.Present
	status.LogsFileCount = logsStatus.FileCount
	status.LogsLatestFile = logsStatus.LatestFile
	status.LogsLatestAt = logsStatus.LatestAt
	status.RuntimeManifestPresent = report.RuntimeManifest.Present
	status.RuntimeManifestVersion = report.RuntimeManifest.Version
	status.RuntimeManifestChannel = report.RuntimeManifest.Channel
	status.RuntimeManifestSHA256 = report.RuntimeManifest.SHA256
	status.RuntimeManifestBundlePath = report.RuntimeManifest.BundlePath
	status.RuntimeManifestBinaryPath = report.RuntimeManifest.BinaryPath
	runtimeStatus, err := managedruntime.Status(report.Paths)
	if err != nil {
		return Status{}, err
	}
	status.RuntimeBundlePresent = runtimeStatus.BundlePresent
	status.RuntimeBinaryPresent = runtimeStatus.BinaryPresent
	status.RuntimeBinaryExecutable = runtimeStatus.BinaryExecutable
	updatesStatus, err := updates.LoadStatus(report.Paths)
	if err != nil {
		return Status{}, err
	}
	status.UpdatesPlanPresent = updatesStatus.PlanPresent
	status.UpdatesPlanPending = updatesStatus.PlanPending
	status.UpdatesPlanInFlight = updatesStatus.PlanInFlight
	status.UpdatesPlanPhase = updatesStatus.PlanPhase

	config, err := settings.NewStore(report.Paths.ConfigFile).Load()
	if err != nil {
		return Status{}, err
	}
	sidebarStatus, err := sidebar.FromSettings(config).StatusWithRuntime(report.Paths)
	if err != nil {
		return Status{}, err
	}
	status.SidebarQAConfigured = sidebarStatus.Configured
	status.SidebarQAReady = sidebarStatus.Ready
	status.SidebarQADefaultProvider = sidebarStatus.DefaultProvider
	status.SidebarQAProvider = sidebarStatus.Provider
	status.SidebarQAModel = sidebarStatus.Model
	status.SidebarQAProviders = sidebarStatus.Providers
	status.SidebarQATimeoutMS = sidebarStatus.TimeoutMS
	status.SidebarQARetryAttempts = sidebarStatus.RetryAttempts
	status.SidebarQATokenBudget = sidebarStatus.TokenBudget
	status.SidebarQALastTraceID = sidebarStatus.LastTraceID
	status.SidebarQALastError = sidebarStatus.LastError
	status.SidebarQALastErrorAt = sidebarStatus.LastErrorAt

	return status, nil
}

func NewMux(_ Status) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", requireMethods(serveStatus, http.MethodGet))
	mux.HandleFunc("/v1/status", requireMethods(serveStatus, http.MethodGet))
	mux.HandleFunc("/v1/doctor", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveDoctor(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/profile", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveProfile(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/policy", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		servePolicy(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/permissions", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		servePermissions(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/settings", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveSettings(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/default-browser", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveDefaultBrowserStatus(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/default-browser/set", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveDefaultBrowserSet(w, r)
	}, http.MethodPost))
	mux.HandleFunc("/v1/logs", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveLogsStatus(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/updates", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveUpdatesStatus(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/memory", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveMemoryList(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/memory/controls", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveMemoryControls(w, r)
	}, http.MethodGet, http.MethodPost))
	mux.HandleFunc("/v1/memory/search", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveMemorySearch(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/history", requireMethods(func(w http.ResponseWriter, _ *http.Request) {
		serveBrowserData(w, browserdata.LoadHistory)
	}, http.MethodGet))
	mux.HandleFunc("/v1/downloads", requireMethods(func(w http.ResponseWriter, _ *http.Request) {
		serveBrowserData(w, browserdata.LoadDownloads)
	}, http.MethodGet))
	mux.HandleFunc("/v1/bookmarks", requireMethods(func(w http.ResponseWriter, _ *http.Request) {
		serveBrowserData(w, browserdata.LoadBookmarks)
	}, http.MethodGet))
	mux.HandleFunc("/v1/history/open", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveBrowserOpenAction(w, r, browserdata.ResolveHistoryURL, "opened_history_index")
	}, http.MethodPost))
	mux.HandleFunc("/v1/downloads/open", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveBrowserOpenAction(w, r, browserdata.ResolveDownloadURL, "opened_download_index")
	}, http.MethodPost))
	mux.HandleFunc("/v1/bookmarks/open", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveBrowserOpenAction(w, r, browserdata.ResolveBookmarkURL, "opened_bookmark_index")
	}, http.MethodPost))
	mux.HandleFunc("/v1/runtime/status", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveRuntimeStatus(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/runtime/stage", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveRuntimeStage(w, r)
	}, http.MethodPost))
	mux.HandleFunc("/v1/runtime/clear", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveRuntimeClear(w, r)
	}, http.MethodPost))
	mux.HandleFunc("/v1/runtime/verify", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveRuntimeVerify(w, r)
	}, http.MethodPost))
	mux.HandleFunc("/v1/runtime/install", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveRuntimeInstall(w, r)
	}, http.MethodPost))
	mux.HandleFunc("/v1/runtime/plan", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveRuntimePlan(w, r)
	}, http.MethodGet, http.MethodPost))
	mux.HandleFunc("/v1/runtime/plan/clear", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveRuntimePlanClear(w, r)
	}, http.MethodPost))
	mux.HandleFunc("/v1/sidebar/status", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveSidebarStatus(w, r)
	}, http.MethodGet))
	mux.HandleFunc("/v1/sidebar/ask", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveSidebarAsk(w, r)
	}, http.MethodPost))
	mux.HandleFunc("/v1/sidebar/selection/ask", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveSidebarSelectionAsk(w, r)
	}, http.MethodPost))
	mux.HandleFunc("/v1/sidebar/summarize", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveSidebarSummarize(w, r)
	}, http.MethodPost))
	mux.HandleFunc("/v1/tabs", func(w http.ResponseWriter, r *http.Request) {
		serveTabsList(w, r)
	})
	mux.HandleFunc("/v1/tabs/search", func(w http.ResponseWriter, r *http.Request) {
		serveTabSearch(w, r)
	})
	mux.HandleFunc("/v1/tabs/context", func(w http.ResponseWriter, r *http.Request) {
		serveTabContext(w, r)
	})
	mux.HandleFunc("/v1/tabs/auth-mode", func(w http.ResponseWriter, r *http.Request) {
		serveTabAuthMode(w, r)
	})
	mux.HandleFunc("/v1/tabs/semantic-context", func(w http.ResponseWriter, r *http.Request) {
		serveTabSemanticContext(w, r)
	})
	mux.HandleFunc("/v1/tabs/selection", func(w http.ResponseWriter, r *http.Request) {
		serveTabSelection(w, r)
	})
	mux.HandleFunc("/v1/tabs/suggestions", func(w http.ResponseWriter, r *http.Request) {
		serveTabSuggestions(w, r)
	})
	mux.HandleFunc("/v1/tabs/agent-plan", func(w http.ResponseWriter, r *http.Request) {
		serveTabAgentPlan(w, r)
	})
	mux.HandleFunc("/v1/tabs/agent-execute", func(w http.ResponseWriter, r *http.Request) {
		serveTabAgentExecute(w, r)
	})
	mux.HandleFunc("/v1/tabs/memories", func(w http.ResponseWriter, r *http.Request) {
		serveTabMemories(w, r)
	})
	mux.HandleFunc("/v1/tabs/context-recommendations", func(w http.ResponseWriter, r *http.Request) {
		serveTabContextRecommendations(w, r)
	})
	mux.HandleFunc("/v1/tabs/groups", func(w http.ResponseWriter, r *http.Request) {
		serveTabGroups(w, r)
	})
	mux.HandleFunc("/v1/tabs/organize", func(w http.ResponseWriter, r *http.Request) {
		serveTabOrganize(w, r)
	})
	mux.HandleFunc("/v1/tabs/organize-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOrganizeWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/organize-group-to-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOrganizeGroupToWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/organize-group-into-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOrganizeGroupIntoWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/organize-to-windows", func(w http.ResponseWriter, r *http.Request) {
		serveTabOrganizeToWindows(w, r)
	})
	mux.HandleFunc("/v1/tabs/organize-into-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOrganizeIntoWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/organize-window-to-windows", func(w http.ResponseWriter, r *http.Request) {
		serveTabOrganizeWindowToWindows(w, r)
	})
	mux.HandleFunc("/v1/tabs/organize-window-into-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOrganizeWindowIntoWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/organize-window-group-to-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOrganizeWindowGroupToWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/organize-window-group-into-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOrganizeWindowGroupIntoWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/devtools", func(w http.ResponseWriter, r *http.Request) {
		serveTabDevTools(w, r)
	})
	mux.HandleFunc("/v1/tabs/devtools-panel", func(w http.ResponseWriter, r *http.Request) {
		serveTabDevToolsPanel(w, r)
	})
	mux.HandleFunc("/v1/tabs/windows", func(w http.ResponseWriter, r *http.Request) {
		serveTabWindows(w, r)
	})
	mux.HandleFunc("/v1/tabs/window-state", func(w http.ResponseWriter, r *http.Request) {
		serveTabWindowState(w, r)
	})
	mux.HandleFunc("/v1/tabs/window-bounds", func(w http.ResponseWriter, r *http.Request) {
		serveTabWindowBounds(w, r)
	})
	mux.HandleFunc("/v1/tabs/set-title", func(w http.ResponseWriter, r *http.Request) {
		serveTabSetTitle(w, r)
	})
	mux.HandleFunc("/v1/tabs/emulate-device", func(w http.ResponseWriter, r *http.Request) {
		serveTabEmulateDevice(w, r)
	})
	mux.HandleFunc("/v1/tabs/open", func(w http.ResponseWriter, r *http.Request) {
		serveTabAction(w, r, func(client tabClient, request tabActionRequest) (any, error) {
			return client.Open(request.URL)
		})
	})
	mux.HandleFunc("/v1/tabs/open-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabAction(w, r, func(client tabClient, request tabActionRequest) (any, error) {
			return client.OpenWindow(request.URL)
		})
	})
	mux.HandleFunc("/v1/tabs/open-in-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOpenInWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/move-to-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabMoveToWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/move-to-new-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabMoveToNewWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/merge-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabMergeWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/open-devtools", func(w http.ResponseWriter, r *http.Request) {
		serveTabAction(w, r, func(client tabClient, request tabActionRequest) (any, error) {
			return client.OpenDevToolsWindow(request.ID)
		})
	})
	mux.HandleFunc("/v1/tabs/open-devtools-in-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOpenDevToolsInWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/open-devtools-panel", func(w http.ResponseWriter, r *http.Request) {
		serveTabOpenDevToolsPanel(w, r)
	})
	mux.HandleFunc("/v1/tabs/open-devtools-panel-in-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOpenDevToolsPanelInWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/open-devtools-window-to-windows", func(w http.ResponseWriter, r *http.Request) {
		serveTabOpenDevToolsWindowToWindows(w, r)
	})
	mux.HandleFunc("/v1/tabs/open-devtools-panel-window-to-windows", func(w http.ResponseWriter, r *http.Request) {
		serveTabOpenDevToolsPanelWindowToWindows(w, r)
	})
	mux.HandleFunc("/v1/tabs/open-devtools-window-into-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOpenDevToolsWindowIntoWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/open-devtools-panel-window-into-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabOpenDevToolsPanelWindowIntoWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/close-duplicates", func(w http.ResponseWriter, r *http.Request) {
		serveTabCloseDuplicates(w, r)
	})
	mux.HandleFunc("/v1/tabs/activate-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabActivateWindow(w, r)
	})
	mux.HandleFunc("/v1/tabs/close-window", func(w http.ResponseWriter, r *http.Request) {
		serveTabCloseWindow(w, r)
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
	mux.HandleFunc("/v1/mirror/scan", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveMirrorScan(w, r)
	}, http.MethodPost))
	mux.HandleFunc("/v1/import/chrome", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveChromeImport(w, r)
	}, http.MethodPost))
	mux.HandleFunc("/v1/import/safari", requireMethods(func(w http.ResponseWriter, r *http.Request) {
		serveSafariImport(w, r)
	}, http.MethodPost))
	return mux
}
