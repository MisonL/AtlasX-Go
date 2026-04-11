//go:build darwin

package daemon

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/tabs"
)

func TestSidebarSelectionAskEndpointReturnsStructuredAnswer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")
	var handlerErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			handlerErr = fmt.Errorf("read body failed: %v", err)
			return
		}
		if !bytes.Contains(body, []byte("Atlas isolates profile state.")) {
			handlerErr = fmt.Errorf("selection text missing from request: %s", string(body))
			return
		}
		if !bytes.Contains(body, []byte("这句话是什么意思？")) {
			handlerErr = fmt.Errorf("selection question missing from request: %s", string(body))
			return
		}
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Selected answer"}}]}`))
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

	restoreDaemonHooks(t, &stubTabsClient{
		selection: tabs.SelectionContext{
			ID:                     "tab-1",
			Title:                  "Atlas",
			URL:                    "https://chatgpt.com/atlas",
			SelectionText:          "Atlas selected text",
			CapturedAt:             "2026-04-06T12:00:00Z",
			SelectionPresent:       true,
			SelectionTextTruncated: false,
			SelectionTextLength:    19,
			SelectionTextLimit:     1024,
		},
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-06T12:00:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
	})

	requestBody := `{"tab_id":"tab-1","selection_text":"Atlas isolates profile state.","question":"这句话是什么意思？"}`
	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/selection/ask", bytes.NewBufferString(requestBody))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if handlerErr != nil {
		t.Fatalf("handler error: %v", handlerErr)
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"answer":"Selected answer"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"trace_id":"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}

	events, err := memory.Load(paths)
	if err != nil {
		t.Fatalf("load memory failed: %v", err)
	}
	if len(events) != 1 || events[0].Kind != memory.EventKindQATurn {
		t.Fatalf("unexpected memory events: %+v", events)
	}
	if !strings.Contains(events[0].Question, "Selected Text:\nAtlas isolates profile state.") {
		t.Fatalf("selection question missing from memory event: %+v", events[0])
	}
	if events[0].Answer != "Selected answer" {
		t.Fatalf("unexpected memory event: %+v", events[0])
	}
}

func TestSidebarSelectionAskEndpointAutoCapturesSelection(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")
	var handlerErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			handlerErr = fmt.Errorf("read body failed: %v", err)
			return
		}
		if !bytes.Contains(body, []byte("Atlas auto selection")) {
			handlerErr = fmt.Errorf("auto selection missing from request: %s", string(body))
			return
		}
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Auto selected answer"}}]}`))
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

	restoreDaemonHooks(t, &stubTabsClient{
		selection: tabs.SelectionContext{
			ID:                     "tab-1",
			Title:                  "Atlas",
			URL:                    "https://chatgpt.com/atlas",
			SelectionText:          "Atlas auto selection",
			CapturedAt:             "2026-04-06T12:00:00Z",
			SelectionPresent:       true,
			SelectionTextTruncated: false,
			SelectionTextLength:    20,
			SelectionTextLimit:     1024,
		},
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-06T12:00:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
	})

	requestBody := `{"tab_id":"tab-1","question":"这句话是什么意思？"}`
	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/selection/ask", bytes.NewBufferString(requestBody))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if handlerErr != nil {
		t.Fatalf("handler error: %v", handlerErr)
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"answer":"Auto selected answer"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestSidebarSelectionAskEndpointRejectsEmptySelection(t *testing.T) {
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

	restoreDaemonHooks(t, &stubTabsClient{
		selection: tabs.SelectionContext{
			ID:                 "tab-1",
			Title:              "Atlas",
			URL:                "https://chatgpt.com/atlas",
			CapturedAt:         "2026-04-06T12:00:00Z",
			SelectionTextLimit: 1024,
		},
	})

	requestBody := `{"tab_id":"tab-1","selection_text":"   ","question":"这句话是什么意思？"}`
	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/selection/ask", bytes.NewBufferString(requestBody))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSidebarSelectionAskEndpointSkipsMemoryWhenPersistenceDisabled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("OPENAI_API_KEY", "test-key")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Selected answer"}}]}`))
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

	restoreDaemonHooks(t, &stubTabsClient{
		selection: tabs.SelectionContext{
			ID:                     "tab-1",
			Title:                  "Atlas",
			URL:                    "https://chatgpt.com/atlas",
			SelectionText:          "Atlas selected text",
			CapturedAt:             "2026-04-06T12:00:00Z",
			SelectionPresent:       true,
			SelectionTextTruncated: false,
			SelectionTextLength:    19,
			SelectionTextLimit:     1024,
		},
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-06T12:00:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/selection/ask", bytes.NewBufferString(`{"tab_id":"tab-1","selection_text":"Atlas selected text","question":"这句话是什么意思？"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	events, err := memory.Load(paths)
	if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
		t.Fatalf("load memory failed: %v", err)
	}
	if err == nil && len(events) != 0 {
		t.Fatalf("expected no memory events, got %+v", events)
	}
}
