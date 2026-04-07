package daemon

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"atlasx/internal/memory"
	"atlasx/internal/sidebar"
)

func serveSidebarSelectionAsk(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request sidebar.SelectionAskRequest
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

	selectionQuestion, err := sidebar.BuildSelectionQuestion(request.SelectionText, request.Question)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
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
		Question: selectionQuestion,
	})
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusInternalServerError, traceID, err)
		return
	}

	response, err := config.AskSelectionWithMemory(request, context, memorySnippets)
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
		Question:   selectionQuestion,
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
