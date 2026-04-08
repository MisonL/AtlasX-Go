package daemon

import (
	"errors"
	"fmt"
	"net/http"

	"atlasx/internal/agentplan"
	"atlasx/internal/sidebar"
)

type tabAgentExecuteRequest struct {
	ID       string   `json:"id"`
	StepID   string   `json:"step_id,omitempty"`
	StepIDs  []string `json:"step_ids,omitempty"`
	MaxSteps int      `json:"max_steps,omitempty"`
	Confirm  bool     `json:"confirm"`
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
	if request.ID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("id is required"))
		return
	}
	if !request.Confirm {
		writeError(w, http.StatusBadRequest, agentplan.ErrConfirmationRequired)
		return
	}
	stepIDs, err := collectRequestedStepIDs(request)
	if err != nil {
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

	if len(stepIDs) == 1 {
		result, err := agentplan.Execute(config, context, memorySnippets, plan, stepIDs[0], request.Confirm, agentplan.ExecutionActions{
			ActivateTab: client.Activate,
		})
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
		return
	}

	batch, err := agentplan.ExecuteBatch(config, context, memorySnippets, plan, stepIDs, request.Confirm, agentplan.ExecutionActions{
		ActivateTab: client.Activate,
	}, request.MaxSteps)
	batch.TraceID = traceID
	for index := range batch.Results {
		batch.Results[index].TraceID = traceID
	}
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		writeJSON(w, agentExecuteStatusCode(err), batch)
		return
	}
	if err := sidebar.SaveRuntimeResult(paths, traceID, nil); err != nil {
		writeSidebarAskError(w, http.StatusInternalServerError, traceID, err)
		return
	}
	writeJSON(w, http.StatusOK, batch)
}

func writeAgentExecuteError(w http.ResponseWriter, traceID string, err error) {
	writeSidebarAskError(w, agentExecuteStatusCode(err), traceID, err)
}

func agentExecuteStatusCode(err error) int {
	switch {
	case errors.Is(err, agentplan.ErrConfirmationRequired),
		errors.Is(err, agentplan.ErrStepSequenceRequired),
		errors.Is(err, agentplan.ErrStepBatchExceeded),
		errors.Is(err, agentplan.ErrStepNotExecutable),
		errors.Is(err, agentplan.ErrStepActionRequired),
		errors.Is(err, agentplan.ErrStepTabIDRequired),
		errors.Is(err, agentplan.ErrStepSnippetRequired),
		errors.Is(err, sidebar.ErrProviderNotFound),
		errors.Is(err, sidebar.ErrTokenBudgetExceeded):
		return http.StatusBadRequest
	case errors.Is(err, agentplan.ErrStepNotFound):
		return http.StatusNotFound
	case errors.Is(err, sidebar.ErrNotConfigured):
		return http.StatusServiceUnavailable
	case errors.Is(err, sidebar.ErrBackendNotImplemented):
		return http.StatusNotImplemented
	case errors.Is(err, sidebar.ErrProviderFailed):
		return http.StatusBadGateway
	default:
		return http.StatusBadRequest
	}
}

func collectRequestedStepIDs(request tabAgentExecuteRequest) ([]string, error) {
	if request.StepID != "" && len(request.StepIDs) > 0 {
		return nil, fmt.Errorf("step_id and step_ids are mutually exclusive")
	}
	if request.StepID != "" {
		return []string{request.StepID}, nil
	}
	if len(request.StepIDs) == 0 {
		return nil, fmt.Errorf("step_id or step_ids is required")
	}
	return request.StepIDs, nil
}
