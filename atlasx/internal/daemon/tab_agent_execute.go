package daemon

import (
	"errors"
	"fmt"
	"net/http"

	"atlasx/internal/agentplan"
	"atlasx/internal/sidebar"
)

type tabAgentExecuteRequest struct {
	ID      string `json:"id"`
	StepID  string `json:"step_id"`
	Confirm bool   `json:"confirm"`
}

func serveTabAgentExecute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request tabAgentExecuteRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if request.ID == "" || request.StepID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("id and step_id are required"))
		return
	}
	if !request.Confirm {
		writeError(w, http.StatusBadRequest, agentplan.ErrConfirmationRequired)
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

	client, err := newTabsClient(paths)
	if err != nil {
		writeSidebarAskError(w, http.StatusConflict, traceID, err)
		return
	}

	context, memorySnippets, plan, err := loadDaemonAgentPlan(paths, client, request.ID)
	if err != nil {
		handleAgentPlanLoadError(w, err)
		return
	}

	result, err := agentplan.Execute(config, context, memorySnippets, plan, request.StepID, request.Confirm)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeAgentExecuteError(w, traceID, err)
		return
	}
	result.TraceID = traceID
	if err := sidebar.SaveRuntimeResult(paths, traceID, nil); err != nil {
		writeSidebarAskError(w, http.StatusInternalServerError, traceID, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeAgentExecuteError(w http.ResponseWriter, traceID string, err error) {
	switch {
	case errors.Is(err, agentplan.ErrConfirmationRequired),
		errors.Is(err, agentplan.ErrStepNotExecutable),
		errors.Is(err, sidebar.ErrProviderNotFound),
		errors.Is(err, sidebar.ErrTokenBudgetExceeded):
		writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
	case errors.Is(err, agentplan.ErrStepNotFound):
		writeSidebarAskError(w, http.StatusNotFound, traceID, err)
	case errors.Is(err, sidebar.ErrNotConfigured):
		writeSidebarAskError(w, http.StatusServiceUnavailable, traceID, err)
	case errors.Is(err, sidebar.ErrBackendNotImplemented):
		writeSidebarAskError(w, http.StatusNotImplemented, traceID, err)
	case errors.Is(err, sidebar.ErrProviderFailed):
		writeSidebarAskError(w, http.StatusBadGateway, traceID, err)
	default:
		writeSidebarAskError(w, http.StatusBadRequest, traceID, err)
	}
}
