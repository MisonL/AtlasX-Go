package daemon

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"atlasx/internal/imports"
	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
	"atlasx/internal/tabs"
)

var discoverPaths = macos.DiscoverPaths
var newTabsClient = func(paths macos.Paths) (tabClient, error) {
	return tabs.New(paths)
}

type tabClient interface {
	List() ([]tabs.Target, error)
	Open(string) (tabs.Target, error)
	Activate(string) error
	Close(string) error
	Navigate(string, string) error
	Capture(string) (tabs.PageContext, error)
}

type dataLoader[T any] func(macos.Paths) ([]T, error)
type indexedURLResolver func(macos.Paths, int) (string, error)
type tabAction func(client tabClient, request tabActionRequest) (any, error)

type tabActionRequest struct {
	ID  string `json:"id"`
	URL string `json:"url"`
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

	config, err := loadSidebarConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, config.Status())
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

	config, err := loadSidebarConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := config.Ask(request); err != nil {
		switch {
		case errors.Is(err, sidebar.ErrNotConfigured):
			writeError(w, http.StatusServiceUnavailable, err)
		case errors.Is(err, sidebar.ErrBackendNotImplemented):
			writeError(w, http.StatusNotImplemented, err)
		default:
			writeError(w, http.StatusBadRequest, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
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

	snapshot, err := mirror.Scan(paths, profileDir)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, snapshot)
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

	report, err := imports.ImportChrome(paths, sourceProfileDir)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func loadSidebarConfig() (sidebar.Config, error) {
	paths, err := discoverPaths()
	if err != nil {
		return sidebar.Config{}, err
	}
	config, err := settings.NewStore(paths.ConfigFile).Load()
	if err != nil {
		return sidebar.Config{}, err
	}
	return sidebar.FromSettings(config), nil
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
