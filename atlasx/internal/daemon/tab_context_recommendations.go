package daemon

import (
	"errors"
	"fmt"
	"net/http"

	"atlasx/internal/contextrec"
	"atlasx/internal/memory"
	"atlasx/internal/tabs"
)

type tabContextRecommendationsResponse struct {
	ID              string                      `json:"id"`
	Title           string                      `json:"title"`
	URL             string                      `json:"url"`
	CapturedAt      string                      `json:"captured_at"`
	Returned        int                         `json:"returned"`
	MemoryReturned  int                         `json:"memory_returned"`
	Recommendations []contextrec.Recommendation `json:"recommendations"`
}

func serveTabContextRecommendations(w http.ResponseWriter, r *http.Request) {
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

	targets, err := client.List()
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	memorySnippets, err := memory.FindRelevantSnippetsForPage(paths, memory.RetrievalInput{
		TabID:    context.ID,
		Title:    context.Title,
		URL:      context.URL,
		Question: buildTabSuggestionsQuery(context),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	recommendations := contextrec.ForPage(context, targets, memorySnippets)
	writeJSON(w, http.StatusOK, tabContextRecommendationsResponse{
		ID:              context.ID,
		Title:           context.Title,
		URL:             context.URL,
		CapturedAt:      context.CapturedAt,
		Returned:        len(recommendations),
		MemoryReturned:  len(memorySnippets),
		Recommendations: recommendations,
	})
}
