package daemon

import (
	"errors"
	"fmt"
	"net/http"

	"atlasx/internal/tabs"
)

func serveTabSelection(w http.ResponseWriter, r *http.Request) {
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

	selection, err := client.CaptureSelection(targetID)
	if err != nil {
		var captureErr *tabs.SelectionCaptureError
		if errors.As(err, &captureErr) {
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"id":                       captureErr.Context.ID,
				"title":                    captureErr.Context.Title,
				"url":                      captureErr.Context.URL,
				"selection_text":           captureErr.Context.SelectionText,
				"captured_at":              captureErr.Context.CapturedAt,
				"selection_present":        captureErr.Context.SelectionPresent,
				"selection_text_truncated": captureErr.Context.SelectionTextTruncated,
				"selection_text_length":    captureErr.Context.SelectionTextLength,
				"selection_text_limit":     captureErr.Context.SelectionTextLimit,
				"capture_error":            captureErr.Context.CaptureError,
				"error":                    captureErr.Error(),
			})
			return
		}
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, selection)
}
