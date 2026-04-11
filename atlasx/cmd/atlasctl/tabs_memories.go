package main

import (
	"errors"
	"fmt"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
)

func runTabsMemories(paths macos.Paths, client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing target id for tabs memories")
	}

	context, err := client.Capture(args[0])
	if err != nil {
		printCaptureContext(context, err)
		return err
	}

	snippets, err := memory.FindRelevantSnippetsForPage(paths, memory.RetrievalInput{
		TabID:    context.ID,
		Title:    context.Title,
		URL:      context.URL,
		Question: buildPageSuggestionQuery(context),
	})
	if err != nil {
		return err
	}

	fmt.Printf(
		"id=%s title=%q url=%s captured_at=%s returned=%d\n",
		context.ID,
		context.Title,
		context.URL,
		context.CapturedAt,
		len(snippets),
	)
	for index, snippet := range snippets {
		fmt.Printf("index=%d snippet=%q\n", index, snippet)
	}
	return nil
}
