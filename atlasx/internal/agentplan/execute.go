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
	ErrStepActionRequired   = errors.New("agent execute action is required for this step kind")
	ErrStepTabIDRequired    = errors.New("agent execute related_tab step is missing tab_id")
	ErrStepSnippetRequired  = errors.New("agent execute memory_snippet step is missing snippet")
)

type ExecutionActions struct {
	ActivateTab func(string) error
}

type ExecutionResult struct {
	TabID           string `json:"tab_id"`
	StepID          string `json:"step_id"`
	StepKind        string `json:"step_kind"`
	StepTitle       string `json:"step_title"`
	ActivatedTabID  string `json:"activated_tab_id,omitempty"`
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

func Execute(config sidebar.Config, context tabs.PageContext, memorySnippets []string, plan Plan, stepID string, confirmed bool, actions ExecutionActions) (ExecutionResult, error) {
	if !confirmed {
		return ExecutionResult{}, ErrConfirmationRequired
	}

	step, err := findStep(plan, stepID)
	if err != nil {
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
		return executeSummarize(config, context, memorySnippets, result)
	case "sidebar_ask":
		return executeAsk(config, context, memorySnippets, step, result)
	case "memory_snippet":
		return executeMemorySnippet(config, context, memorySnippets, step, result)
	case "related_tab":
		return executeRelatedTab(actions, step, result)
	default:
		return ExecutionResult{}, fmt.Errorf("%w: %s (%s)", ErrStepNotExecutable, step.ID, step.Kind)
	}
}

func executeSummarize(config sidebar.Config, context tabs.PageContext, memorySnippets []string, result ExecutionResult) (ExecutionResult, error) {
	if err := config.Validate(""); err != nil {
		return ExecutionResult{}, err
	}

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
}

func executeAsk(config sidebar.Config, context tabs.PageContext, memorySnippets []string, step Step, result ExecutionResult) (ExecutionResult, error) {
	if err := config.Validate(""); err != nil {
		return ExecutionResult{}, err
	}
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
}

func executeMemorySnippet(config sidebar.Config, context tabs.PageContext, memorySnippets []string, step Step, result ExecutionResult) (ExecutionResult, error) {
	if err := config.Validate(""); err != nil {
		return ExecutionResult{}, err
	}
	if strings.TrimSpace(step.Snippet) == "" {
		return ExecutionResult{}, fmt.Errorf("%w: %s", ErrStepSnippetRequired, step.ID)
	}

	question := fmt.Sprintf(
		"Using this memory snippet, explain how it relates to the current page and suggest the next safe action.\n\nMemory snippet:\n%s",
		step.Snippet,
	)
	response, err := config.AskWithMemory(sidebar.AskRequest{
		TabID:    context.ID,
		Question: question,
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
}

func executeRelatedTab(actions ExecutionActions, step Step, result ExecutionResult) (ExecutionResult, error) {
	if strings.TrimSpace(step.TabID) == "" {
		return ExecutionResult{}, fmt.Errorf("%w: %s", ErrStepTabIDRequired, step.ID)
	}
	if actions.ActivateTab == nil {
		return ExecutionResult{}, fmt.Errorf("%w: %s", ErrStepActionRequired, step.Kind)
	}
	if err := actions.ActivateTab(step.TabID); err != nil {
		return ExecutionResult{}, err
	}

	result.Executed = true
	result.ActivatedTabID = step.TabID
	result.Result = fmt.Sprintf("activated related tab %s", step.TabID)
	result.ContextSummary = step.URL
	result.Rollback = "manual_reactivate_previous_tab_if_needed"
	return result, nil
}

func findStep(plan Plan, stepID string) (Step, error) {
	for _, step := range plan.Steps {
		if step.ID == stepID {
			return step, nil
		}
	}
	return Step{}, fmt.Errorf("%w: %s", ErrStepNotFound, stepID)
}
