package contextrec

import (
	"testing"

	"atlasx/internal/tabs"
)

func TestForPageReturnsRelatedTabsAndMemory(t *testing.T) {
	recommendations := ForPage(
		tabs.PageContext{ID: "tab-1", Title: "Atlas", URL: "https://chatgpt.com/atlas", Text: "Atlas page"},
		[]tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
			{ID: "tab-2", Type: "page", Title: "Atlas docs", URL: "https://chatgpt.com/docs"},
			{ID: "tab-3", Type: "page", Title: "Elsewhere", URL: "https://example.com/other"},
		},
		[]string{"qa_turn occurred_at=2026-04-07T12:00:00Z title=\"Atlas\""},
	)

	if len(recommendations) != 2 {
		t.Fatalf("unexpected recommendations: %+v", recommendations)
	}
	if recommendations[0].Kind != "related_tab" || recommendations[1].Kind != "memory_snippet" {
		t.Fatalf("unexpected recommendations: %+v", recommendations)
	}
}

func TestForPageReturnsEmptyWhenNothingMatches(t *testing.T) {
	recommendations := ForPage(
		tabs.PageContext{ID: "tab-1", Title: "Atlas", URL: "about:blank"},
		[]tabs.Target{{ID: "tab-1", Type: "page", Title: "Atlas", URL: "about:blank"}},
		nil,
	)

	if len(recommendations) != 0 {
		t.Fatalf("unexpected recommendations: %+v", recommendations)
	}
}
