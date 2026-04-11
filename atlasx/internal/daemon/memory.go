package daemon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"atlasx/internal/memory"
)

const maxMemoryQueryLimit = 1000
const memoryOperationTimeout = 2 * time.Second

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
		writeMethodNotAllowed(w, r.Method, http.MethodGet)
		return
	}

	limit, err := parseOptionalLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if limit > maxMemoryQueryLimit {
		writeError(w, http.StatusBadRequest, fmt.Errorf("limit exceeds maximum of %d", maxMemoryQueryLimit))
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	recent, err := runWithTimeout(r.Context(), memoryOperationTimeout, func() (memoryRecentResult, error) {
		summary, events, err := memory.LoadRecent(paths, limit)
		return memoryRecentResult{summary: summary, events: events}, err
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusGatewayTimeout, fmt.Errorf("memory list timed out"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, memoryListResponse{
		Root:          recent.summary.Root,
		EventsFile:    recent.summary.EventsFile,
		Present:       recent.summary.Present,
		EventCount:    recent.summary.EventCount,
		LastEventAt:   recent.summary.LastEventAt,
		LastEventKind: recent.summary.LastEventKind,
		Returned:      len(recent.events),
		Events:        recent.events,
	})
}

func serveMemorySearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method, http.MethodGet)
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
	if limit > maxMemoryQueryLimit {
		writeError(w, http.StatusBadRequest, fmt.Errorf("limit exceeds maximum of %d", maxMemoryQueryLimit))
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
	snippets, err := runWithTimeout(r.Context(), memoryOperationTimeout, func() ([]string, error) {
		return memory.FindRelevantSnippets(paths, input)
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusGatewayTimeout, fmt.Errorf("memory search timed out"))
			return
		}
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

func serveMemoryControls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		paths, err := discoverPaths()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		controls, err := runWithTimeout(r.Context(), memoryOperationTimeout, func() (memory.Controls, error) {
			return memory.LoadControls(paths)
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				writeError(w, http.StatusGatewayTimeout, fmt.Errorf("memory controls load timed out"))
				return
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, controls)
	case http.MethodPost:
		var request memory.ControlsUpdate
		if err := decodeRequiredJSON(r, &request); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if request.PersistEnabled == nil && request.PageVisibilityEnabled == nil && request.SiteVisibilityEnabled == nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("persist_enabled or page_visibility_enabled or site_host + site_visibility_enabled is required"))
			return
		}
		if request.SiteVisibilityEnabled != nil && request.SiteHost == "" {
			writeError(w, http.StatusBadRequest, fmt.Errorf("site_host is required when site_visibility_enabled is provided"))
			return
		}
		paths, err := discoverPaths()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		controls, err := runWithTimeout(r.Context(), memoryOperationTimeout, func() (memory.Controls, error) {
			return memory.UpdateControls(paths, request)
		})
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				writeError(w, http.StatusGatewayTimeout, fmt.Errorf("memory controls update timed out"))
				return
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, controls)
	default:
		writeMethodNotAllowed(w, r.Method, http.MethodGet, http.MethodPost)
	}
}

type memoryRecentResult struct {
	summary memory.Summary
	events  []memory.Event
}

type memoryTimeoutResult[T any] struct {
	value T
	err   error
}

func runWithTimeout[T any](ctx context.Context, timeout time.Duration, fn func() (T, error)) (T, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resultCh := make(chan memoryTimeoutResult[T], 1)
	go func() {
		value, err := fn()
		resultCh <- memoryTimeoutResult[T]{value: value, err: err}
	}()

	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case result := <-resultCh:
		return result.value, result.err
	}
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
