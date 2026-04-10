package memory

import (
	"fmt"
	"strings"
	"testing"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
)

func TestFindRelevantSnippetsReturnsRankedMatches(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if err := AppendPageCapture(paths, PageCaptureInput{
		OccurredAt: "2026-04-07T00:00:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
	}); err != nil {
		t.Fatalf("append page capture failed: %v", err)
	}
	if err := AppendQATurn(paths, QATurnInput{
		OccurredAt: "2026-04-07T00:01:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "how does memory retrieval work",
		Answer:     "Memory retrieval reuses prior Atlas context.",
		CitedURLs:  []string{"https://chatgpt.com/atlas"},
		TraceID:    "trace-1",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}
	if err := AppendQATurn(paths, QATurnInput{
		OccurredAt: "2026-04-07T00:02:00Z",
		TabID:      "tab-2",
		Title:      "Elsewhere",
		URL:        "https://example.com/other",
		Question:   "unrelated",
		Answer:     "Not relevant",
		CitedURLs:  []string{"https://example.com/other"},
		TraceID:    "trace-2",
	}); err != nil {
		t.Fatalf("append unrelated qa turn failed: %v", err)
	}

	snippets, err := FindRelevantSnippets(paths, RetrievalInput{
		TabID:    "tab-1",
		Title:    "Atlas",
		URL:      "https://chatgpt.com/atlas",
		Question: "summarize memory retrieval",
	})
	if err != nil {
		t.Fatalf("find relevant snippets failed: %v", err)
	}
	if len(snippets) != 2 {
		t.Fatalf("unexpected snippets: %+v", snippets)
	}
	if !strings.Contains(snippets[0], "qa_turn") || !strings.Contains(snippets[0], "memory retrieval") {
		t.Fatalf("unexpected first snippet: %s", snippets[0])
	}
	if !strings.Contains(snippets[1], "page_capture") {
		t.Fatalf("unexpected second snippet: %s", snippets[1])
	}
}

func TestFindRelevantSnippetsTruncatesLongAnswers(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if err := AppendQATurn(paths, QATurnInput{
		OccurredAt: "2026-04-07T00:01:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "memory",
		Answer:     strings.Repeat("A", maxSnippetRunes*2),
		CitedURLs:  []string{"https://chatgpt.com/atlas"},
		TraceID:    "trace-1",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}

	snippets, err := FindRelevantSnippets(paths, RetrievalInput{
		URL:      "https://chatgpt.com/atlas",
		Question: "memory",
	})
	if err != nil {
		t.Fatalf("find relevant snippets failed: %v", err)
	}
	if len(snippets) != 1 {
		t.Fatalf("unexpected snippets: %+v", snippets)
	}
	if len([]rune(snippets[0])) > maxSnippetRunes {
		t.Fatalf("expected snippet truncation: %s", snippets[0])
	}
	if !strings.HasSuffix(snippets[0], "...") {
		t.Fatalf("expected truncated snippet suffix: %s", snippets[0])
	}
}

func TestFindRelevantSnippetsWithoutMemoryReturnsEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	snippets, err := FindRelevantSnippets(paths, RetrievalInput{
		URL:      "https://chatgpt.com/atlas",
		Question: "memory",
	})
	if err != nil {
		t.Fatalf("find relevant snippets failed: %v", err)
	}
	if len(snippets) != 0 {
		t.Fatalf("expected empty snippets, got %+v", snippets)
	}
}

func TestFindRelevantSnippetsRespectsCustomLimit(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	for index := 0; index < 3; index++ {
		if err := AppendQATurn(paths, QATurnInput{
			OccurredAt: fmt.Sprintf("2026-04-07T00:0%d:00Z", index+1),
			TabID:      "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Question:   fmt.Sprintf("memory retrieval %d", index+1),
			Answer:     "Atlas answer",
			CitedURLs:  []string{"https://chatgpt.com/atlas"},
			TraceID:    fmt.Sprintf("trace-%d", index+1),
		}); err != nil {
			t.Fatalf("append qa turn failed: %v", err)
		}
	}

	snippets, err := FindRelevantSnippets(paths, RetrievalInput{
		URL:      "https://chatgpt.com/atlas",
		Question: "memory retrieval",
		Limit:    1,
	})
	if err != nil {
		t.Fatalf("find relevant snippets failed: %v", err)
	}
	if len(snippets) != 1 {
		t.Fatalf("unexpected snippets: %+v", snippets)
	}
}

func TestFindRelevantSnippetsForPageRespectsVisibilityControl(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := AppendQATurn(paths, QATurnInput{
		OccurredAt: "2026-04-07T00:01:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "memory",
		Answer:     "Atlas memory answer",
		TraceID:    "trace-1",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
		MemoryPageVisibility: settings.Bool(false),
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	pageSnippets, err := FindRelevantSnippetsForPage(paths, RetrievalInput{
		URL:      "https://chatgpt.com/atlas",
		Question: "memory",
	})
	if err != nil {
		t.Fatalf("find relevant snippets for page failed: %v", err)
	}
	if len(pageSnippets) != 0 {
		t.Fatalf("expected hidden page snippets, got %+v", pageSnippets)
	}

	searchSnippets, err := FindRelevantSnippets(paths, RetrievalInput{
		URL:      "https://chatgpt.com/atlas",
		Question: "memory",
	})
	if err != nil {
		t.Fatalf("find relevant snippets failed: %v", err)
	}
	if len(searchSnippets) == 0 {
		t.Fatalf("expected direct memory search to remain available, got %+v", searchSnippets)
	}
	if !strings.Contains(searchSnippets[0], "chatgpt.com") && (len(searchSnippets) < 2 || !strings.Contains(searchSnippets[1], "chatgpt.com")) {
		t.Fatalf("expected direct memory search to retain hidden-site snippets, got %+v", searchSnippets)
	}
}

func TestFindRelevantSnippetsForPageRespectsHiddenHosts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := AppendQATurn(paths, QATurnInput{
		OccurredAt: "2026-04-07T00:01:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "memory",
		Answer:     "Atlas memory answer",
		TraceID:    "trace-1",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}
	if err := AppendQATurn(paths, QATurnInput{
		OccurredAt: "2026-04-07T00:02:00Z",
		TabID:      "tab-2",
		Title:      "Example",
		URL:        "https://example.com/page",
		Question:   "memory",
		Answer:     "Example memory answer",
		TraceID:    "trace-2",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
		MemoryHiddenHosts: []string{"chatgpt.com"},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	hiddenSnippets, err := FindRelevantSnippetsForPage(paths, RetrievalInput{
		URL:      "https://chatgpt.com/atlas",
		Question: "memory",
	})
	if err != nil {
		t.Fatalf("find relevant snippets for hidden site failed: %v", err)
	}
	if len(hiddenSnippets) != 0 {
		t.Fatalf("expected hidden site snippets to be empty, got %+v", hiddenSnippets)
	}

	visibleSnippets, err := FindRelevantSnippetsForPage(paths, RetrievalInput{
		URL:      "https://example.com/page",
		Question: "memory",
	})
	if err != nil {
		t.Fatalf("find relevant snippets for visible site failed: %v", err)
	}
	if len(visibleSnippets) != 1 {
		t.Fatalf("expected visible site snippets, got %+v", visibleSnippets)
	}

	searchSnippets, err := FindRelevantSnippets(paths, RetrievalInput{
		URL:      "https://chatgpt.com/atlas",
		Question: "memory",
	})
	if err != nil {
		t.Fatalf("find relevant snippets failed: %v", err)
	}
	if len(searchSnippets) == 0 {
		t.Fatalf("expected direct memory search to remain available, got %+v", searchSnippets)
	}
	joined := strings.Join(searchSnippets, "\n")
	if !strings.Contains(joined, "chatgpt.com/atlas") {
		t.Fatalf("expected direct memory search to retain hidden-site snippets, got %+v", searchSnippets)
	}
}
