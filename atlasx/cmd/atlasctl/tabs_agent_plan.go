package main

import (
	"errors"
	"flag"
	"fmt"

	"atlasx/internal/agentplan"
	"atlasx/internal/contextrec"
	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
	"atlasx/internal/suggestions"
	"atlasx/internal/tabs"
)

func runTabsAgentPlan(paths macos.Paths, client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing target id for tabs agent-plan")
	}

	_, _, plan, err := loadCommandAgentPlan(paths, client, args[0])
	if err != nil {
		return err
	}
	printAgentPlan(plan)
	return nil
}

func runTabsAgentExecute(paths macos.Paths, client commandTabsClient, args []string) error {
	fs := flag.NewFlagSet("tabs agent-execute", flag.ContinueOnError)
	fs.SetOutput(discardCommandOutput{})

	confirm := fs.Bool("confirm", false, "confirm executing one or more agent plan steps")
	maxSteps := fs.Int("max-steps", 1, "maximum number of agent plan steps allowed in this request")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		return errors.New("missing target id or step id list for tabs agent-execute")
	}
	if !*confirm {
		return agentplan.ErrConfirmationRequired
	}

	tabID := fs.Arg(0)
	stepIDs := fs.Args()[1:]
	context, memorySnippets, plan, err := loadCommandAgentPlan(paths, client, tabID)
	if err != nil {
		return err
	}

	config, err := loadAgentSidebarConfig(paths)
	if err != nil {
		return err
	}

	traceID := sidebar.NewTraceID()
	if len(stepIDs) == 1 {
		result, err := agentplan.Execute(config, context, memorySnippets, plan, stepIDs[0], *confirm, agentplan.ExecutionActions{
			ActivateTab: client.Activate,
		})
		if err != nil {
			return finishSidebarCommand(paths, traceID, err)
		}
		result.TraceID = traceID
		if err := finishSidebarCommand(paths, traceID, nil); err != nil {
			return err
		}
		printAgentExecution(result)
		return nil
	}

	batch, err := agentplan.ExecuteBatch(config, context, memorySnippets, plan, stepIDs, *confirm, agentplan.ExecutionActions{
		ActivateTab: client.Activate,
	}, *maxSteps)
	batch.TraceID = traceID
	for index := range batch.Results {
		batch.Results[index].TraceID = traceID
	}
	if err != nil {
		_ = finishSidebarCommand(paths, traceID, err)
		printAgentBatchExecution(batch)
		return err
	}

	if err := finishSidebarCommand(paths, traceID, nil); err != nil {
		return err
	}
	printAgentBatchExecution(batch)
	return nil
}

func loadCommandAgentPlan(paths macos.Paths, client commandTabsClient, targetID string) (tabs.PageContext, []string, agentplan.Plan, error) {
	context, err := client.Capture(targetID)
	if err != nil {
		printPageContext(context)
		return tabs.PageContext{}, nil, agentplan.Plan{}, err
	}

	targets, err := client.List()
	if err != nil {
		return tabs.PageContext{}, nil, agentplan.Plan{}, err
	}

	memorySnippets, err := memory.FindRelevantSnippetsForPage(paths, memory.RetrievalInput{
		TabID:    context.ID,
		Title:    context.Title,
		URL:      context.URL,
		Question: buildPageSuggestionQuery(context),
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

func loadAgentSidebarConfig(paths macos.Paths) (sidebar.Config, error) {
	config, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return sidebar.Config{}, err
	}
	return sidebar.FromSettings(config), nil
}

func printAgentPlan(plan agentplan.Plan) {
	fmt.Printf(
		"id=%s title=%q url=%s captured_at=%s goal=%q returned=%d read_only=%t executed=%t memory_returned=%d suggestion_returned=%d recommendation_returned=%d rollback=%s\n",
		plan.ID,
		plan.Title,
		plan.URL,
		plan.CapturedAt,
		plan.Goal,
		plan.Returned,
		plan.ReadOnly,
		plan.Executed,
		plan.MemoryReturned,
		plan.SuggestionReturned,
		plan.RecommendationReturned,
		plan.Rollback,
	)
	for index, guardrail := range plan.Guardrails {
		fmt.Printf("guardrail[%d]=%s\n", index, guardrail)
	}
	for index, step := range plan.Steps {
		fmt.Printf(
			"index=%d step_id=%s kind=%s title=%q source=%s executable=%t execution_path=%s requires_provider=%t requires_confirmation=%t tab_id=%s url=%s\n",
			index,
			step.ID,
			step.Kind,
			step.Title,
			step.Source,
			step.Executable,
			step.ExecutionPath,
			step.RequiresProvider,
			step.RequiresConfirmation,
			step.TabID,
			step.URL,
		)
		fmt.Printf("reason=%q\n", step.Reason)
		fmt.Printf("prompt=%q\n", step.Prompt)
		fmt.Printf("snippet=%q\n", step.Snippet)
	}
}

func printAgentExecution(result agentplan.ExecutionResult) {
	fmt.Printf(
		"tab_id=%s step_id=%s step_kind=%s step_title=%q activated_tab_id=%s executed=%t confirmed=%t trace_id=%s provider=%s model=%s memory_persisted=%t rollback=%s\n",
		result.TabID,
		result.StepID,
		result.StepKind,
		result.StepTitle,
		result.ActivatedTabID,
		result.Executed,
		result.Confirmed,
		result.TraceID,
		result.Provider,
		result.Model,
		result.MemoryPersisted,
		result.Rollback,
	)
	fmt.Printf("result=%q\n", result.Result)
	fmt.Printf("context_summary=%q\n", result.ContextSummary)
}

func printAgentBatchExecution(batch agentplan.BatchExecutionResult) {
	fmt.Printf(
		"tab_id=%s requested=%d executed=%d stopped=%t failed_step_id=%s max_steps=%d trace_id=%s memory_persisted=%t rollback=%s\n",
		batch.TabID,
		batch.Requested,
		batch.Executed,
		batch.Stopped,
		batch.FailedStepID,
		batch.MaxSteps,
		batch.TraceID,
		batch.MemoryPersisted,
		batch.Rollback,
	)
	if batch.Failure != "" {
		fmt.Printf("failure=%q\n", batch.Failure)
	}
	for index, result := range batch.Results {
		fmt.Printf("batch_index=%d ", index)
		printAgentExecution(result)
	}
}
