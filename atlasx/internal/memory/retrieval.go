package memory

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"atlasx/internal/platform/macos"
)

const (
	maxRetrievedSnippets = 3
	maxSnippetRunes      = 240
)

type RetrievalInput struct {
	TabID    string
	Title    string
	URL      string
	Question string
	Limit    int
}

type retrievalCandidate struct {
	event Event
	score int
	index int
}

func FindRelevantSnippets(paths macos.Paths, input RetrievalInput) ([]string, error) {
	events, err := Load(paths)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return findRelevantSnippetsFromEvents(events, input, nil), nil
}

func findRelevantSnippetsFromEvents(events []Event, input RetrievalInput, include func(Event) bool) []string {
	candidates := make([]retrievalCandidate, 0, len(events))
	for index, event := range events {
		if include != nil && !include(event) {
			continue
		}
		score := scoreEvent(event, input)
		if score == 0 {
			continue
		}
		candidates = append(candidates, retrievalCandidate{
			event: event,
			score: score,
			index: index,
		})
	}
	if len(candidates) == 0 {
		return nil
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].index > candidates[j].index
		}
		return candidates[i].score > candidates[j].score
	})

	limit := input.Limit
	if limit <= 0 {
		limit = maxRetrievedSnippets
	}

	snippets := make([]string, 0, limit)
	for _, candidate := range candidates {
		snippet := renderSnippet(candidate.event)
		if snippet == "" {
			continue
		}
		snippets = append(snippets, snippet)
		if len(snippets) == limit {
			break
		}
	}
	return snippets
}

func FindRelevantSnippetsForPage(paths macos.Paths, input RetrievalInput) ([]string, error) {
	controls, err := LoadControls(paths)
	if err != nil {
		return nil, err
	}
	if !controls.PageVisibilityEnabled {
		return nil, nil
	}
	if isHiddenHostForPage(input.URL, controls.HiddenHosts) {
		return nil, nil
	}

	events, err := Load(paths)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return findRelevantSnippetsFromEvents(events, input, func(event Event) bool {
		return !isHiddenHostForPage(event.URL, controls.HiddenHosts)
	}), nil
}

func isHiddenHostForPage(rawURL string, hiddenHosts []string) bool {
	if len(hiddenHosts) == 0 {
		return false
	}
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	host := strings.TrimSpace(strings.ToLower(strings.TrimSuffix(parsed.Hostname(), ".")))
	if host == "" {
		return false
	}
	for _, hiddenHost := range hiddenHosts {
		if hiddenHost == host {
			return true
		}
	}
	return false
}

func scoreEvent(event Event, input RetrievalInput) int {
	score := 0
	if input.URL != "" {
		if event.URL == input.URL {
			score += 8
		}
		for _, citedURL := range event.CitedURLs {
			if citedURL == input.URL {
				score += 8
				break
			}
		}
	}
	if input.TabID != "" && event.TabID == input.TabID {
		score += 4
	}
	if input.Title != "" && strings.EqualFold(strings.TrimSpace(event.Title), strings.TrimSpace(input.Title)) {
		score += 2
	}

	questionTokens := tokenize(input.Question)
	if len(questionTokens) > 0 {
		eventTokens := tokenize(strings.Join([]string{
			event.Title,
			event.URL,
			event.Question,
			event.Answer,
			strings.Join(event.CitedURLs, " "),
		}, " "))
		score += sharedTokenCount(questionTokens, eventTokens)
	}
	return score
}

func renderSnippet(event Event) string {
	switch event.Kind {
	case EventKindPageCapture:
		return truncateRunes(fmt.Sprintf(
			"page_capture occurred_at=%s title=%q url=%s",
			event.OccurredAt,
			event.Title,
			event.URL,
		), maxSnippetRunes)
	case EventKindQATurn:
		return truncateRunes(fmt.Sprintf(
			"qa_turn occurred_at=%s title=%q url=%s question=%q answer=%q cited_urls=%s",
			event.OccurredAt,
			event.Title,
			event.URL,
			event.Question,
			event.Answer,
			strings.Join(event.CitedURLs, ","),
		), maxSnippetRunes)
	default:
		return ""
	}
}

func tokenize(value string) map[string]struct{} {
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})

	tokens := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		if len(field) < 3 {
			continue
		}
		tokens[field] = struct{}{}
	}
	return tokens
}

func sharedTokenCount(left map[string]struct{}, right map[string]struct{}) int {
	shared := 0
	for token := range left {
		if _, ok := right[token]; ok {
			shared++
		}
	}
	return shared
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	if limit <= 3 {
		return string(runes[:limit])
	}
	return string(runes[:limit-3]) + "..."
}
