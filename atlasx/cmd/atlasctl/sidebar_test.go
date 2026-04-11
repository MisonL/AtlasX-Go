package main

import (
	"errors"
	"io"
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

func TestSidebarSetProviderWritesRegistryAndShowsStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{
			"sidebar", "set-provider",
			"--id", "primary",
			"--provider", "openai",
			"--model", "gpt-5.4",
			"--base-url", "https://api.openai.com/v1",
			"--api-key-env", "ATLASX_TEST_OPENAI_API_KEY",
			"--default",
		})
	})
	if err != nil {
		t.Fatalf("run sidebar set-provider failed: %v", err)
	}

	for _, fragment := range []string{
		"configured=true",
		"ready=false",
		"default_provider=primary",
		"provider=openai",
		"model=gpt-5.4",
		"api_key_env=ATLASX_TEST_OPENAI_API_KEY",
		"provider_count=1",
		"reason=sidebar qa api key env ATLASX_TEST_OPENAI_API_KEY is not set",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	config, err := settings.NewStore(paths.ConfigFile).Load()
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if config.SidebarDefaultProvider != "primary" {
		t.Fatalf("unexpected default provider: %+v", config)
	}
	if len(config.SidebarProviders) != 1 {
		t.Fatalf("unexpected providers: %+v", config.SidebarProviders)
	}
	if config.SidebarProviders[0].APIKeyEnv != "ATLASX_TEST_OPENAI_API_KEY" {
		t.Fatalf("unexpected provider env: %+v", config.SidebarProviders[0])
	}
}

func TestSidebarSetProviderRejectsMissingRequiredFields(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{
			"sidebar", "set-provider",
			"--id", "primary",
			"--provider", "openai",
			"--model", "gpt-5.4",
			"--base-url", "https://api.openai.com/v1",
		})
	})
	if err == nil {
		t.Fatal("expected sidebar set-provider to fail")
	}
	if !strings.Contains(err.Error(), settings.ErrSidebarProviderAPIKeyEnvRequired.Error()) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSidebarSetProviderRejectsWhitespaceOnlyID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{
			"sidebar", "set-provider",
			"--id", "   ",
			"--provider", "openai",
			"--model", "gpt-5.4",
			"--base-url", "https://api.openai.com/v1",
			"--api-key-env", "OPENAI_API_KEY",
		})
	})
	if err == nil {
		t.Fatal("expected sidebar set-provider to fail")
	}
	if !strings.Contains(err.Error(), settings.ErrSidebarProviderIDRequired.Error()) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSidebarSummarizeRejectsExtraArguments(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"sidebar", "summarize", "tab-1", "extra"})
	})
	if err == nil {
		t.Fatal("expected sidebar summarize to fail")
	}
	if !strings.Contains(err.Error(), "sidebar summarize does not accept extra arguments") {
		t.Fatalf("unexpected error: %v", err)
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

func TestSidebarSummarizeHidesMemoryWhenPageVisibilityDisabled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	var capturedRequestBody string
	handlerErrCh := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			handlerErrCh <- err
			http.Error(w, "read request body failed", http.StatusInternalServerError)
			return
		}
		capturedRequestBody = string(body)
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Atlas summary"}}]}`))
	}))
	defer server.Close()

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := memory.AppendQATurn(paths, memory.QATurnInput{
		OccurredAt: "2026-04-07T11:59:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "what was asked before",
		Answer:     "Hidden memory answer",
		TraceID:    "trace-1",
	}); err != nil {
		t.Fatalf("append qa turn failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{
		MemoryPageVisibility:   settings.Bool(false),
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
	select {
	case handlerErr := <-handlerErrCh:
		t.Fatalf("read request body failed: %v", handlerErr)
	default:
	}
	if !strings.Contains(output, `summary="Atlas summary"`) {
		t.Fatalf("unexpected output: %s", output)
	}
	if strings.Contains(capturedRequestBody, "Hidden memory answer") {
		t.Fatalf("expected page visibility control to hide memory, got %s", capturedRequestBody)
	}
}

func TestSidebarSummarizeHandlerReadErrorReportedInMainGoroutine(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	handlerErrCh := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerErrCh <- errors.New("synthetic handler read failure")
		http.Error(w, "internal error", http.StatusInternalServerError)
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
	select {
	case handlerErr := <-handlerErrCh:
		if handlerErr.Error() != "synthetic handler read failure" {
			t.Fatalf("unexpected handler error: %v", handlerErr)
		}
	default:
		t.Fatal("expected handler error to be reported")
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

func TestSidebarAskSkipsMemoryEventWhenPersistenceDisabled(t *testing.T) {
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
		MemoryPersistEnabled:   settings.Bool(false),
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

	_, err = captureStdout(t, func() error {
		return run([]string{"sidebar", "ask", "tab-1", "What", "is", "this", "page", "about?"})
	})
	if err != nil {
		t.Fatalf("run sidebar ask failed: %v", err)
	}

	events, err := loadCommandMemoryEvents(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no memory events, got %+v", events)
	}
}
