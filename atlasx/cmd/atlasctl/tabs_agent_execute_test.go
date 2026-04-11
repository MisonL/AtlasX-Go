//go:build darwin

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
	"atlasx/internal/tabs"
)

func TestTabsAgentExecuteRunsSummarizeStep(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Atlas summary"}}]}`))
	}))
	defer server.Close()

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{{
			ID:        "primary",
			Provider:  "openai",
			Model:     "gpt-5.4",
			BaseURL:   server.URL,
			APIKeyEnv: "OPENAI_API_KEY",
		}},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas context",
			CapturedAt: "2026-04-08T14:10:00Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
			{ID: "tab-2", Type: "page", Title: "Atlas docs", URL: "https://chatgpt.com/docs"},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "agent-execute", "--confirm", "tab-1", "suggest-summarize_page"})
	})
	if err != nil {
		t.Fatalf("run tabs agent-execute failed: %v", err)
	}
	for _, fragment := range []string{
		"step_id=suggest-summarize_page",
		"step_kind=sidebar_summarize",
		"executed=true",
		"confirmed=true",
		"provider=openai",
		"memory_persisted=false",
		`result="Atlas summary"`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}

	events, err := loadCommandMemoryEvents(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no memory events, got %+v", events)
	}

	runtimeState, err := sidebar.LoadRuntimeState(paths)
	if err != nil {
		t.Fatalf("load runtime state failed: %v", err)
	}
	if runtimeState.LastTraceID == "" || runtimeState.LastError != "" {
		t.Fatalf("unexpected runtime state: %+v", runtimeState)
	}
}

func TestTabsAgentExecuteRejectsMissingConfirm(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:    "tab-1",
			Title: "Atlas",
			URL:   "https://chatgpt.com/atlas",
			Text:  "Atlas context",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
		},
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "agent-execute", "tab-1", "suggest-summarize_page"})
	})
	if err == nil {
		t.Fatal("expected tabs agent-execute to fail without confirm")
	}
	if !strings.Contains(err.Error(), "agent execute requires explicit confirmation") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsAgentExecuteRunsRelatedTabStep(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	client := &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas context",
			CapturedAt: "2026-04-08T14:30:00Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
			{ID: "tab-2", Type: "page", Title: "Atlas docs", URL: "https://chatgpt.com/docs"},
		},
	}
	restoreCommandTabsClient(t, client)

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "agent-execute", "--confirm", "tab-1", "recommend-related-tab-tab-2"})
	})
	if err != nil {
		t.Fatalf("run tabs agent-execute related_tab failed: %v", err)
	}
	for _, fragment := range []string{
		"step_id=recommend-related-tab-tab-2",
		"step_kind=related_tab",
		"activated_tab_id=tab-2",
		"executed=true",
		"confirmed=true",
		"memory_persisted=false",
		"rollback=manual_reactivate_previous_tab_if_needed",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
	if client.activatedTargetID != "tab-2" {
		t.Fatalf("expected activated target id tab-2, got %q", client.activatedTargetID)
	}

	events, err := loadCommandMemoryEvents(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no memory events, got %+v", events)
	}
}

func TestTabsAgentExecuteRunsMemorySnippetStep(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Memory relevance answer"}}]}`))
	}))
	defer server.Close()

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{{
			ID:        "primary",
			Provider:  "openai",
			Model:     "gpt-5.4",
			BaseURL:   server.URL,
			APIKeyEnv: "OPENAI_API_KEY",
		}},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}
	if err := memory.AppendPageCapture(paths, memory.PageCaptureInput{
		OccurredAt: "2026-04-08T15:20:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
	}); err != nil {
		t.Fatalf("append page capture failed: %v", err)
	}

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas context",
			CapturedAt: "2026-04-08T15:21:00Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "agent-execute", "--confirm", "tab-1", "recommend-memory-relevant-page-capture"})
	})
	if err != nil {
		t.Fatalf("run tabs agent-execute memory_snippet failed: %v", err)
	}
	for _, fragment := range []string{
		"step_id=recommend-memory-relevant-page-capture",
		"step_kind=memory_snippet",
		"executed=true",
		"confirmed=true",
		"provider=openai",
		"memory_persisted=false",
		`result="Memory relevance answer"`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}

	events, err := loadCommandMemoryEvents(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one existing memory event without new writes, got %+v", events)
	}
}

func TestTabsAgentExecuteRejectsMissingStep(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := memory.AppendPageCapture(paths, memory.PageCaptureInput{
		OccurredAt: "2026-04-08T14:31:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
	}); err != nil {
		t.Fatalf("append page capture failed: %v", err)
	}

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas context",
			CapturedAt: "2026-04-08T14:31:10Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
		},
	})

	_, err = captureStdout(t, func() error {
		return run([]string{"tabs", "agent-execute", "--confirm", "tab-1", "recommend-step-not-present"})
	})
	if err == nil {
		t.Fatal("expected tabs agent-execute to fail for missing step")
	}
	if !strings.Contains(err.Error(), "step was not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsAgentExecuteRunsBoundedBatchSteps(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Memory relevance answer"}}]}`))
	}))
	defer server.Close()

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{{
			ID:        "primary",
			Provider:  "openai",
			Model:     "gpt-5.4",
			BaseURL:   server.URL,
			APIKeyEnv: "OPENAI_API_KEY",
		}},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}
	if err := memory.AppendPageCapture(paths, memory.PageCaptureInput{
		OccurredAt: "2026-04-08T16:10:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
	}); err != nil {
		t.Fatalf("append page capture failed: %v", err)
	}

	client := &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas context",
			CapturedAt: "2026-04-08T16:11:00Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
			{ID: "tab-2", Type: "page", Title: "Atlas docs", URL: "https://chatgpt.com/docs"},
		},
	}
	restoreCommandTabsClient(t, client)

	output, err := captureStdout(t, func() error {
		return run([]string{
			"tabs", "agent-execute", "--confirm", "--max-steps", "2",
			"tab-1",
			"recommend-related-tab-tab-2",
			"recommend-memory-relevant-page-capture",
		})
	})
	if err != nil {
		t.Fatalf("run tabs agent-execute batch failed: %v", err)
	}
	for _, fragment := range []string{
		"requested=2",
		"executed=2",
		"stopped=false",
		"max_steps=2",
		"batch_index=0",
		"step_kind=related_tab",
		"batch_index=1",
		"step_kind=memory_snippet",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
	if client.activatedTargetID != "tab-2" {
		t.Fatalf("expected activated target id tab-2, got %q", client.activatedTargetID)
	}
}

func TestTabsAgentExecuteRejectsBatchBeyondMaxSteps(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas context",
			CapturedAt: "2026-04-08T16:20:00Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
			{ID: "tab-2", Type: "page", Title: "Atlas docs", URL: "https://chatgpt.com/docs"},
		},
	})

	_, err := captureStdout(t, func() error {
		return run([]string{
			"tabs", "agent-execute", "--confirm", "--max-steps", "1",
			"tab-1",
			"recommend-related-tab-tab-2",
			"recommend-step-not-present",
		})
	})
	if err == nil {
		t.Fatal("expected tabs agent-execute to fail when requested steps exceed max_steps")
	}
	if !strings.Contains(err.Error(), "exceeds max_steps") {
		t.Fatalf("unexpected error: %v", err)
	}
}
