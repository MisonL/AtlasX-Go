package agentplan

import (
	"fmt"
	"strings"

	"atlasx/internal/contextrec"
	"atlasx/internal/suggestions"
	"atlasx/internal/tabs"
)

type Step struct {
	ID                   string `json:"id"`
	Kind                 string `json:"kind"`
	Title                string `json:"title"`
	Source               string `json:"source"`
	Reason               string `json:"reason"`
	Prompt               string `json:"prompt,omitempty"`
	TabID                string `json:"tab_id,omitempty"`
	URL                  string `json:"url,omitempty"`
	Snippet              string `json:"snippet,omitempty"`
	RequiresConfirmation bool   `json:"requires_confirmation"`
}

type Plan struct {
	ID                     string   `json:"id"`
	Title                  string   `json:"title"`
	URL                    string   `json:"url"`
	CapturedAt             string   `json:"captured_at"`
	Goal                   string   `json:"goal"`
	ReadOnly               bool     `json:"read_only"`
	Executed               bool     `json:"executed"`
	Returned               int      `json:"returned"`
	MemoryReturned         int      `json:"memory_returned"`
	SuggestionReturned     int      `json:"suggestion_returned"`
	RecommendationReturned int      `json:"recommendation_returned"`
	Rollback               string   `json:"rollback"`
	Guardrails             []string `json:"guardrails"`
	Steps                  []Step   `json:"steps"`
}

func Build(context tabs.PageContext, pageSuggestions []suggestions.PageSuggestion, recommendations []contextrec.Recommendation, memoryReturned int) Plan {
	steps := make([]Step, 0, len(pageSuggestions)+len(recommendations))
	for _, suggestion := range pageSuggestions {
		steps = append(steps, Step{
			ID:                   "suggest-" + suggestion.ID,
			Kind:                 suggestion.Kind,
			Title:                suggestion.Label,
			Source:               "suggestions",
			Reason:               suggestion.Reason,
			Prompt:               suggestion.Prompt,
			RequiresConfirmation: true,
		})
	}
	for _, recommendation := range recommendations {
		steps = append(steps, Step{
			ID:                   "recommend-" + recommendation.ID,
			Kind:                 recommendation.Kind,
			Title:                recommendation.Label,
			Source:               recommendation.Source,
			Reason:               recommendation.Reason,
			TabID:                recommendation.TabID,
			URL:                  recommendation.URL,
			Snippet:              recommendation.Snippet,
			RequiresConfirmation: true,
		})
	}

	return Plan{
		ID:                     context.ID,
		Title:                  context.Title,
		URL:                    context.URL,
		CapturedAt:             context.CapturedAt,
		Goal:                   buildGoal(context),
		ReadOnly:               true,
		Executed:               false,
		Returned:               len(steps),
		MemoryReturned:         memoryReturned,
		SuggestionReturned:     len(pageSuggestions),
		RecommendationReturned: len(recommendations),
		Rollback:               "read_only_preview_no_actions_executed",
		Guardrails: []string{
			"preview_only",
			"requires_confirmation_before_any_action",
			"no_browser_or_provider_side_effects",
		},
		Steps: steps,
	}
}

func buildGoal(context tabs.PageContext) string {
	title := strings.TrimSpace(context.Title)
	if title == "" {
		return "inspect the current page and propose the next safe actions"
	}
	return fmt.Sprintf("inspect %q and propose the next safe actions", title)
}
