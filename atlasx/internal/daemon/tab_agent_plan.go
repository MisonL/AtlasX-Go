package daemon

import (
	"errors"
	"fmt"
	"net/http"

	"atlasx/internal/agentplan"
	"atlasx/internal/contextrec"
	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/suggestions"
	"atlasx/internal/tabs"
)

func serveTabAgentPlan(w http.ResponseWriter, r *http.Request) {
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

	_, _, plan, err := loadDaemonAgentPlan(paths, client, targetID)
	if err != nil {
		handleAgentPlanLoadError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, plan)
}

func loadDaemonAgentPlan(paths macos.Paths, client tabClient, targetID string) (tabs.PageContext, []string, agentplan.Plan, error) {
	context, err := client.Capture(targetID)
	if err != nil {
		return tabs.PageContext{}, nil, agentplan.Plan{}, err
	}

	targets, err := client.List()
	if err != nil {
		return tabs.PageContext{}, nil, agentplan.Plan{}, err
	}

	memorySnippets, err := memory.FindRelevantSnippets(paths, memory.RetrievalInput{
		TabID:    context.ID,
		Title:    context.Title,
		URL:      context.URL,
		Question: buildTabSuggestionsQuery(context),
	})
	if err != nil {
		return tabs.PageContext{}, nil, agentplan.Plan{}, err
	}

	plan := agentplan.Build(
		context,
		suggestions.ForPage(context, memorySnippets),
		contextrec.ForPage(context, targets, memorySnippets),
		len(memorySnippets),
	)
	return context, memorySnippets, plan, nil
}

func handleAgentPlanLoadError(w http.ResponseWriter, err error) {
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
	var retrievalErr *tabs.CaptureError
	if errors.As(err, &retrievalErr) {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeError(w, http.StatusBadGateway, err)
}
