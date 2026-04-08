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
	}, "suggest-summarize_page", true)
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
	}, "suggest-summarize_page", false)
	if !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExecuteRejectsPreviewOnlyStep(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	config := sidebar.FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{{
			ID:        "primary",
			Provider:  "openai",
			Model:     "gpt-5.4",
			BaseURL:   "https://example.com",
			APIKeyEnv: "OPENAI_API_KEY",
		}},
	})

	_, err := Execute(config, tabs.PageContext{ID: "tab-1"}, nil, Plan{
		Steps: []Step{{
			ID:                   "recommend-related-tab-tab-2",
			Kind:                 "related_tab",
			Title:                "Atlas docs",
			RequiresConfirmation: true,
		}},
	}, "recommend-related-tab-tab-2", true)
	if !errors.Is(err, ErrStepNotExecutable) {
		t.Fatalf("unexpected error: %v", err)
	}
}
