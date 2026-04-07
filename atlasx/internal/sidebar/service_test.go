package sidebar

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"atlasx/internal/settings"
	"atlasx/internal/tabs"
)

func TestStatusWithoutProvider(t *testing.T) {
	status := FromSettings(settings.Config{}).Status()
	if status.Configured {
		t.Fatal("expected unconfigured sidebar status")
	}
	if status.Reason == "" {
		t.Fatal("expected explicit reason")
	}
}

func TestStatusWithIncompleteConfig(t *testing.T) {
	status := FromSettings(settings.Config{
		SidebarProvider: "openai",
	}).Status()
	if !status.Configured {
		t.Fatal("expected configured status")
	}
	if status.Reason != "sidebar qa config is incomplete" {
		t.Fatalf("unexpected reason: %s", status.Reason)
	}
}

func TestStatusWithProviderRegistry(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	status := FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   "https://api.openai.com/v1",
				APIKeyEnv: "OPENAI_API_KEY",
			},
			{
				ID:        "backup",
				Provider:  "openrouter",
				Model:     "openai/gpt-5",
				BaseURL:   "https://openrouter.ai/api/v1",
				APIKeyEnv: "OPENROUTER_API_KEY",
			},
		},
	}).Status()

	if status.DefaultProvider != "primary" {
		t.Fatalf("unexpected default provider: %s", status.DefaultProvider)
	}
	if len(status.Providers) != 2 {
		t.Fatalf("unexpected providers: %+v", status.Providers)
	}
	if status.APIKeyEnv != "OPENAI_API_KEY" {
		t.Fatalf("unexpected api key env: %s", status.APIKeyEnv)
	}
	if !status.Ready {
		t.Fatalf("expected ready status: %+v", status)
	}
}

func TestStatusPrefersRegistryOverLegacyFields(t *testing.T) {
	t.Setenv("OPENROUTER_API_KEY", "test-key")

	status := FromSettings(settings.Config{
		SidebarProvider:        "legacy-openai",
		SidebarModel:           "legacy-model",
		SidebarBaseURL:         "https://legacy.example.com/v1",
		SidebarDefaultProvider: "registry-openrouter",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "registry-openrouter",
				Provider:  "openrouter",
				Model:     "openai/gpt-5",
				BaseURL:   "https://openrouter.ai/api/v1",
				APIKeyEnv: "OPENROUTER_API_KEY",
			},
		},
	}).Status()

	if status.DefaultProvider != "registry-openrouter" {
		t.Fatalf("unexpected default provider: %s", status.DefaultProvider)
	}
	if status.Provider != "openrouter" {
		t.Fatalf("unexpected provider: %s", status.Provider)
	}
	if len(status.Providers) != 1 || status.Providers[0].ID != "registry-openrouter" {
		t.Fatalf("unexpected providers: %+v", status.Providers)
	}
}

func TestAskRejectsUnconfiguredBackend(t *testing.T) {
	_, err := Config{}.Ask(AskRequest{
		TabID:    "tab-1",
		Question: "summarize this page",
	}, tabs.PageContext{})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAskWithOpenAICompatibleProvider(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body failed: %v", err)
		}
		if err := json.Unmarshal(body, &capturedBody); err != nil {
			t.Fatalf("decode body failed: %v", err)
		}
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Atlas answer"}}]}`))
	}))
	defer server.Close()

	config := FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   server.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	})
	response, err := config.Ask(AskRequest{
		TabID:    "tab-1",
		Question: "summarize this page",
	}, tabs.PageContext{
		ID:            "tab-1",
		Title:         "Atlas",
		URL:           "https://chatgpt.com/atlas",
		Text:          "Atlas page text",
		CapturedAt:    "2026-04-06T12:00:00Z",
		TextLength:    15,
		TextLimit:     4096,
		TextTruncated: false,
	})
	if err != nil {
		t.Fatalf("ask failed: %v", err)
	}
	if response.Answer != "Atlas answer" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.Provider != "openai" || response.Model != "gpt-5.4" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if !strings.Contains(response.ContextSummary, `title="Atlas"`) {
		t.Fatalf("unexpected context summary: %s", response.ContextSummary)
	}

	messages, ok := capturedBody["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("unexpected request body: %+v", capturedBody)
	}
	userMessage, ok := messages[1].(map[string]any)
	if !ok {
		t.Fatalf("unexpected request body: %+v", capturedBody)
	}
	content, _ := userMessage["content"].(string)
	if !strings.Contains(content, "Atlas page text") || !strings.Contains(content, "https://chatgpt.com/atlas") {
		t.Fatalf("tab context not included in request: %s", content)
	}
}

func TestAskWithMemoryIncludesRelevantHistory(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body failed: %v", err)
		}
		if err := json.Unmarshal(body, &capturedBody); err != nil {
			t.Fatalf("decode body failed: %v", err)
		}
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Atlas answer"}}]}`))
	}))
	defer server.Close()

	config := FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   server.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	})
	_, err := config.AskWithMemory(AskRequest{
		TabID:    "tab-1",
		Question: "summarize this page",
	}, tabs.PageContext{
		ID:         "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Text:       "Atlas page text",
		CapturedAt: "2026-04-06T12:00:00Z",
	}, []string{
		`qa_turn occurred_at=2026-04-06T11:00:00Z title="Atlas" url=https://chatgpt.com/atlas question="what is atlas" answer="Atlas is memory-aware." cited_urls=https://chatgpt.com/atlas`,
	})
	if err != nil {
		t.Fatalf("ask with memory failed: %v", err)
	}

	messages, ok := capturedBody["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("unexpected request body: %+v", capturedBody)
	}
	userMessage, ok := messages[1].(map[string]any)
	if !ok {
		t.Fatalf("unexpected request body: %+v", capturedBody)
	}
	content, _ := userMessage["content"].(string)
	if !strings.Contains(content, "Relevant Memory:") {
		t.Fatalf("missing memory section: %s", content)
	}
	if !strings.Contains(content, "Atlas is memory-aware.") {
		t.Fatalf("missing memory snippet: %s", content)
	}
}

func TestSummarizeWithOpenAICompatibleProvider(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body failed: %v", err)
		}
		if err := json.Unmarshal(body, &capturedBody); err != nil {
			t.Fatalf("decode body failed: %v", err)
		}
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Atlas summary"}}]}`))
	}))
	defer server.Close()

	config := FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   server.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	})
	response, err := config.SummarizeWithMemory(SummarizeRequest{
		TabID: "tab-1",
	}, tabs.PageContext{
		ID:            "tab-1",
		Title:         "Atlas",
		URL:           "https://chatgpt.com/atlas",
		Text:          "Atlas page text",
		CapturedAt:    "2026-04-07T10:00:00Z",
		TextLength:    15,
		TextLimit:     4096,
		TextTruncated: false,
	}, []string{
		`qa_turn occurred_at=2026-04-07T09:59:00Z title="Atlas" url=https://chatgpt.com/atlas question="what is atlas" answer="Atlas is memory-aware." cited_urls=https://chatgpt.com/atlas`,
	})
	if err != nil {
		t.Fatalf("summarize failed: %v", err)
	}
	if response.Summary != "Atlas summary" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.Provider != "openai" || response.Model != "gpt-5.4" {
		t.Fatalf("unexpected response: %+v", response)
	}

	messages, ok := capturedBody["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("unexpected request body: %+v", capturedBody)
	}
	userMessage, ok := messages[1].(map[string]any)
	if !ok {
		t.Fatalf("unexpected request body: %+v", capturedBody)
	}
	content, _ := userMessage["content"].(string)
	if !strings.Contains(content, PageSummaryQuestion) {
		t.Fatalf("missing summary prompt: %s", content)
	}
	if !strings.Contains(content, "Relevant Memory:") {
		t.Fatalf("missing memory section: %s", content)
	}
}

func TestAskAllowsProviderOverride(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")
	t.Setenv("OPENROUTER_API_KEY", "router-key")

	openAIServerCalled := false
	openAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		openAIServerCalled = true
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"wrong server"}}]}`))
	}))
	defer openAIServer.Close()

	openRouterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HTTP-Referer") != "https://atlasx.local" {
			t.Fatalf("missing openrouter header: %s", r.Header.Get("HTTP-Referer"))
		}
		_, _ = w.Write([]byte(`{"model":"openai/gpt-5","choices":[{"message":{"content":"OpenRouter answer"}}]}`))
	}))
	defer openRouterServer.Close()

	config := FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   openAIServer.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
			{
				ID:        "backup",
				Provider:  "openrouter",
				Model:     "openai/gpt-5",
				BaseURL:   openRouterServer.URL,
				APIKeyEnv: "OPENROUTER_API_KEY",
			},
		},
	})
	response, err := config.Ask(AskRequest{
		TabID:      "tab-1",
		Question:   "summarize this page",
		ProviderID: "backup",
	}, tabs.PageContext{
		ID:    "tab-1",
		Title: "Atlas",
		URL:   "https://chatgpt.com/atlas",
		Text:  "Atlas page text",
	})
	if err != nil {
		t.Fatalf("ask failed: %v", err)
	}
	if response.Provider != "openrouter" || response.Model != "openai/gpt-5" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.Answer != "OpenRouter answer" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if openAIServerCalled {
		t.Fatal("expected request override to skip default provider")
	}
}

func TestAskRejectsUnknownProviderID(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	config := FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   "https://api.openai.com/v1",
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	})
	_, err := config.Ask(AskRequest{
		TabID:      "tab-1",
		Question:   "summarize this page",
		ProviderID: "missing",
	}, tabs.PageContext{ID: "tab-1"})
	if !errors.Is(err, ErrProviderNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAskReturnsProviderFailure(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"message":"upstream failed"}}`))
	}))
	defer server.Close()

	config := FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   server.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	})
	_, err := config.Ask(AskRequest{
		TabID:    "tab-1",
		Question: "summarize this page",
	}, tabs.PageContext{
		ID:    "tab-1",
		Title: "Atlas",
		URL:   "https://chatgpt.com/atlas",
	})
	if !errors.Is(err, ErrProviderFailed) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAskRejectsTokenBudgetOverflow(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	config := FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   "https://api.openai.com/v1",
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	})
	_, err := config.Ask(AskRequest{
		TabID:    "tab-1",
		Question: "summarize this page",
	}, tabs.PageContext{
		ID:   "tab-1",
		Text: strings.Repeat("A", providerTokenBudget*10),
	})
	if !errors.Is(err, ErrTokenBudgetExceeded) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildPromptDropsMemoryWhenBudgetIsTight(t *testing.T) {
	prompt := buildPrompt("summarize this page", tabs.PageContext{
		ID:   "tab-1",
		Text: strings.Repeat("A", providerTokenBudget*4),
	}, []string{
		"qa_turn occurred_at=2026-04-06T11:00:00Z title=\"Atlas\" url=https://chatgpt.com/atlas question=\"what is atlas\" answer=\"Atlas is memory-aware.\" cited_urls=https://chatgpt.com/atlas",
	})

	if strings.Contains(prompt, "Relevant Memory:") {
		t.Fatalf("expected prompt to drop memory section when budget is tight: %s", prompt)
	}
}

func TestAskRetriesTimeoutOnceAndReturnsProviderFailure(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		time.Sleep(providerRequestTimeout + 200*time.Millisecond)
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"late answer"}}]}`))
	}))
	defer server.Close()

	config := FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   server.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	})
	_, err := config.Ask(AskRequest{
		TabID:    "tab-1",
		Question: "summarize this page",
	}, tabs.PageContext{
		ID:    "tab-1",
		Title: "Atlas",
		URL:   "https://chatgpt.com/atlas",
		Text:  "Atlas context",
	})
	if !errors.Is(err, ErrProviderFailed) {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != providerRetryAttempts+1 {
		t.Fatalf("unexpected attempt count: %d", attempts)
	}
}
