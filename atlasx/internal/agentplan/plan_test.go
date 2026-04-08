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
	if !plan.Steps[0].Executable || plan.Steps[0].ExecutionPath != "sidebar_summarize" || !plan.Steps[0].RequiresProvider {
		t.Fatalf("unexpected suggestion step execution meta: %+v", plan.Steps[0])
	}
	if !plan.Steps[1].Executable || plan.Steps[1].ExecutionPath != "tabs_activate" || plan.Steps[1].RequiresProvider {
		t.Fatalf("unexpected recommendation step execution meta: %+v", plan.Steps[1])
	}
}

func TestDeriveStepExecutionMeta(t *testing.T) {
	cases := []struct {
		kind           string
		wantExecutable bool
		wantExecution  string
		wantProvider   bool
	}{
		{kind: "sidebar_summarize", wantExecutable: true, wantExecution: "sidebar_summarize", wantProvider: true},
		{kind: "sidebar_ask", wantExecutable: true, wantExecution: "sidebar_ask", wantProvider: true},
		{kind: "related_tab", wantExecutable: true, wantExecution: "tabs_activate", wantProvider: false},
		{kind: "memory_snippet", wantExecutable: true, wantExecution: "sidebar_memory_ask", wantProvider: true},
		{kind: "custom_preview", wantExecutable: false, wantExecution: "preview_only", wantProvider: false},
	}

	for _, tc := range cases {
		gotExecutable, gotExecution, gotProvider := deriveStepExecutionMeta(tc.kind)
		if gotExecutable != tc.wantExecutable || gotExecution != tc.wantExecution || gotProvider != tc.wantProvider {
			t.Fatalf("kind=%s unexpected meta executable=%t execution=%s provider=%t", tc.kind, gotExecutable, gotExecution, gotProvider)
		}
	}
}
