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
	ErrStepSequenceRequired = errors.New("agent execute requires at least one step")
	ErrStepBatchExceeded    = errors.New("agent execute requested step count exceeds max_steps")
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

type BatchExecutionResult struct {
	TabID           string            `json:"tab_id"`
	Requested       int               `json:"requested"`
	Executed        int               `json:"executed"`
	Stopped         bool              `json:"stopped"`
	FailedStepID    string            `json:"failed_step_id,omitempty"`
	Failure         string            `json:"failure,omitempty"`
	MaxSteps        int               `json:"max_steps"`
	TraceID         string            `json:"trace_id"`
	MemoryPersisted bool              `json:"memory_persisted"`
	Rollback        string            `json:"rollback"`
	Results         []ExecutionResult `json:"results"`
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

func ExecuteBatch(config sidebar.Config, context tabs.PageContext, memorySnippets []string, plan Plan, stepIDs []string, confirmed bool, actions ExecutionActions, maxSteps int) (BatchExecutionResult, error) {
	if !confirmed {
		return BatchExecutionResult{}, ErrConfirmationRequired
	}
	if len(stepIDs) == 0 {
		return BatchExecutionResult{}, ErrStepSequenceRequired
	}
	if maxSteps <= 0 {
		maxSteps = 1
	}
	if len(stepIDs) > maxSteps {
		err := fmt.Errorf("%w: requested=%d max_steps=%d", ErrStepBatchExceeded, len(stepIDs), maxSteps)
		return BatchExecutionResult{
			TabID:           context.ID,
			Requested:       len(stepIDs),
			Executed:        0,
			Stopped:         true,
			Failure:         err.Error(),
			MaxSteps:        maxSteps,
			MemoryPersisted: false,
			Rollback:        "not_required_no_steps_executed",
			Results:         nil,
		}, err
	}

	results := make([]ExecutionResult, 0, len(stepIDs))
	for _, stepID := range stepIDs {
		result, err := Execute(config, context, memorySnippets, plan, stepID, confirmed, actions)
		if err != nil {
			return BatchExecutionResult{
				TabID:           context.ID,
				Requested:       len(stepIDs),
				Executed:        len(results),
				Stopped:         true,
				FailedStepID:    stepID,
				Failure:         err.Error(),
				MaxSteps:        maxSteps,
				MemoryPersisted: false,
				Rollback:        deriveBatchRollback(results),
				Results:         results,
			}, err
		}
		results = append(results, result)
	}

	return BatchExecutionResult{
		TabID:           context.ID,
		Requested:       len(stepIDs),
		Executed:        len(results),
		Stopped:         false,
		MaxSteps:        maxSteps,
		MemoryPersisted: false,
		Rollback:        deriveBatchRollback(results),
		Results:         results,
	}, nil
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

func deriveBatchRollback(results []ExecutionResult) string {
	if len(results) == 0 {
		return "not_required_no_steps_executed"
	}
	if len(results) == 1 {
		return results[0].Rollback
	}

	rollbacks := make([]string, 0, len(results))
	seen := make(map[string]struct{}, len(results))
	for _, result := range results {
		key := strings.TrimSpace(result.Rollback)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		rollbacks = append(rollbacks, key)
	}
	if len(rollbacks) == 0 {
		return "not_required_no_memory_persisted"
	}
	return strings.Join(rollbacks, ";")
}
