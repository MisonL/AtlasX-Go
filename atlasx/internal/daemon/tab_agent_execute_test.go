package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/sidebar"
	"atlasx/internal/tabs"
)

func TestTabAgentExecuteRunsSummarizeStep(t *testing.T) {
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
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{}); err != nil {
		t.Fatalf("save config failed: %v", err)
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

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas context",
			CapturedAt: "2026-04-08T14:20:00Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
			{ID: "tab-2", Type: "page", Title: "Atlas docs", URL: "https://chatgpt.com/docs"},
		},
	})

	body := bytes.NewBufferString(`{"id":"tab-1","step_id":"suggest-summarize_page","confirm":true}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/agent-execute", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{
		`"step_id":"suggest-summarize_page"`,
		`"step_kind":"sidebar_summarize"`,
		`"executed":true`,
		`"memory_persisted":false`,
		`"result":"Atlas summary"`,
	} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected body to contain %q, got %s", fragment, recorder.Body.String())
		}
	}

	events, err := memory.Load(paths)
	if err != nil && !os.IsNotExist(err) {
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

func TestTabAgentExecuteRejectsMissingConfirm(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:    "tab-1",
			Title: "Atlas",
			URL:   "https://chatgpt.com/atlas",
			Text:  "Atlas context",
		},
		targets: []tabs.Target{{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"}},
	})

	body := bytes.NewBufferString(`{"id":"tab-1","step_id":"suggest-summarize_page","confirm":false}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/agent-execute", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "explicit confirmation") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestTabAgentExecuteRunsRelatedTabStep(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	client := &stubTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas context",
			CapturedAt: "2026-04-08T14:32:00Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
			{ID: "tab-2", Type: "page", Title: "Atlas docs", URL: "https://chatgpt.com/docs"},
		},
	}
	restoreDaemonHooks(t, client)

	body := bytes.NewBufferString(`{"id":"tab-1","step_id":"recommend-related-tab-tab-2","confirm":true}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/agent-execute", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{
		`"step_id":"recommend-related-tab-tab-2"`,
		`"step_kind":"related_tab"`,
		`"activated_tab_id":"tab-2"`,
		`"executed":true`,
		`"memory_persisted":false`,
		`"rollback":"manual_reactivate_previous_tab_if_needed"`,
	} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected body to contain %q, got %s", fragment, recorder.Body.String())
		}
	}
	if client.activatedTargetID != "tab-2" {
		t.Fatalf("expected activated target id tab-2, got %q", client.activatedTargetID)
	}

	events, err := memory.Load(paths)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no memory events, got %+v", events)
	}
}

func TestTabAgentExecuteRejectsPreviewOnlyStep(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}
	if err := memory.AppendPageCapture(paths, memory.PageCaptureInput{
		OccurredAt: "2026-04-08T14:33:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
	}); err != nil {
		t.Fatalf("append page capture failed: %v", err)
	}

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas context",
			CapturedAt: "2026-04-08T14:33:10Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
		},
	})

	body := bytes.NewBufferString(`{"id":"tab-1","step_id":"recommend-memory-relevant-page-capture","confirm":true}`)
	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/agent-execute", body)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "preview-only") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
