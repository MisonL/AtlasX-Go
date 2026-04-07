package main

import (
	"errors"
	"fmt"
	"time"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
)

func runSidebar(args []string) error {
	if len(args) == 0 {
		return errors.New("missing sidebar subcommand: status, summarize")
	}

	switch args[0] {
	case "status":
		return runSidebarStatus(args[1:])
	case "summarize":
		return runSidebarSummarize(args[1:])
	default:
		return fmt.Errorf("unknown sidebar subcommand %q", args[0])
	}
}

func runSidebarStatus(args []string) error {
	if len(args) != 0 {
		return errors.New("sidebar status does not accept extra arguments")
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	config, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return err
	}

	status, err := sidebar.FromSettings(config).StatusWithRuntime(paths)
	if err != nil {
		return err
	}

	fmt.Print(status.Render())
	return nil
}

func runSidebarSummarize(args []string) error {
	if len(args) < 1 {
		return errors.New("missing target id for sidebar summarize")
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	config, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return err
	}

	traceID := sidebar.NewTraceID()
	sidebarConfig := sidebar.FromSettings(config)
	if err := sidebarConfig.Validate(""); err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		return err
	}

	client, err := newCommandTabsClient(paths)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		return err
	}

	context, err := client.Capture(args[0])
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		return err
	}

	memorySnippets, err := memory.FindRelevantSnippets(paths, memory.RetrievalInput{
		TabID:    context.ID,
		Title:    context.Title,
		URL:      context.URL,
		Question: sidebar.PageSummaryQuestion,
	})
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		return err
	}

	response, err := sidebarConfig.SummarizeWithMemory(sidebar.SummarizeRequest{
		TabID: args[0],
	}, context, memorySnippets)
	if err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		return err
	}
	response.TraceID = traceID

	if err := memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: time.Now().UTC().Format(time.RFC3339Nano),
		TabID:      context.ID,
		Title:      context.Title,
		URL:        context.URL,
		Question:   sidebar.PageSummaryQuestion,
		Answer:     response.Summary,
		CitedURLs:  nonEmptyCommandURLs(context.URL),
		TraceID:    traceID,
	}); err != nil {
		_ = sidebar.SaveRuntimeResult(paths, traceID, err)
		return err
	}

	_ = sidebar.SaveRuntimeResult(paths, traceID, nil)
	printSidebarSummary(response)
	return nil
}

func printSidebarSummary(response sidebar.SummarizeResponse) {
	fmt.Printf(
		"summary=%q provider=%s model=%s trace_id=%s\n",
		response.Summary,
		response.Provider,
		response.Model,
		response.TraceID,
	)
	fmt.Printf("context_summary=%q\n", response.ContextSummary)
}

func nonEmptyCommandURLs(urls ...string) []string {
	nonEmpty := make([]string, 0, len(urls))
	for _, value := range urls {
		if value == "" {
			continue
		}
		nonEmpty = append(nonEmpty, value)
	}
	return nonEmpty
}
