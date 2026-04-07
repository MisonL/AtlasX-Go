package memory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"atlasx/internal/platform/macos"
)

const (
	EventKindPageCapture = "page_capture"
	EventKindQATurn      = "qa_turn"
)

type Event struct {
	Kind       string   `json:"kind"`
	OccurredAt string   `json:"occurred_at"`
	TabID      string   `json:"tab_id"`
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	Question   string   `json:"question,omitempty"`
	Answer     string   `json:"answer,omitempty"`
	CitedURLs  []string `json:"cited_urls,omitempty"`
	TraceID    string   `json:"trace_id,omitempty"`
}

type Summary struct {
	Root          string `json:"root"`
	EventsFile    string `json:"events_file"`
	Present       bool   `json:"present"`
	EventCount    int    `json:"event_count"`
	LastEventAt   string `json:"last_event_at"`
	LastEventKind string `json:"last_event_kind"`
}

type PageCaptureInput struct {
	OccurredAt string
	TabID      string
	Title      string
	URL        string
}

type QATurnInput struct {
	OccurredAt string
	TabID      string
	Title      string
	URL        string
	Question   string
	Answer     string
	CitedURLs  []string
	TraceID    string
}

func AppendPageCapture(paths macos.Paths, input PageCaptureInput) error {
	return appendEvent(paths, Event{
		Kind:       EventKindPageCapture,
		OccurredAt: normalizeOccurredAt(input.OccurredAt),
		TabID:      input.TabID,
		Title:      input.Title,
		URL:        input.URL,
	})
}

func AppendQATurn(paths macos.Paths, input QATurnInput) error {
	return appendEvent(paths, Event{
		Kind:       EventKindQATurn,
		OccurredAt: normalizeOccurredAt(input.OccurredAt),
		TabID:      input.TabID,
		Title:      input.Title,
		URL:        input.URL,
		Question:   input.Question,
		Answer:     input.Answer,
		CitedURLs:  append([]string(nil), input.CitedURLs...),
		TraceID:    input.TraceID,
	})
}

func Load(paths macos.Paths) ([]Event, error) {
	file, err := os.Open(paths.MemoryEventsFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	events := make([]Event, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var event Event
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, fmt.Errorf("decode memory event: %w", err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func LoadSummary(paths macos.Paths) (Summary, error) {
	summary := Summary{
		Root:       paths.MemoryRoot,
		EventsFile: paths.MemoryEventsFile,
	}

	events, err := Load(paths)
	if err != nil {
		if os.IsNotExist(err) {
			return summary, nil
		}
		return Summary{}, err
	}

	summary.Present = true
	summary.EventCount = len(events)
	if len(events) > 0 {
		last := events[len(events)-1]
		summary.LastEventAt = last.OccurredAt
		summary.LastEventKind = last.Kind
	}
	return summary, nil
}

func LoadRecent(paths macos.Paths, limit int) (Summary, []Event, error) {
	summary := Summary{
		Root:       paths.MemoryRoot,
		EventsFile: paths.MemoryEventsFile,
	}

	events, err := Load(paths)
	if err != nil {
		if os.IsNotExist(err) {
			return summary, []Event{}, nil
		}
		return Summary{}, nil, err
	}

	summary.Present = true
	summary.EventCount = len(events)
	if len(events) > 0 {
		last := events[len(events)-1]
		summary.LastEventAt = last.OccurredAt
		summary.LastEventKind = last.Kind
	}

	if limit > 0 && len(events) > limit {
		events = append([]Event(nil), events[len(events)-limit:]...)
		return summary, events, nil
	}
	return summary, append([]Event(nil), events...), nil
}

func appendEvent(paths macos.Paths, event Event) error {
	if err := macos.EnsureDir(paths.MemoryRoot); err != nil {
		return err
	}

	file, err := os.OpenFile(paths.MemoryEventsFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

func normalizeOccurredAt(value string) string {
	if value != "" {
		return value
	}
	return time.Now().UTC().Format(time.RFC3339Nano)
}
