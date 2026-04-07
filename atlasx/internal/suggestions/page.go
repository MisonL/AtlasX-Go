package suggestions

import (
	"strings"

	"atlasx/internal/sidebar"
	"atlasx/internal/tabs"
)

type PageSuggestion struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Label  string `json:"label"`
	Prompt string `json:"prompt"`
	Reason string `json:"reason"`
}

func ForPage(context tabs.PageContext, memorySnippets []string) []PageSuggestion {
	suggestions := []PageSuggestion{
		{
			ID:     "summarize_page",
			Kind:   "sidebar_summarize",
			Label:  "Summarize this page",
			Prompt: sidebar.PageSummaryQuestion,
			Reason: "Current page context is available for direct summarization.",
		},
		{
			ID:     "extract_key_points",
			Kind:   "sidebar_ask",
			Label:  "Extract key points",
			Prompt: buildKeyPointsPrompt(context),
			Reason: "The current page title and body text can be reduced to the main takeaways.",
		},
	}

	if len(memorySnippets) > 0 {
		suggestions = append(suggestions, PageSuggestion{
			ID:     "relate_memory",
			Kind:   "sidebar_ask",
			Label:  "Relate to memory",
			Prompt: "How does this page relate to the previously captured context and memory?",
			Reason: "Relevant memory snippets were found for this page.",
		})
		return suggestions
	}

	if looksLikeIssuePage(context) {
		suggestions = append(suggestions, PageSuggestion{
			ID:     "debug_page_issue",
			Kind:   "sidebar_ask",
			Label:  "Debug this issue",
			Prompt: "What issue is shown on this page, and what should I check next?",
			Reason: "The current page text looks like an error, warning, or debugging surface.",
		})
		return suggestions
	}

	suggestions = append(suggestions, PageSuggestion{
		ID:     "next_action",
		Kind:   "sidebar_ask",
		Label:  "Suggest next action",
		Prompt: "What should I do next based on the current page?",
		Reason: "No matching memory was found, so the next best action should come from current page context.",
	})
	return suggestions
}

func buildKeyPointsPrompt(context tabs.PageContext) string {
	title := strings.TrimSpace(context.Title)
	if title == "" {
		return "What are the key points of the current page?"
	}
	return `What are the key points of "` + title + `"?`
}

func looksLikeIssuePage(context tabs.PageContext) bool {
	content := strings.ToLower(strings.Join([]string{
		context.Title,
		context.URL,
		context.Text,
	}, " "))
	indicators := []string{
		"error",
		"failed",
		"exception",
		"warning",
		"traceback",
		"debug",
		"timeout",
		"not found",
		"cannot",
	}
	for _, indicator := range indicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}
	return false
}
