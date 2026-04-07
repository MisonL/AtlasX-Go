package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
	"atlasx/internal/tabs"
)

func TestSidebarSelectionAskOutputsStructuredAnswerWithExplicitSelection(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Selected answer"}}]}`))
	}))
	defer server.Close()

	paths := saveSidebarSelectionTestConfig(t, server.URL)
	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-07T11:10:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{
			"sidebar",
			"selection-ask",
			"--selection-text", "Atlas selected text",
			"tab-1",
			"What", "does", "this", "mean?",
		})
	})
	if err != nil {
		t.Fatalf("run sidebar selection-ask failed: %v", err)
	}
	for _, fragment := range []string{
		`answer="Selected answer"`,
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
	if len(events) != 1 {
		t.Fatalf("unexpected memory events: %+v", events)
	}

	expectedQuestion, err := sidebar.BuildSelectionQuestion("Atlas selected text", "What does this mean?")
	if err != nil {
		t.Fatalf("build selection question failed: %v", err)
	}
	if events[0].Question != expectedQuestion || events[0].Answer != "Selected answer" {
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

func TestSidebarSelectionAskCapturesBrowserSelectionWhenFlagIsAbsent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Captured answer"}}]}`))
	}))
	defer server.Close()

	paths := saveSidebarSelectionTestConfig(t, server.URL)
	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-07T11:15:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
		selection: tabs.SelectionContext{
			ID:                     "tab-1",
			Title:                  "Atlas",
			URL:                    "https://chatgpt.com/atlas",
			SelectionText:          "Captured browser selection",
			CapturedAt:             "2026-04-07T11:15:01Z",
			SelectionPresent:       true,
			SelectionTextLength:    26,
			SelectionTextLimit:     1024,
			SelectionTextTruncated: false,
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"sidebar", "selection-ask", "tab-1", "Why", "is", "this", "important?"})
	})
	if err != nil {
		t.Fatalf("run sidebar selection-ask failed: %v", err)
	}
	if !strings.Contains(output, `answer="Captured answer"`) {
		t.Fatalf("unexpected output: %s", output)
	}

	events, err := loadCommandMemoryEvents(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	expectedQuestion, err := sidebar.BuildSelectionQuestion("Captured browser selection", "Why is this important?")
	if err != nil {
		t.Fatalf("build selection question failed: %v", err)
	}
	if len(events) != 1 || events[0].Question != expectedQuestion {
		t.Fatalf("unexpected memory events: %+v", events)
	}
}

func TestSidebarSelectionAskWritesRuntimeErrorOnProviderFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"message":"provider failed"}}`))
	}))
	defer server.Close()

	paths := saveSidebarSelectionTestConfig(t, server.URL)
	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:    "tab-1",
			Title: "Atlas",
			URL:   "https://chatgpt.com/atlas",
			Text:  "Atlas context",
		},
		selection: tabs.SelectionContext{
			ID:               "tab-1",
			Title:            "Atlas",
			URL:              "https://chatgpt.com/atlas",
			SelectionText:    "Captured browser selection",
			SelectionPresent: true,
		},
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"sidebar", "selection-ask", "tab-1", "Why", "is", "this", "important?"})
	})
	if err == nil {
		t.Fatal("expected sidebar selection-ask to fail")
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

func TestSidebarSelectionAskFailsWhenBrowserSelectionIsEmpty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	paths := saveSidebarSelectionTestConfig(t, "https://example.invalid")
	restoreCommandTabsClient(t, &stubCommandTabsClient{
		selection: tabs.SelectionContext{
			ID:                 "tab-1",
			Title:              "Atlas",
			URL:                "https://chatgpt.com/atlas",
			CapturedAt:         "2026-04-07T11:20:00Z",
			SelectionPresent:   false,
			SelectionTextLimit: 1024,
		},
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"sidebar", "selection-ask", "tab-1", "Why", "is", "this", "important?"})
	})
	if err == nil {
		t.Fatal("expected sidebar selection-ask to fail")
	}
	if !strings.Contains(err.Error(), "selection_text is required when page selection is empty") {
		t.Fatalf("unexpected error: %v", err)
	}

	runtimeState, err := sidebar.LoadRuntimeState(paths)
	if err != nil {
		t.Fatalf("load runtime state failed: %v", err)
	}
	if runtimeState.LastTraceID == "" || runtimeState.LastError == "" {
		t.Fatalf("unexpected runtime state: %+v", runtimeState)
	}

	events, err := loadCommandMemoryEvents(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no memory events, got %+v", events)
	}
}

func saveSidebarSelectionTestConfig(t *testing.T, baseURL string) macos.Paths {
	t.Helper()

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
				BaseURL:   baseURL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}
	return paths
}
