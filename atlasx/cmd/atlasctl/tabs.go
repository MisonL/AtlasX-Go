package main

import (
	"errors"
	"fmt"
	"strconv"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/tabs"
)

var newCommandTabsClient = func(paths macos.Paths) (commandTabsClient, error) {
	return tabs.New(paths)
}

type commandTabsClient interface {
	List() ([]tabs.Target, error)
	Search(string) ([]tabs.Target, error)
	Windows() ([]tabs.WindowSummary, error)
	CloseDuplicates() (tabs.CloseDuplicatesResult, error)
	OpenInWindow(int, string) (tabs.WindowOpenResult, error)
	MoveToWindow(string, int) (tabs.WindowMoveResult, error)
	MoveToNewWindow(string) (tabs.WindowMoveToNewResult, error)
	MergeWindow(int, int) (tabs.WindowMergeResult, error)
	ActivateWindow(int) (tabs.WindowActivateResult, error)
	CloseWindow(int) (tabs.WindowCloseResult, error)
	SetWindowState(int, string) (tabs.WindowBounds, error)
	SetWindowBounds(int, int, int, int, int) (tabs.WindowBounds, error)
	OpenDevToolsWindow(string) (tabs.Target, error)
	Open(string) (tabs.Target, error)
	OpenWindow(string) (tabs.Target, error)
	Activate(string) error
	Close(string) error
	Navigate(string, string) error
	Capture(string) (tabs.PageContext, error)
	CaptureSemanticContext(string) (tabs.SemanticContext, error)
	CaptureSelection(string) (tabs.SelectionContext, error)
	DevTools(string) (tabs.DevToolsTarget, error)
	EmulateDevice(string, string) (tabs.DeviceEmulationResult, error)
}

func runTabs(args []string) error {
	if len(args) == 0 {
		return errors.New("missing tabs subcommand: list, search, windows, open, open-window, open-in-window, move-to-window, move-to-new-window, merge-window, open-devtools, close-duplicates, activate-window, close-window, set-window-state, set-window-bounds, activate, close, navigate, capture, extract-context, selection, devtools, emulate-device, suggest, memories, organize, recommend-context")
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
	case "search":
		return runTabsSearch(client, args[1:])
	case "windows":
		return runTabsWindows(client)
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
	case "open-window":
		if len(args) < 2 {
			return errors.New("missing url for tabs open-window")
		}
		target, err := client.OpenWindow(args[1])
		if err != nil {
			return err
		}
		fmt.Printf("id=%s type=%s title=%q url=%s\n", target.ID, target.Type, target.Title, target.URL)
		return nil
	case "open-in-window":
		if len(args) < 3 {
			return errors.New("missing window id or url for tabs open-in-window")
		}
		windowID, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid window id %q", args[1])
		}
		result, err := client.OpenInWindow(windowID, args[2])
		if err != nil {
			return err
		}
		fmt.Printf(
			"window_id=%d activated_target_id=%s id=%s type=%s title=%q url=%s\n",
			result.WindowID,
			result.ActivatedTargetID,
			result.Target.ID,
			result.Target.Type,
			result.Target.Title,
			result.Target.URL,
		)
		return nil
	case "move-to-window":
		if len(args) < 3 {
			return errors.New("missing target id or target window id for tabs move-to-window")
		}
		targetWindowID, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid target window id %q", args[2])
		}
		result, err := client.MoveToWindow(args[1], targetWindowID)
		if err != nil {
			return err
		}
		fmt.Printf(
			"source_window_id=%d target_window_id=%d source_target_id=%s activated_target_id=%s id=%s type=%s title=%q url=%s\n",
			result.SourceWindowID,
			result.TargetWindowID,
			result.SourceTargetID,
			result.ActivatedTargetID,
			result.Target.ID,
			result.Target.Type,
			result.Target.Title,
			result.Target.URL,
		)
		return nil
	case "move-to-new-window":
		if len(args) < 2 {
			return errors.New("missing target id for tabs move-to-new-window")
		}
		result, err := client.MoveToNewWindow(args[1])
		if err != nil {
			return err
		}
		fmt.Printf(
			"source_window_id=%d source_target_id=%s id=%s type=%s title=%q url=%s\n",
			result.SourceWindowID,
			result.SourceTargetID,
			result.Target.ID,
			result.Target.Type,
			result.Target.Title,
			result.Target.URL,
		)
		return nil
	case "merge-window":
		if len(args) < 3 {
			return errors.New("missing source or target window id for tabs merge-window")
		}
		sourceWindowID, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid source window id %q", args[1])
		}
		targetWindowID, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid target window id %q", args[2])
		}
		result, err := client.MergeWindow(sourceWindowID, targetWindowID)
		if err != nil {
			return err
		}
		fmt.Printf("source_window_id=%d target_window_id=%d returned=%d\n", result.SourceWindowID, result.TargetWindowID, result.Returned)
		for index, moved := range result.MovedTargets {
			fmt.Printf(
				"index=%d source_target_id=%s activated_target_id=%s id=%s type=%s title=%q url=%s\n",
				index,
				moved.SourceTargetID,
				moved.ActivatedTargetID,
				moved.Target.ID,
				moved.Target.Type,
				moved.Target.Title,
				moved.Target.URL,
			)
		}
		return nil
	case "open-devtools":
		if len(args) < 2 {
			return errors.New("missing target id for tabs open-devtools")
		}
		target, err := client.OpenDevToolsWindow(args[1])
		if err != nil {
			return err
		}
		fmt.Printf("id=%s type=%s title=%q url=%s\n", target.ID, target.Type, target.Title, target.URL)
		return nil
	case "close-duplicates":
		return runTabsCloseDuplicates(client)
	case "activate-window":
		if len(args) < 2 {
			return errors.New("missing window id for tabs activate-window")
		}
		windowID, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid window id %q", args[1])
		}
		result, err := client.ActivateWindow(windowID)
		if err != nil {
			return err
		}
		fmt.Printf("window_id=%d activated_target_id=%s\n", result.WindowID, result.ActivatedTargetID)
		return nil
	case "close-window":
		if len(args) < 2 {
			return errors.New("missing window id for tabs close-window")
		}
		windowID, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid window id %q", args[1])
		}
		result, err := client.CloseWindow(windowID)
		if err != nil {
			return err
		}
		fmt.Printf("window_id=%d returned=%d\n", result.WindowID, result.Returned)
		for index, targetID := range result.ClosedTargets {
			fmt.Printf("closed_target_index=%d id=%s\n", index, targetID)
		}
		return nil
	case "set-window-state":
		if len(args) < 3 {
			return errors.New("missing window id or state for tabs set-window-state")
		}
		windowID, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid window id %q", args[1])
		}
		result, err := client.SetWindowState(windowID, args[2])
		if err != nil {
			return err
		}
		fmt.Printf(
			"window_id=%d state=%s left=%d top=%d width=%d height=%d\n",
			result.WindowID,
			result.State,
			result.Left,
			result.Top,
			result.Width,
			result.Height,
		)
		return nil
	case "set-window-bounds":
		if len(args) < 6 {
			return errors.New("missing window id or bounds for tabs set-window-bounds")
		}
		windowID, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid window id %q", args[1])
		}
		left, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("invalid left %q", args[2])
		}
		top, err := strconv.Atoi(args[3])
		if err != nil {
			return fmt.Errorf("invalid top %q", args[3])
		}
		width, err := strconv.Atoi(args[4])
		if err != nil {
			return fmt.Errorf("invalid width %q", args[4])
		}
		height, err := strconv.Atoi(args[5])
		if err != nil {
			return fmt.Errorf("invalid height %q", args[5])
		}
		result, err := client.SetWindowBounds(windowID, left, top, width, height)
		if err != nil {
			return err
		}
		fmt.Printf(
			"window_id=%d state=%s left=%d top=%d width=%d height=%d\n",
			result.WindowID,
			result.State,
			result.Left,
			result.Top,
			result.Width,
			result.Height,
		)
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
	case "extract-context":
		return runTabsExtractContext(client, args[1:])
	case "selection":
		return runTabsSelection(client, args[1:])
	case "devtools":
		return runTabsDevTools(client, args[1:])
	case "emulate-device":
		return runTabsEmulateDevice(client, args[1:])
	case "suggest":
		return runTabsSuggest(paths, client, args[1:])
	case "memories":
		return runTabsMemories(paths, client, args[1:])
	case "organize":
		return runTabsOrganize(paths, client)
	case "recommend-context":
		return runTabsContextRecommend(paths, client, args[1:])
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
