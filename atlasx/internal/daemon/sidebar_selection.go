package daemon

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"atlasx/internal/memory"
	"atlasx/internal/sidebar"
	"atlasx/internal/tabs"
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

	selectionText, err := resolveSelectionText(client, request)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		var captureErr *tabs.SelectionCaptureError
		if errors.As(err, &captureErr) {
			writeSidebarAskError(w, http.StatusBadGateway, traceID, err)
			return
		}
		writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		return
	}

	context, err := client.Capture(request.TabID)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusBadGateway, traceID, err)
		return
	}

	selectionQuestion, err := sidebar.BuildSelectionQuestion(selectionText, request.Question)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
		return
	}

	memorySnippets, err := memory.FindRelevantSnippetsForPage(paths, memory.RetrievalInput{
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

	request.SelectionText = selectionText
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
	if _, err := memory.AppendQATurnControlled(paths, memory.QATurnInput{
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

func resolveSelectionText(client tabClient, request sidebar.SelectionAskRequest) (string, error) {
	if strings.TrimSpace(request.SelectionText) != "" {
		return request.SelectionText, nil
	}

	selection, err := client.CaptureSelection(request.TabID)
	if err != nil {
		return "", err
	}
	if !selection.SelectionPresent || strings.TrimSpace(selection.SelectionText) == "" {
		return "", errors.New("selection_text is required when page selection is empty")
	}
	return selection.SelectionText, nil
}
