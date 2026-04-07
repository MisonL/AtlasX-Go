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

func TestSidebarStatusOutputsUnconfiguredReason(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"sidebar", "status"})
	})
	if err != nil {
		t.Fatalf("run sidebar status failed: %v", err)
	}

	for _, fragment := range []string{
		"AtlasX Sidebar",
		"configured=false",
		"ready=false",
		"provider_count=0",
		"reason=sidebar qa provider is not configured",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestSidebarStatusOutputsConfiguredProvider(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
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
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	output, err := captureStdout(t, func() error {
		return run([]string{"sidebar", "status"})
	})
	if err != nil {
		t.Fatalf("run sidebar status failed: %v", err)
	}

	for _, fragment := range []string{
		"configured=true",
		"ready=true",
		"default_provider=primary",
		"provider=openai",
		"model=gpt-5.4",
		"api_key_env=OPENAI_API_KEY",
		"provider_count=1",
		"provider[0].id=primary",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestSidebarSummarizeOutputsStructuredSummary(t *testing.T) {
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
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   server.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-07T11:00:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"sidebar", "summarize", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run sidebar summarize failed: %v", err)
	}
	for _, fragment := range []string{
		`summary="Atlas summary"`,
		"provider=openai",
		"model=gpt-5.4",
		`context_summary="title=\"Atlas\"`,
		"trace_id=",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}

	events, err := loadCommandMemoryEvents(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 1 || events[0].Kind != memory.EventKindQATurn {
		t.Fatalf("unexpected memory events: %+v", events)
	}
	if events[0].Question != sidebar.PageSummaryQuestion || events[0].Answer != "Atlas summary" {
		t.Fatalf("unexpected memory event: %+v", events[0])
	}

	runtimeState, err := sidebar.LoadRuntimeState(paths)
	if err != nil {
		t.Fatalf("load runtime state failed: %v", err)
	}
	if runtimeState.LastTraceID == "" || runtimeState.LastError != "" {
		t.Fatalf("unexpected runtime state: %+v", runtimeState)
	}
}

func TestSidebarSummarizeWritesRuntimeErrorOnProviderFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"message":"provider failed"}}`))
	}))
	defer server.Close()

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
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
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:    "tab-1",
			Title: "Atlas",
			URL:   "https://chatgpt.com/atlas",
			Text:  "Atlas context",
		},
	})

	_, err = captureStdout(t, func() error {
		return run([]string{"sidebar", "summarize", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected sidebar summarize to fail")
	}
	if !strings.Contains(err.Error(), "sidebar qa provider request failed") {
		t.Fatalf("unexpected error: %v", err)
	}

	runtimeState, err := sidebar.LoadRuntimeState(paths)
	if err != nil {
		t.Fatalf("load runtime state failed: %v", err)
	}
	if runtimeState.LastTraceID == "" || runtimeState.LastError == "" {
		t.Fatalf("unexpected runtime state: %+v", runtimeState)
	}
}

func TestSidebarAskOutputsStructuredAnswer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Atlas answer"}}]}`))
	}))
	defer server.Close()

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
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
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-07T11:05:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"sidebar", "ask", "tab-1", "What", "is", "this", "page", "about?"})
	})
	if err != nil {
		t.Fatalf("run sidebar ask failed: %v", err)
	}
	for _, fragment := range []string{
		`answer="Atlas answer"`,
		"provider=openai",
		"model=gpt-5.4",
		`context_summary="title=\"Atlas\"`,
		"trace_id=",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}

	events, err := loadCommandMemoryEvents(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 1 || events[0].Kind != memory.EventKindQATurn {
		t.Fatalf("unexpected memory events: %+v", events)
	}
	if events[0].Question != "What is this page about?" || events[0].Answer != "Atlas answer" {
		t.Fatalf("unexpected memory event: %+v", events[0])
	}

	runtimeState, err := sidebar.LoadRuntimeState(paths)
	if err != nil {
		t.Fatalf("load runtime state failed: %v", err)
	}
	if runtimeState.LastTraceID == "" || runtimeState.LastError != "" {
		t.Fatalf("unexpected runtime state: %+v", runtimeState)
	}
}

func TestSidebarAskWritesRuntimeErrorOnProviderFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"message":"provider failed"}}`))
	}))
	defer server.Close()

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
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
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:    "tab-1",
			Title: "Atlas",
			URL:   "https://chatgpt.com/atlas",
			Text:  "Atlas context",
		},
	})

	_, err = captureStdout(t, func() error {
		return run([]string{"sidebar", "ask", "tab-1", "What", "is", "this", "page", "about?"})
	})
	if err == nil {
		t.Fatal("expected sidebar ask to fail")
	}
	if !strings.Contains(err.Error(), "sidebar qa provider request failed") {
		t.Fatalf("unexpected error: %v", err)
	}

	runtimeState, err := sidebar.LoadRuntimeState(paths)
	if err != nil {
		t.Fatalf("load runtime state failed: %v", err)
	}
	if runtimeState.LastTraceID == "" || runtimeState.LastError == "" {
		t.Fatalf("unexpected runtime state: %+v", runtimeState)
	}
}
