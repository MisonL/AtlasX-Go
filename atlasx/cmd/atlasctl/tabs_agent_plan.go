package main

import (
	"errors"
	"fmt"

	"atlasx/internal/agentplan"
	"atlasx/internal/contextrec"
	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/suggestions"
)

func runTabsAgentPlan(paths macos.Paths, client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing target id for tabs agent-plan")
	}

	context, err := client.Capture(args[0])
	if err != nil {
		printPageContext(context)
		return err
	}

	targets, err := client.List()
	if err != nil {
		return err
	}

	memorySnippets, err := memory.FindRelevantSnippets(paths, memory.RetrievalInput{
		TabID:    context.ID,
		Title:    context.Title,
		URL:      context.URL,
		Question: buildPageSuggestionQuery(context),
	})
	if err != nil {
		return err
	}

	plan := agentplan.Build(
		context,
		suggestions.ForPage(context, memorySnippets),
		contextrec.ForPage(context, targets, memorySnippets),
		len(memorySnippets),
	)
	printAgentPlan(plan)
	return nil
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
			"index=%d step_id=%s kind=%s title=%q source=%s requires_confirmation=%t tab_id=%s url=%s\n",
			index,
			step.ID,
			step.Kind,
			step.Title,
			step.Source,
			step.RequiresConfirmation,
			step.TabID,
			step.URL,
		)
		fmt.Printf("reason=%q\n", step.Reason)
		fmt.Printf("prompt=%q\n", step.Prompt)
		fmt.Printf("snippet=%q\n", step.Snippet)
	}
}
