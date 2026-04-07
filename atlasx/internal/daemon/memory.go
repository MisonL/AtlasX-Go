package daemon

import (
	"fmt"
	"net/http"
	"strconv"

	"atlasx/internal/memory"
)

type memoryListResponse struct {
	Root          string         `json:"root"`
	EventsFile    string         `json:"events_file"`
	Present       bool           `json:"present"`
	EventCount    int            `json:"event_count"`
	LastEventAt   string         `json:"last_event_at"`
	LastEventKind string         `json:"last_event_kind"`
	Returned      int            `json:"returned"`
	Events        []memory.Event `json:"events"`
}

func serveMemoryList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	limit, err := parseOptionalLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	summary, events, err := memory.LoadRecent(paths, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, memoryListResponse{
		Root:          summary.Root,
		EventsFile:    summary.EventsFile,
		Present:       summary.Present,
		EventCount:    summary.EventCount,
		LastEventAt:   summary.LastEventAt,
		LastEventKind: summary.LastEventKind,
		Returned:      len(events),
		Events:        events,
	})
}

func parseOptionalLimit(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}

	limit, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid limit %q", raw)
	}
	if limit < 0 {
		return 0, fmt.Errorf("limit must be >= 0")
	}
	return limit, nil
}
