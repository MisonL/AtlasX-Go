package agentplan

import (
	"errors"
	"fmt"
	"strings"

	"atlasx/internal/sidebar"
	"atlasx/internal/tabs"
)

var (
	ErrConfirmationRequired = errors.New("agent execute requires explicit confirmation")
	ErrStepNotFound         = errors.New("agent plan step was not found")
	ErrStepNotExecutable    = errors.New("agent plan step is preview-only and cannot be executed")
)

type ExecutionResult struct {
	TabID           string `json:"tab_id"`
	StepID          string `json:"step_id"`
	StepKind        string `json:"step_kind"`
	StepTitle       string `json:"step_title"`
	Executed        bool   `json:"executed"`
	Confirmed       bool   `json:"confirmed"`
	TraceID         string `json:"trace_id"`
	Provider        string `json:"provider,omitempty"`
	Model           string `json:"model,omitempty"`
	Result          string `json:"result,omitempty"`
	ContextSummary  string `json:"context_summary,omitempty"`
	MemoryPersisted bool   `json:"memory_persisted"`
	Rollback        string `json:"rollback"`
}

func Execute(config sidebar.Config, context tabs.PageContext, memorySnippets []string, plan Plan, stepID string, confirmed bool) (ExecutionResult, error) {
	if !confirmed {
		return ExecutionResult{}, ErrConfirmationRequired
	}

	step, err := findStep(plan, stepID)
	if err != nil {
		return ExecutionResult{}, err
	}

	if err := config.Validate(""); err != nil {
		return ExecutionResult{}, err
	}

	result := ExecutionResult{
		TabID:           context.ID,
		StepID:          step.ID,
		StepKind:        step.Kind,
		StepTitle:       step.Title,
		Confirmed:       true,
		MemoryPersisted: false,
		Rollback:        "not_required_no_memory_persisted",
	}

	switch step.Kind {
	case "sidebar_summarize":
		response, err := config.SummarizeWithMemory(sidebar.SummarizeRequest{
			TabID: context.ID,
		}, context, memorySnippets)
		if err != nil {
			return ExecutionResult{}, err
		}
		result.Executed = true
		result.Provider = response.Provider
		result.Model = response.Model
		result.Result = response.Summary
		result.ContextSummary = response.ContextSummary
		return result, nil
	case "sidebar_ask":
		if strings.TrimSpace(step.Prompt) == "" {
			return ExecutionResult{}, fmt.Errorf("agent executable step %s has empty prompt", step.ID)
		}
		response, err := config.AskWithMemory(sidebar.AskRequest{
			TabID:    context.ID,
			Question: step.Prompt,
		}, context, memorySnippets)
		if err != nil {
			return ExecutionResult{}, err
		}
		result.Executed = true
		result.Provider = response.Provider
		result.Model = response.Model
		result.Result = response.Answer
		result.ContextSummary = response.ContextSummary
		return result, nil
	default:
		return ExecutionResult{}, fmt.Errorf("%w: %s (%s)", ErrStepNotExecutable, step.ID, step.Kind)
	}
}

func findStep(plan Plan, stepID string) (Step, error) {
	for _, step := range plan.Steps {
		if step.ID == stepID {
			return step, nil
		}
	}
	return Step{}, fmt.Errorf("%w: %s", ErrStepNotFound, stepID)
}
