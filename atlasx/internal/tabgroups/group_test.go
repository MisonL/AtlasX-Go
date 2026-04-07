package tabgroups

import (
	"testing"

	"atlasx/internal/tabs"
)

func TestSuggestGroupsTabsByHost(t *testing.T) {
	groups := Suggest([]tabs.Target{
		{ID: "1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
		{ID: "2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
		{ID: "3", Type: "page", Title: "Elsewhere", URL: "https://example.com/other"},
	})

	if len(groups) != 1 {
		t.Fatalf("unexpected groups: %+v", groups)
	}
	if groups[0].ID != "host:chatgpt.com" || groups[0].Returned != 2 {
		t.Fatalf("unexpected group: %+v", groups[0])
	}
}

func TestSuggestGroupsTabsByTitlePrefixWhenHostMissing(t *testing.T) {
	groups := Suggest([]tabs.Target{
		{ID: "1", Type: "page", Title: "Build Log - A", URL: "about:blank"},
		{ID: "2", Type: "page", Title: "Build Log - B", URL: "about:blank"},
	})

	if len(groups) != 1 {
		t.Fatalf("unexpected groups: %+v", groups)
	}
	if groups[0].ID != "title:build log" {
		t.Fatalf("unexpected group: %+v", groups[0])
	}
}
