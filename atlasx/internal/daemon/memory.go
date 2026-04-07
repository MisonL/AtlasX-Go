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

type memorySearchResponse struct {
	Question string   `json:"question"`
	TabID    string   `json:"tab_id,omitempty"`
	Title    string   `json:"title,omitempty"`
	URL      string   `json:"url,omitempty"`
	Returned int      `json:"returned"`
	Snippets []string `json:"snippets"`
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

func serveMemorySearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	question := r.URL.Query().Get("question")
	if question == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("question is required"))
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

	input := memory.RetrievalInput{
		TabID:    r.URL.Query().Get("tab_id"),
		Title:    r.URL.Query().Get("title"),
		URL:      r.URL.Query().Get("url"),
		Question: question,
		Limit:    limit,
	}
	snippets, err := memory.FindRelevantSnippets(paths, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, memorySearchResponse{
		Question: input.Question,
		TabID:    input.TabID,
		Title:    input.Title,
		URL:      input.URL,
		Returned: len(snippets),
		Snippets: snippets,
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
