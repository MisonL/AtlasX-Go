package suggestions

import (
	"testing"

	"atlasx/internal/tabs"
)

func TestForPageUsesMemoryAwareSuggestionWhenMemoryExists(t *testing.T) {
	suggestions := ForPage(tabs.PageContext{
		ID:    "tab-1",
		Title: "Atlas",
		URL:   "https://chatgpt.com/atlas",
		Text:  "Atlas page text",
	}, []string{"qa_turn occurred_at=2026-04-07T12:00:00Z title=\"Atlas\""})

	if len(suggestions) != 3 {
		t.Fatalf("unexpected suggestions: %+v", suggestions)
	}
	if suggestions[2].ID != "relate_memory" {
		t.Fatalf("unexpected suggestions: %+v", suggestions)
	}
}

func TestForPageUsesDebugSuggestionOnIssuePage(t *testing.T) {
	suggestions := ForPage(tabs.PageContext{
		ID:    "tab-1",
		Title: "Build failed",
		URL:   "https://chatgpt.com/atlas",
		Text:  "Error: timeout while running tests",
	}, nil)

	if len(suggestions) != 3 {
		t.Fatalf("unexpected suggestions: %+v", suggestions)
	}
	if suggestions[2].ID != "debug_page_issue" {
		t.Fatalf("unexpected suggestions: %+v", suggestions)
	}
}
