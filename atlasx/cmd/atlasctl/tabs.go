package main

import (
	"errors"
	"fmt"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/tabs"
)

var newCommandTabsClient = func(paths macos.Paths) (commandTabsClient, error) {
	return tabs.New(paths)
}

type commandTabsClient interface {
	List() ([]tabs.Target, error)
	Open(string) (tabs.Target, error)
	Activate(string) error
	Close(string) error
	Navigate(string, string) error
	Capture(string) (tabs.PageContext, error)
	CaptureSelection(string) (tabs.SelectionContext, error)
	DevTools(string) (tabs.DevToolsTarget, error)
}

func runTabs(args []string) error {
	if len(args) == 0 {
		return errors.New("missing tabs subcommand: list, open, activate, close, navigate, capture, selection, devtools, suggest, organize")
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	client, err := newCommandTabsClient(paths)
	if err != nil {
		return err
	}

	switch args[0] {
	case "list":
		targets, err := client.List()
		if err != nil {
			return err
		}
		for _, target := range tabs.PageTargets(targets) {
			fmt.Printf("id=%s type=%s title=%q url=%s\n", target.ID, target.Type, target.Title, target.URL)
		}
		return nil
	case "open":
		if len(args) < 2 {
			return errors.New("missing url for tabs open")
		}
		target, err := client.Open(args[1])
		if err != nil {
			return err
		}
		fmt.Printf("id=%s type=%s title=%q url=%s\n", target.ID, target.Type, target.Title, target.URL)
		return nil
	case "activate":
		if len(args) < 2 {
			return errors.New("missing target id for tabs activate")
		}
		if err := client.Activate(args[1]); err != nil {
			return err
		}
		fmt.Printf("activated=%s\n", args[1])
		return nil
	case "close":
		if len(args) < 2 {
			return errors.New("missing target id for tabs close")
		}
		if err := client.Close(args[1]); err != nil {
			return err
		}
		fmt.Printf("closed=%s\n", args[1])
		return nil
	case "navigate":
		if len(args) < 3 {
			return errors.New("missing target id or url for tabs navigate")
		}
		if err := client.Navigate(args[1], args[2]); err != nil {
			return err
		}
		fmt.Printf("navigated=%s url=%s\n", args[1], args[2])
		return nil
	case "capture":
		return runTabsCapture(paths, client, args[1:])
	case "selection":
		return runTabsSelection(client, args[1:])
	case "devtools":
		return runTabsDevTools(client, args[1:])
	case "suggest":
		return runTabsSuggest(paths, client, args[1:])
	case "organize":
		return runTabsOrganize(paths, client)
	default:
		return fmt.Errorf("unknown tabs subcommand %q", args[0])
	}
}

func runTabsCapture(paths macos.Paths, client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing target id for tabs capture")
	}

	context, err := client.Capture(args[0])
	if err != nil {
		printPageContext(context)
		return err
	}
	if err := memory.AppendPageCapture(paths, memory.PageCaptureInput{
		OccurredAt: context.CapturedAt,
		TabID:      context.ID,
		Title:      context.Title,
		URL:        context.URL,
	}); err != nil {
		return err
	}
	printPageContext(context)
	return nil
}

func runTabsSelection(client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing target id for tabs selection")
	}

	selection, err := client.CaptureSelection(args[0])
	if err != nil {
		printSelectionContext(selection)
		return err
	}
	printSelectionContext(selection)
	return nil
}

func runTabsDevTools(client commandTabsClient, args []string) error {
	if len(args) < 1 {
		return errors.New("missing target id for tabs devtools")
	}

	target, err := client.DevTools(args[0])
	if err != nil {
		return err
	}
	printDevToolsTarget(target)
	return nil
}

func printPageContext(context tabs.PageContext) {
	fmt.Printf(
		"id=%s title=%q url=%s captured_at=%s text_length=%d text_limit=%d text_truncated=%t capture_error=%q\n",
		context.ID,
		context.Title,
		context.URL,
		context.CapturedAt,
		context.TextLength,
		context.TextLimit,
		context.TextTruncated,
		context.CaptureError,
	)
	fmt.Printf("text=%q\n", context.Text)
}

func printSelectionContext(context tabs.SelectionContext) {
	fmt.Printf(
		"id=%s title=%q url=%s captured_at=%s selection_present=%t selection_text_length=%d selection_text_limit=%d selection_text_truncated=%t capture_error=%q\n",
		context.ID,
		context.Title,
		context.URL,
		context.CapturedAt,
		context.SelectionPresent,
		context.SelectionTextLength,
		context.SelectionTextLimit,
		context.SelectionTextTruncated,
		context.CaptureError,
	)
	fmt.Printf("selection_text=%q\n", context.SelectionText)
}

func printDevToolsTarget(target tabs.DevToolsTarget) {
	fmt.Printf(
		"id=%s title=%q url=%s devtools_frontend_url=%s\n",
		target.ID,
		target.Title,
		target.URL,
		target.DevToolsFrontendURL,
	)
}
