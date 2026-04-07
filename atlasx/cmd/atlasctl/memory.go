package main

import (
	"errors"
	"flag"
	"fmt"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
)

func runMemory(args []string) error {
	if len(args) == 0 {
		return errors.New("memory supports subcommands: list")
	}

	switch args[0] {
	case "list":
		return runMemoryList(args[1:])
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
