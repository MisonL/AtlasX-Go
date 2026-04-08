package agentplan

import (
	"testing"

	"atlasx/internal/contextrec"
	"atlasx/internal/suggestions"
	"atlasx/internal/tabs"
)

func TestBuildReturnsReadOnlyPlan(t *testing.T) {
	plan := Build(
		tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			CapturedAt: "2026-04-08T12:00:00Z",
		},
		[]suggestions.PageSuggestion{{
			ID:     "summarize_page",
			Kind:   "sidebar_summarize",
			Label:  "Summarize this page",
			Prompt: "Summarize this page",
			Reason: "Current page context is available.",
		}},
		[]contextrec.Recommendation{{
			ID:     "related-tab-tab-2",
			Kind:   "related_tab",
			Label:  "Atlas docs",
			Reason: "Same host.",
			TabID:  "tab-2",
			URL:    "https://chatgpt.com/docs",
			Source: "tabs",
		}},
		1,
	)

	if !plan.ReadOnly || plan.Executed {
		t.Fatalf("unexpected plan flags: %+v", plan)
	}
	if plan.Returned != 2 || plan.SuggestionReturned != 1 || plan.RecommendationReturned != 1 {
		t.Fatalf("unexpected plan counts: %+v", plan)
	}
	if len(plan.Guardrails) != 3 {
		t.Fatalf("unexpected guardrails: %+v", plan)
	}
	if plan.Rollback != "read_only_preview_no_actions_executed" {
		t.Fatalf("unexpected rollback: %+v", plan)
	}
	if !plan.Steps[0].RequiresConfirmation || !plan.Steps[1].RequiresConfirmation {
		t.Fatalf("unexpected step confirmation flags: %+v", plan)
	}
}
