package daemon

import (
	"errors"
	"fmt"
	"net/http"

	"atlasx/internal/tabs"
)

func serveTabSemanticContext(w http.ResponseWriter, r *http.Request) {
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

	context, err := client.CaptureSemanticContext(targetID)
	if err != nil {
		var captureErr *tabs.SemanticCaptureError
		if errors.As(err, &captureErr) {
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"id":                captureErr.Context.ID,
				"title":             captureErr.Context.Title,
				"url":               captureErr.Context.URL,
				"captured_at":       captureErr.Context.CapturedAt,
				"returned":          captureErr.Context.Returned,
				"headings_returned": captureErr.Context.HeadingsReturned,
				"links_returned":    captureErr.Context.LinksReturned,
				"forms_returned":    captureErr.Context.FormsReturned,
				"headings":          captureErr.Context.Headings,
				"links":             captureErr.Context.Links,
				"forms":             captureErr.Context.Forms,
				"capture_error":     captureErr.Context.CaptureError,
				"error":             captureErr.Error(),
			})
			return
		}
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, context)
}
