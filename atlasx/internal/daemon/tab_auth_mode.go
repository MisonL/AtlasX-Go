package daemon

import (
	"errors"
	"fmt"
	"net/http"

	"atlasx/internal/tabs"
)

func serveTabAuthMode(w http.ResponseWriter, r *http.Request) {
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

	view, err := client.AuthMode(targetID)
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

	writeJSON(w, http.StatusOK, view)
}
