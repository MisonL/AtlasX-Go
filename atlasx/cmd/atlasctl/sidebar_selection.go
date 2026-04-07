package main

import (
	"errors"
	"flag"
	"strings"

	"atlasx/internal/sidebar"
)

func runSidebarSelectionAsk(args []string) error {
	fs := flag.NewFlagSet("sidebar selection-ask", flag.ContinueOnError)
	fs.SetOutput(discardCommandOutput{})

	providerID := fs.String("provider-id", "", "override sidebar provider id")
	selectionText := fs.String("selection-text", "", "override browser selection text")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		return errors.New("missing target id or question for sidebar selection-ask")
	}

	tabID := fs.Arg(0)
	question := strings.Join(fs.Args()[1:], " ")
	paths, command, err := prepareSidebarCommand()
	if err != nil {
		return err
	}

	traceID := sidebar.NewTraceID()
	resolvedSelectionText, err := resolveSidebarSelectionText(command.Client, tabID, *selectionText)
	if err != nil {
		return finishSidebarCommand(paths, traceID, err)
	}

	selectionQuestion, err := sidebar.BuildSelectionQuestion(resolvedSelectionText, question)
	if err != nil {
		return finishSidebarCommand(paths, traceID, err)
	}

	context, memorySnippets, err := loadSidebarCommandContext(paths, command, traceID, *providerID, tabID, selectionQuestion)
	if err != nil {
		return err
	}

	response, err := command.Config.AskSelectionWithMemory(sidebar.SelectionAskRequest{
		TabID:         tabID,
		SelectionText: resolvedSelectionText,
		Question:      question,
		ProviderID:    *providerID,
	}, context, memorySnippets)
	if err != nil {
		return finishSidebarCommand(paths, traceID, err)
	}
	response.TraceID = traceID

	if err := appendSidebarCommandTurn(paths, traceID, context, selectionQuestion, response.Answer); err != nil {
		return finishSidebarCommand(paths, traceID, err)
	}
	if err := finishSidebarCommand(paths, traceID, nil); err != nil {
		return err
	}

	printSidebarAsk(response)
	return nil
}

func resolveSidebarSelectionText(client commandTabsClient, tabID string, explicitSelectionText string) (string, error) {
	if strings.TrimSpace(explicitSelectionText) != "" {
		return explicitSelectionText, nil
	}

	selection, err := client.CaptureSelection(tabID)
	if err != nil {
		return "", err
	}
	if !selection.SelectionPresent || strings.TrimSpace(selection.SelectionText) == "" {
		return "", errors.New("selection_text is required when page selection is empty")
	}
	return selection.SelectionText, nil
}
