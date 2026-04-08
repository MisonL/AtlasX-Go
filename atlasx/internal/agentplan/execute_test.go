package agentplan

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
	"atlasx/internal/tabs"
)

func TestExecuteRunsSummarizeStepWithoutPersistingMemory(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Atlas summary"}}]}`))
	}))
	defer server.Close()

	config := sidebar.FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{{
			ID:        "primary",
			Provider:  "openai",
			Model:     "gpt-5.4",
			BaseURL:   server.URL,
			APIKeyEnv: "OPENAI_API_KEY",
		}},
	})

	result, err := Execute(config, tabs.PageContext{
		ID:         "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Text:       "Atlas context",
		CapturedAt: "2026-04-08T14:00:00Z",
	}, nil, Plan{
		Steps: []Step{{
			ID:                   "suggest-summarize_page",
			Kind:                 "sidebar_summarize",
			Title:                "Summarize this page",
			RequiresConfirmation: true,
		}},
	}, "suggest-summarize_page", true, ExecutionActions{})
	if err != nil {
		t.Fatalf("execute summarize failed: %v", err)
	}
	if !result.Executed || result.MemoryPersisted {
		t.Fatalf("unexpected result flags: %+v", result)
	}
	if result.Result != "Atlas summary" || result.Provider != "openai" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestExecuteRejectsMissingConfirmation(t *testing.T) {
	_, err := Execute(sidebar.Config{}, tabs.PageContext{}, nil, Plan{
		Steps: []Step{{ID: "suggest-summarize_page", Kind: "sidebar_summarize"}},
	}, "suggest-summarize_page", false, ExecutionActions{})
	if !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteRunsRelatedTabStep(t *testing.T) {
	activatedTabID := ""
	result, err := Execute(sidebar.Config{}, tabs.PageContext{ID: "tab-1"}, nil, Plan{
		Steps: []Step{{
			ID:                   "recommend-related-tab-tab-2",
			Kind:                 "related_tab",
			Title:                "Atlas docs",
			TabID:                "tab-2",
			URL:                  "https://chatgpt.com/docs",
			RequiresConfirmation: true,
		}},
	}, "recommend-related-tab-tab-2", true, ExecutionActions{
		ActivateTab: func(tabID string) error {
			activatedTabID = tabID
			return nil
		},
	})
	if err != nil {
		t.Fatalf("execute related_tab failed: %v", err)
	}
	if activatedTabID != "tab-2" {
		t.Fatalf("expected activated tab id tab-2, got %q", activatedTabID)
	}
	if !result.Executed || result.ActivatedTabID != "tab-2" || result.StepKind != "related_tab" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.MemoryPersisted || result.Rollback != "manual_reactivate_previous_tab_if_needed" {
		t.Fatalf("unexpected side effect flags: %+v", result)
	}
}

func TestExecuteRejectsRelatedTabStepWhenTabIDIsMissing(t *testing.T) {
	_, err := Execute(sidebar.Config{}, tabs.PageContext{ID: "tab-1"}, nil, Plan{
		Steps: []Step{{
			ID:                   "recommend-related-tab-tab-2",
			Kind:                 "related_tab",
			Title:                "Atlas docs",
			RequiresConfirmation: true,
		}},
	}, "recommend-related-tab-tab-2", true, ExecutionActions{
		ActivateTab: func(tabID string) error {
			return nil
		},
	})
	if !errors.Is(err, ErrStepTabIDRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteRejectsRelatedTabStepWhenActionIsMissing(t *testing.T) {
	_, err := Execute(sidebar.Config{}, tabs.PageContext{ID: "tab-1"}, nil, Plan{
		Steps: []Step{{
			ID:                   "recommend-related-tab-tab-2",
			Kind:                 "related_tab",
			Title:                "Atlas docs",
			TabID:                "tab-2",
			RequiresConfirmation: true,
		}},
	}, "recommend-related-tab-tab-2", true, ExecutionActions{})
	if !errors.Is(err, ErrStepActionRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteRejectsPreviewOnlyStep(t *testing.T) {
	_, err := Execute(sidebar.Config{}, tabs.PageContext{ID: "tab-1"}, nil, Plan{
		Steps: []Step{{
			ID:                   "recommend-memory-relevant-page-capture",
			Kind:                 "memory_snippet",
			Title:                "Relevant page capture",
			RequiresConfirmation: true,
		}},
	}, "recommend-memory-relevant-page-capture", true, ExecutionActions{})
	if !errors.Is(err, ErrStepNotExecutable) {
		t.Fatalf("unexpected error: %v", err)
	}
}
