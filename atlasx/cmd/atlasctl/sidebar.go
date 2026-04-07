package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
	"atlasx/internal/tabs"
)

func runSidebar(args []string) error {
	if len(args) == 0 {
		return errors.New("missing sidebar subcommand: status, ask, summarize, selection-ask")
	}

	switch args[0] {
	case "status":
		return runSidebarStatus(args[1:])
	case "ask":
		return runSidebarAsk(args[1:])
	case "summarize":
		return runSidebarSummarize(args[1:])
	case "selection-ask":
		return runSidebarSelectionAsk(args[1:])
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

	paths, command, err := prepareSidebarCommand()
	if err != nil {
		return err
	}

	traceID := sidebar.NewTraceID()
	context, memorySnippets, err := loadSidebarCommandContext(paths, command, traceID, "", args[0], sidebar.PageSummaryQuestion)
	if err != nil {
		return err
	}

	response, err := command.Config.SummarizeWithMemory(sidebar.SummarizeRequest{
		TabID: args[0],
	}, context, memorySnippets)
	if err != nil {
		return finishSidebarCommand(paths, traceID, err)
	}
	response.TraceID = traceID

	if err := appendSidebarCommandTurn(paths, traceID, context, sidebar.PageSummaryQuestion, response.Summary); err != nil {
		return finishSidebarCommand(paths, traceID, err)
	}

	if err := finishSidebarCommand(paths, traceID, nil); err != nil {
		return err
	}
	printSidebarSummary(response)
	return nil
}

func runSidebarAsk(args []string) error {
	fs := flag.NewFlagSet("sidebar ask", flag.ContinueOnError)
	fs.SetOutput(discardCommandOutput{})

	providerID := fs.String("provider-id", "", "override sidebar provider id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		return errors.New("missing target id or question for sidebar ask")
	}

	tabID := fs.Arg(0)
	question := strings.Join(fs.Args()[1:], " ")
	paths, command, err := prepareSidebarCommand()
	if err != nil {
		return err
	}

	traceID := sidebar.NewTraceID()
	context, memorySnippets, err := loadSidebarCommandContext(paths, command, traceID, *providerID, tabID, question)
	if err != nil {
		return err
	}

	response, err := command.Config.AskWithMemory(sidebar.AskRequest{
		TabID:      tabID,
		Question:   question,
		ProviderID: *providerID,
	}, context, memorySnippets)
	if err != nil {
		return finishSidebarCommand(paths, traceID, err)
	}
	response.TraceID = traceID

	if err := appendSidebarCommandTurn(paths, traceID, context, question, response.Answer); err != nil {
		return finishSidebarCommand(paths, traceID, err)
	}
	if err := finishSidebarCommand(paths, traceID, nil); err != nil {
		return err
	}

	printSidebarAsk(response)
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

func printSidebarAsk(response sidebar.AskResponse) {
	fmt.Printf(
		"answer=%q provider=%s model=%s trace_id=%s\n",
		response.Answer,
		response.Provider,
		response.Model,
		response.TraceID,
	)
	fmt.Printf("context_summary=%q\n", response.ContextSummary)
}

type sidebarCommand struct {
	Config sidebar.Config
	Client commandTabsClient
}

func prepareSidebarCommand() (macos.Paths, sidebarCommand, error) {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return macos.Paths{}, sidebarCommand{}, err
	}

	config, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return macos.Paths{}, sidebarCommand{}, err
	}
	command := sidebarCommand{
		Config: sidebar.FromSettings(config),
	}
	command.Client, err = newCommandTabsClient(paths)
	if err != nil {
		return macos.Paths{}, sidebarCommand{}, err
	}
	return paths, command, nil
}

func loadSidebarCommandContext(paths macos.Paths, command sidebarCommand, traceID string, providerID string, tabID string, question string) (tabs.PageContext, []string, error) {
	if err := command.Config.Validate(providerID); err != nil {
		return tabs.PageContext{}, nil, finishSidebarCommand(paths, traceID, err)
	}

	context, err := command.Client.Capture(tabID)
	if err != nil {
		return tabs.PageContext{}, nil, finishSidebarCommand(paths, traceID, err)
	}

	memorySnippets, err := memory.FindRelevantSnippets(paths, memory.RetrievalInput{
		TabID:    context.ID,
		Title:    context.Title,
		URL:      context.URL,
		Question: question,
	})
	if err != nil {
		return tabs.PageContext{}, nil, finishSidebarCommand(paths, traceID, err)
	}
	return context, memorySnippets, nil
}

func appendSidebarCommandTurn(paths macos.Paths, traceID string, context tabs.PageContext, question string, answer string) error {
	return memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: time.Now().UTC().Format(time.RFC3339Nano),
		TabID:      context.ID,
		Title:      context.Title,
		URL:        context.URL,
		Question:   question,
		Answer:     answer,
		CitedURLs:  nonEmptyCommandURLs(context.URL),
		TraceID:    traceID,
	})
}

func finishSidebarCommand(paths macos.Paths, traceID string, commandErr error) error {
	if saveErr := sidebar.SaveRuntimeResult(paths, traceID, commandErr); saveErr != nil {
		return saveErr
	}
	return commandErr
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

type discardCommandOutput struct{}

func (discardCommandOutput) Write(p []byte) (int, error) {
	return len(p), nil
}
