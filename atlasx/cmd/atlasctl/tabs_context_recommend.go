package main

import (
	"errors"
	"fmt"

	"atlasx/internal/contextrec"
	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
)

func runTabsContextRecommend(paths macos.Paths, client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing target id for tabs recommend-context")
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

	memorySnippets, err := memory.FindRelevantSnippetsForPage(paths, memory.RetrievalInput{
		TabID:    context.ID,
		Title:    context.Title,
		URL:      context.URL,
		Question: buildPageSuggestionQuery(context),
	})
	if err != nil {
		return err
	}

	recommendations := contextrec.ForPage(context, targets, memorySnippets)
	fmt.Printf(
		"id=%s title=%q url=%s captured_at=%s returned=%d memory_returned=%d\n",
		context.ID,
		context.Title,
		context.URL,
		context.CapturedAt,
		len(recommendations),
		len(memorySnippets),
	)
	for index, recommendation := range recommendations {
		fmt.Printf(
			"index=%d recommendation_id=%s kind=%s label=%q source=%s tab_id=%s url=%s\n",
			index,
			recommendation.ID,
			recommendation.Kind,
			recommendation.Label,
			recommendation.Source,
			recommendation.TabID,
			recommendation.URL,
		)
		fmt.Printf("reason=%q\n", recommendation.Reason)
		fmt.Printf("snippet=%q\n", recommendation.Snippet)
	}
	return nil
}
