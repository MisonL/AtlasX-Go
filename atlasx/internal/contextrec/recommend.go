package contextrec

import (
	"net/url"
	"sort"
	"strings"

	"atlasx/internal/tabs"
)

const maxRecommendedTabs = 3

type Recommendation struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Label    string `json:"label"`
	Reason   string `json:"reason"`
	TabID    string `json:"tab_id,omitempty"`
	URL      string `json:"url,omitempty"`
	Snippet  string `json:"snippet,omitempty"`
	Source   string `json:"source,omitempty"`
	Position int    `json:"position"`
}

func ForPage(context tabs.PageContext, targets []tabs.Target, memorySnippets []string) []Recommendation {
	recommendations := make([]Recommendation, 0)
	recommendations = append(recommendations, recommendTabs(context, targets)...)
	recommendations = append(recommendations, recommendMemory(memorySnippets, len(recommendations))...)
	return recommendations
}

func recommendTabs(context tabs.PageContext, targets []tabs.Target) []Recommendation {
	currentHost := normalizedHost(context.URL)
	if currentHost == "" {
		return nil
	}

	matches := make([]tabs.Target, 0)
	for _, target := range tabs.PageTargets(targets) {
		if target.ID == context.ID {
			continue
		}
		if normalizedHost(target.URL) != currentHost {
			continue
		}
		matches = append(matches, target)
	}
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Title == matches[j].Title {
			return matches[i].ID < matches[j].ID
		}
		return matches[i].Title < matches[j].Title
	})
	if len(matches) > maxRecommendedTabs {
		matches = matches[:maxRecommendedTabs]
	}

	recommendations := make([]Recommendation, 0, len(matches))
	for index, target := range matches {
		recommendations = append(recommendations, Recommendation{
			ID:       "related-tab-" + target.ID,
			Kind:     "related_tab",
			Label:    target.Title,
			Reason:   "This tab shares the same host as the current page.",
			TabID:    target.ID,
			URL:      target.URL,
			Source:   "tabs",
			Position: index,
		})
	}
	return recommendations
}

func recommendMemory(memorySnippets []string, offset int) []Recommendation {
	recommendations := make([]Recommendation, 0, len(memorySnippets))
	for index, snippet := range memorySnippets {
		recommendations = append(recommendations, Recommendation{
			ID:       "memory-" + strings.TrimSpace(strings.ReplaceAll(strings.ToLower(snippetKey(snippet)), " ", "-")),
			Kind:     "memory_snippet",
			Label:    snippetKey(snippet),
			Reason:   "This memory snippet appears relevant to the current page.",
			Snippet:  snippet,
			Source:   "memory",
			Position: offset + index,
		})
	}
	return recommendations
}

func snippetKey(snippet string) string {
	if strings.HasPrefix(snippet, "qa_turn") {
		return "Relevant QA turn"
	}
	if strings.HasPrefix(snippet, "page_capture") {
		return "Relevant page capture"
	}
	return "Relevant memory"
}

func normalizedHost(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.ToLower(parsed.Hostname()))
}
