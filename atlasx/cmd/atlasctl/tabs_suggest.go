package main

import (
	"errors"
	"fmt"
	"strings"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/suggestions"
	"atlasx/internal/tabs"
)

func runTabsSuggest(paths macos.Paths, client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing target id for tabs suggest")
	}

	context, err := client.Capture(args[0])
	if err != nil {
		printPageContext(context)
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

	pageSuggestions := suggestions.ForPage(context, memorySnippets)
	printPageSuggestions(context, pageSuggestions, len(memorySnippets))
	return nil
}

func printPageSuggestions(context tabs.PageContext, pageSuggestions []suggestions.PageSuggestion, memoryReturned int) {
	fmt.Printf(
		"id=%s title=%q url=%s captured_at=%s returned=%d memory_returned=%d\n",
		context.ID,
		context.Title,
		context.URL,
		context.CapturedAt,
		len(pageSuggestions),
		memoryReturned,
	)
	for index, suggestion := range pageSuggestions {
		fmt.Printf(
			"index=%d suggestion_id=%s kind=%s label=%q\n",
			index,
			suggestion.ID,
			suggestion.Kind,
			suggestion.Label,
		)
		fmt.Printf("prompt=%q\n", suggestion.Prompt)
		fmt.Printf("reason=%q\n", suggestion.Reason)
	}
}

func buildPageSuggestionQuery(context tabs.PageContext) string {
	return strings.TrimSpace(strings.Join([]string{
		context.Title,
		context.URL,
		context.Text,
	}, " "))
}
