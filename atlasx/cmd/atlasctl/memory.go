package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
)

func runMemory(args []string) error {
	if len(args) == 0 {
		return errors.New("memory supports subcommands: list, search, controls, set-persist")
	}

	switch args[0] {
	case "list":
		return runMemoryList(args[1:])
	case "search":
		return runMemorySearch(args[1:])
	case "controls":
		return runMemoryControls(args[1:])
	case "set-persist":
		return runMemorySetPersist(args[1:])
	default:
		return fmt.Errorf("unknown memory subcommand %q", args[0])
	}
}

func runMemoryList(args []string) error {
	fs := flag.NewFlagSet("memory list", flag.ContinueOnError)
	fs.SetOutput(discardCommandOutput{})

	limit := fs.Int("limit", 0, "return only the most recent N memory events")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *limit < 0 {
		return errors.New("limit must be >= 0")
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	summary, events, err := memory.LoadRecent(paths, *limit)
	if err != nil {
		return err
	}

	fmt.Printf(
		"memory_root=%s events_file=%s present=%t event_count=%d last_event_at=%s last_event_kind=%s returned=%d\n",
		summary.Root,
		summary.EventsFile,
		summary.Present,
		summary.EventCount,
		summary.LastEventAt,
		summary.LastEventKind,
		len(events),
	)
	for index, event := range events {
		fmt.Printf(
			"index=%d kind=%s occurred_at=%s tab_id=%s title=%q url=%s trace_id=%s\n",
			index,
			event.Kind,
			event.OccurredAt,
			event.TabID,
			event.Title,
			event.URL,
			event.TraceID,
		)
		fmt.Printf("question=%q answer=%q cited_urls=%v\n", event.Question, event.Answer, event.CitedURLs)
	}
	return nil
}

func runMemorySearch(args []string) error {
	fs := flag.NewFlagSet("memory search", flag.ContinueOnError)
	fs.SetOutput(discardCommandOutput{})

	tabID := fs.String("tab-id", "", "prefer memory events from the specified tab")
	title := fs.String("title", "", "prefer memory events with the specified title")
	url := fs.String("url", "", "prefer memory events matching the specified url")
	limit := fs.Int("limit", 0, "return only the most relevant N snippets")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *limit < 0 {
		return errors.New("limit must be >= 0")
	}
	if fs.NArg() == 0 {
		return errors.New("missing question for memory search")
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	question := strings.Join(fs.Args(), " ")

	snippets, err := memory.FindRelevantSnippets(paths, memory.RetrievalInput{
		TabID:    *tabID,
		Title:    *title,
		URL:      *url,
		Question: question,
		Limit:    *limit,
	})
	if err != nil {
		return err
	}

	fmt.Printf("question=%q tab_id=%s title=%q url=%s returned=%d\n", question, *tabID, *title, *url, len(snippets))
	for index, snippet := range snippets {
		fmt.Printf("index=%d snippet=%q\n", index, snippet)
	}
	return nil
}

func runMemoryControls(args []string) error {
	if len(args) != 0 {
		return errors.New("memory controls does not accept extra arguments")
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	controls, err := memory.LoadControls(paths)
	if err != nil {
		return err
	}

	fmt.Printf("config_file=%s persist_enabled=%t\n", controls.ConfigFile, controls.PersistEnabled)
	return nil
}

func runMemorySetPersist(args []string) error {
	if len(args) != 1 {
		return errors.New("memory set-persist requires enabled or disabled")
	}

	enabled, err := memory.ParsePersistValue(strings.ToLower(strings.TrimSpace(args[0])))
	if err != nil {
		return err
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}

	controls, err := memory.SetPersistEnabled(paths, enabled)
	if err != nil {
		return err
	}

	fmt.Printf("config_file=%s persist_enabled=%t\n", controls.ConfigFile, controls.PersistEnabled)
	return nil
}
