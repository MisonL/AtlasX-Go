package daemon

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"atlasx/internal/imports"
	"atlasx/internal/memory"
	"atlasx/internal/mirror"
	"atlasx/internal/openurl"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
	"atlasx/internal/sourcepaths"
	"atlasx/internal/tabs"
)

var discoverPaths = macos.DiscoverPaths
var newTabsClient = func(paths macos.Paths) (tabClient, error) {
	return tabs.New(paths)
}

type tabClient interface {
	List() ([]tabs.Target, error)
	Search(string) ([]tabs.Target, error)
	Windows() ([]tabs.WindowSummary, error)
	CloseDuplicates() (tabs.CloseDuplicatesResult, error)
	OpenInWindow(int, string) (tabs.WindowOpenResult, error)
	MoveToWindow(string, int) (tabs.WindowMoveResult, error)
	MoveToNewWindow(string) (tabs.WindowMoveToNewResult, error)
	MergeWindow(int, int) (tabs.WindowMergeResult, error)
	ActivateWindow(int) (tabs.WindowActivateResult, error)
	CloseWindow(int) (tabs.WindowCloseResult, error)
	SetWindowState(int, string) (tabs.WindowBounds, error)
	SetWindowBounds(int, int, int, int, int) (tabs.WindowBounds, error)
	OpenDevToolsWindow(string) (tabs.Target, error)
	OpenDevToolsInWindow(string, int) (tabs.WindowOpenResult, error)
	OpenDevToolsPanelWindow(string, string) (tabs.Target, error)
	OpenDevToolsPanelInWindow(string, string, int) (tabs.WindowOpenResult, error)
	OpenDevToolsWindowToWindows(int) (tabs.DevToolsWindowToWindowsResult, error)
	OpenDevToolsPanelWindowIntoWindow(int, string, int) (tabs.DevToolsPanelWindowOpenResult, error)
	OpenDevToolsWindowIntoWindow(int, int) (tabs.DevToolsWindowOpenResult, error)
	Open(string) (tabs.Target, error)
	OpenWindow(string) (tabs.Target, error)
	Activate(string) error
	Close(string) error
	Navigate(string, string) error
	Capture(string) (tabs.PageContext, error)
	CaptureSemanticContext(string) (tabs.SemanticContext, error)
	CaptureSelection(string) (tabs.SelectionContext, error)
	DevTools(string) (tabs.DevToolsTarget, error)
	DevToolsPanel(string, string) (tabs.DevToolsTarget, error)
	EmulateDevice(string, string) (tabs.DeviceEmulationResult, error)
}

type dataLoader[T any] func(macos.Paths) ([]T, error)
type indexedURLResolver func(macos.Paths, int) (string, error)
type tabAction func(client tabClient, request tabActionRequest) (any, error)

type tabActionRequest struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Panel string `json:"panel"`
}

type tabOpenDevToolsPanelInWindowRequest struct {
	ID       string `json:"id"`
	Panel    string `json:"panel"`
	WindowID int    `json:"window_id"`
}

type tabOpenDevToolsInWindowRequest struct {
	ID       string `json:"id"`
	WindowID int    `json:"window_id"`
}

type tabOpenDevToolsPanelWindowIntoWindowRequest struct {
	SourceWindowID int    `json:"source_window_id"`
	Panel          string `json:"panel"`
	TargetWindowID int    `json:"target_window_id"`
}

type tabOpenDevToolsWindowToWindowsRequest struct {
	SourceWindowID int `json:"source_window_id"`
}

type tabEmulationRequest struct {
	ID     string `json:"id"`
	Preset string `json:"preset"`
}

type windowStateRequest struct {
	WindowID int    `json:"window_id"`
	State    string `json:"state"`
}

type windowBoundsRequest struct {
	WindowID int `json:"window_id"`
	Left     int `json:"left"`
	Top      int `json:"top"`
	Width    int `json:"width"`
	Height   int `json:"height"`
}

type windowIDRequest struct {
	WindowID int `json:"window_id"`
}

type windowOpenRequest struct {
	WindowID int    `json:"window_id"`
	URL      string `json:"url"`
}

type windowMoveRequest struct {
	ID       string `json:"id"`
	WindowID int    `json:"window_id"`
}

type targetIDRequest struct {
	ID string `json:"id"`
}

type windowMergeRequest struct {
	SourceWindowID int `json:"source_window_id"`
	TargetWindowID int `json:"target_window_id"`
}

type openIndexRequest struct {
	Index int `json:"index"`
}

type mirrorScanRequest struct {
	ProfileDir string `json:"profile_dir"`
}

type chromeImportRequest struct {
	SourceProfileDir string `json:"source_profile_dir"`
}

func serveSidebarStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	config, err := loadSidebarConfig(paths)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	status, err := config.StatusWithRuntime(paths)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func serveSidebarAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request sidebar.AskRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	traceID := sidebar.NewTraceID()

	paths, err := discoverPaths()
	if err != nil {
		writeSidebarAskError(w, http.StatusInternalServerError, traceID, err)
		return
	}

	config, err := loadSidebarConfig(paths)
	if err != nil {
		writeSidebarAskError(w, http.StatusInternalServerError, traceID, err)
		return
	}

	if err := config.Validate(request.ProviderID); err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		switch {
		case errors.Is(err, sidebar.ErrNotConfigured):
			writeSidebarAskError(w, http.StatusServiceUnavailable, traceID, err)
		case errors.Is(err, sidebar.ErrBackendNotImplemented):
			writeSidebarAskError(w, http.StatusNotImplemented, traceID, err)
		case errors.Is(err, sidebar.ErrProviderNotFound):
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		case errors.Is(err, sidebar.ErrTokenBudgetExceeded):
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		default:
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		}
		return
	}

	client, err := newTabsClient(paths)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusConflict, traceID, err)
		return
	}

	context, err := client.Capture(request.TabID)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusBadGateway, traceID, err)
		return
	}

	memorySnippets, err := memory.FindRelevantSnippets(paths, memory.RetrievalInput{
		TabID:    context.ID,
		Title:    context.Title,
		URL:      context.URL,
		Question: request.Question,
	})
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusInternalServerError, traceID, err)
		return
	}

	response, err := config.AskWithMemory(request, context, memorySnippets)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		switch {
		case errors.Is(err, sidebar.ErrNotConfigured):
			writeSidebarAskError(w, http.StatusServiceUnavailable, traceID, err)
		case errors.Is(err, sidebar.ErrBackendNotImplemented):
			writeSidebarAskError(w, http.StatusNotImplemented, traceID, err)
		case errors.Is(err, sidebar.ErrProviderNotFound):
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		case errors.Is(err, sidebar.ErrTokenBudgetExceeded):
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		case errors.Is(err, sidebar.ErrProviderFailed):
			writeSidebarAskError(w, http.StatusBadGateway, traceID, err)
		default:
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		}
		return
	}

	response.TraceID = traceID
	if err := memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: time.Now().UTC().Format(time.RFC3339Nano),
		TabID:      context.ID,
		Title:      context.Title,
		URL:        context.URL,
		Question:   request.Question,
		Answer:     response.Answer,
		CitedURLs:  nonEmptyURLs(context.URL),
		TraceID:    traceID,
	}); err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusInternalServerError, traceID, err)
		return
	}
	_ = sidebar.SaveRuntimeResult(paths, traceID, nil)
	writeJSON(w, http.StatusOK, response)
}

func serveSidebarSummarize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request sidebar.SummarizeRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	traceID := sidebar.NewTraceID()

	paths, err := discoverPaths()
	if err != nil {
		writeSidebarAskError(w, http.StatusInternalServerError, traceID, err)
		return
	}

	config, err := loadSidebarConfig(paths)
	if err != nil {
		writeSidebarAskError(w, http.StatusInternalServerError, traceID, err)
		return
	}

	if err := config.Validate(request.ProviderID); err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		switch {
		case errors.Is(err, sidebar.ErrNotConfigured):
			writeSidebarAskError(w, http.StatusServiceUnavailable, traceID, err)
		case errors.Is(err, sidebar.ErrBackendNotImplemented):
			writeSidebarAskError(w, http.StatusNotImplemented, traceID, err)
		case errors.Is(err, sidebar.ErrProviderNotFound):
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		case errors.Is(err, sidebar.ErrTokenBudgetExceeded):
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		default:
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		}
		return
	}

	client, err := newTabsClient(paths)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusConflict, traceID, err)
		return
	}

	context, err := client.Capture(request.TabID)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusBadGateway, traceID, err)
		return
	}

	memorySnippets, err := memory.FindRelevantSnippets(paths, memory.RetrievalInput{
		TabID:    context.ID,
		Title:    context.Title,
		URL:      context.URL,
		Question: sidebar.PageSummaryQuestion,
	})
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusInternalServerError, traceID, err)
		return
	}

	response, err := config.SummarizeWithMemory(request, context, memorySnippets)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		switch {
		case errors.Is(err, sidebar.ErrNotConfigured):
			writeSidebarAskError(w, http.StatusServiceUnavailable, traceID, err)
		case errors.Is(err, sidebar.ErrBackendNotImplemented):
			writeSidebarAskError(w, http.StatusNotImplemented, traceID, err)
		case errors.Is(err, sidebar.ErrProviderNotFound):
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		case errors.Is(err, sidebar.ErrTokenBudgetExceeded):
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		case errors.Is(err, sidebar.ErrProviderFailed):
			writeSidebarAskError(w, http.StatusBadGateway, traceID, err)
		default:
			writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		}
		return
	}

	response.TraceID = traceID
	if err := memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: time.Now().UTC().Format(time.RFC3339Nano),
		TabID:      context.ID,
		Title:      context.Title,
		URL:        context.URL,
		Question:   sidebar.PageSummaryQuestion,
		Answer:     response.Summary,
		CitedURLs:  nonEmptyURLs(context.URL),
		TraceID:    traceID,
	}); err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusInternalServerError, traceID, err)
		return
	}
	_ = sidebar.SaveRuntimeResult(paths, traceID, nil)
	writeJSON(w, http.StatusOK, response)
}

func serveBrowserData[T any](w http.ResponseWriter, loader dataLoader[T]) {
	paths, err := discoverPaths()
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

	writeJSON(w, http.StatusOK, payload)
}

func serveBrowserOpenAction(w http.ResponseWriter, r *http.Request, resolver indexedURLResolver, successKey string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request openIndexRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if request.Index < 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("index must be >= 0"))
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	targetURL, err := resolver(paths, request.Index)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusConflict, err)
		return
	}

	client, err := newTabsClient(paths)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}

	target, err := client.Open(targetURL)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		successKey: request.Index,
		"id":       target.ID,
		"url":      target.URL,
	})
}

func serveTabsList(w http.ResponseWriter) {
	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	client, err := newTabsClient(paths)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}

	targets, err := client.List()
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, tabs.PageTargets(targets))
}

func serveTabContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	targetID := r.URL.Query().Get("id")
	if targetID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missing query parameter id"))
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	client, err := newTabsClient(paths)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}

	context, err := client.Capture(targetID)
	if err != nil {
		var captureErr *tabs.CaptureError
		if errors.As(err, &captureErr) {
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"id":             captureErr.Context.ID,
				"title":          captureErr.Context.Title,
				"url":            captureErr.Context.URL,
				"text":           captureErr.Context.Text,
				"captured_at":    captureErr.Context.CapturedAt,
				"text_truncated": captureErr.Context.TextTruncated,
				"text_length":    captureErr.Context.TextLength,
				"text_limit":     captureErr.Context.TextLimit,
				"capture_error":  captureErr.Context.CaptureError,
				"error":          captureErr.Error(),
			})
			return
		}
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if err := memory.AppendPageCapture(paths, memory.PageCaptureInput{
		OccurredAt: context.CapturedAt,
		TabID:      context.ID,
		Title:      context.Title,
		URL:        context.URL,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, context)
}

func serveTabAction(w http.ResponseWriter, r *http.Request, action tabAction) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request tabActionRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if request.URL != "" {
		validatedURL, validateErr := openurl.Validate(request.URL)
		if validateErr != nil {
			writeError(w, http.StatusBadRequest, validateErr)
			return
		}
		request.URL = validatedURL
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	client, err := newTabsClient(paths)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}

	payload, err := action(client, request)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, payload)
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

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	profileDir := request.ProfileDir
	if profileDir == "" {
		profileDir = mirror.DefaultProfilePath(paths)
	}
	if err := sourcepaths.ValidateMirrorProfileDir(paths, profileDir); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	snapshot, err := mirror.Scan(paths, profileDir)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, snapshot)
}

func decodePositiveIntQuery(r *http.Request, name string) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return 0, fmt.Errorf("missing query parameter %s", name)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid query parameter %s", name)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("query parameter %s must be positive", name)
	}
	return parsed, nil
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

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	sourceProfileDir := request.SourceProfileDir
	if sourceProfileDir == "" {
		sourceProfileDir = imports.DefaultChromeProfileDir(paths)
	}
	if err := sourcepaths.ValidateChromeImportSourceDir(paths, sourceProfileDir); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	report, err := imports.ImportChrome(paths, sourceProfileDir)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func loadSidebarConfig(paths macos.Paths) (sidebar.Config, error) {
	config, err := settings.NewStore(paths.ConfigFile).Load()
	if err != nil {
		return sidebar.Config{}, err
	}
	return sidebar.FromSettings(config), nil
}

func writeSidebarAskError(w http.ResponseWriter, code int, traceID string, err error) {
	writeJSON(w, code, map[string]string{
		"error":    err.Error(),
		"trace_id": traceID,
	})
}

func nonEmptyURLs(values ...string) []string {
	urls := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		urls = append(urls, value)
	}
	return urls
}

func serveSafariImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	report, err := imports.ImportSafari(paths)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, report)
}
