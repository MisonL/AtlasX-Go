package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"atlasx/internal/defaultbrowser"
	"atlasx/internal/managedruntime"
	"atlasx/internal/memory"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/tabs"
)

func TestDefaultBrowserStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	previous := readDefaultBrowserStatus
	readDefaultBrowserStatus = func() (defaultbrowser.Status, error) {
		return defaultbrowser.Status{
			Source:        "launchservices",
			HTTPBundleID:  "org.mozilla.firefox",
			HTTPRole:      "all",
			HTTPKnown:     true,
			HTTPSBundleID: "org.mozilla.firefox",
			HTTPSRole:     "all",
			HTTPSKnown:    true,
			Consistent:    true,
		}, nil
	}
	t.Cleanup(func() {
		readDefaultBrowserStatus = previous
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"default-browser", "status"})
	})
	if err != nil {
		t.Fatalf("run default-browser status failed: %v", err)
	}

	assertContainsAll(t, output,
		"source=launchservices",
		"http_bundle_id=org.mozilla.firefox",
		"http_role=all",
		"https_bundle_id=org.mozilla.firefox",
		"https_role=all",
		"consistent=true",
	)
}

func TestDefaultBrowserSetContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	previous := setDefaultBrowserBundleID
	setDefaultBrowserBundleID = func(bundleID string) (defaultbrowser.Status, error) {
		return defaultbrowser.Status{
			Source:        "launchservices",
			HTTPBundleID:  bundleID,
			HTTPRole:      "all",
			HTTPKnown:     true,
			HTTPSBundleID: bundleID,
			HTTPSRole:     "all",
			HTTPSKnown:    true,
			Consistent:    true,
		}, nil
	}
	t.Cleanup(func() {
		setDefaultBrowserBundleID = previous
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"default-browser", "set", "com.openai.atlasx"})
	})
	if err != nil {
		t.Fatalf("run default-browser set failed: %v", err)
	}

	assertContainsAll(t, output,
		"source=",
		"http_bundle_id=",
		"http_role=",
		"https_bundle_id=",
		"https_role=",
		"consistent=",
	)
}

func TestLogsStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"logs", "status"})
	})
	if err != nil {
		t.Fatalf("run logs status failed: %v", err)
	}

	assertContainsAll(t, output,
		"logs_root=",
		"present=",
		"file_count=",
		"total_bytes=",
		"returned=",
	)
}

func TestUpdatesStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"updates", "status"})
	})
	if err != nil {
		t.Fatalf("run updates status failed: %v", err)
	}

	assertContainsAll(t, output,
		"runtime_root=",
		"manifest_present=",
		"staged_version=",
		"staged_ready=",
		"plan_present=",
		"plan_phase=",
		"plan_pending=",
		"plan_in_flight=",
	)
}

func TestMemoryControlsContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"memory", "controls"})
	})
	if err != nil {
		t.Fatalf("run memory controls failed: %v", err)
	}

	assertContainsAll(t, output,
		"config_file=",
		"persist_enabled=",
		"page_visibility_enabled=",
		"hidden_host_count=",
	)
}

func TestDoctorJSONContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"doctor", "--json"})
	})
	if err != nil {
		t.Fatalf("run doctor --json failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("decode doctor json failed: %v output=%s", err, output)
	}
	for _, key := range []string{"Paths", "Config", "Chrome", "ChromeStatus", "RuntimeManifest", "Session"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("expected key %q in payload: %+v", key, payload)
		}
	}
}

func TestProfileStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"profile", "status"})
	})
	if err != nil {
		t.Fatalf("run profile status failed: %v", err)
	}

	assertContainsAll(t, output,
		"profiles_root=",
		"default_profile=",
		"selected_mode=",
		"selected_user_data_dir=",
		"isolated_user_data_dir=",
		"isolated_present=",
		"shared_managed=",
	)
}

func TestPolicyStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"policy", "status"})
	})
	if err != nil {
		t.Fatalf("run policy status failed: %v", err)
	}

	assertContainsAll(t, output,
		"config_file=",
		"default_listen_addr=",
		"loopback_only_default=",
		"remote_control_flag=",
		"remote_control_flag_required=",
		"shared_profile_managed=",
		"sidebar_secrets_persisted=",
		"sidebar_provider_count=",
		"mirror_allowed_root_count=",
		"chrome_import_allowed_root_count=",
	)
}

func TestPermissionsStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"permissions", "status"})
	})
	if err != nil {
		t.Fatalf("run permissions status failed: %v", err)
	}

	assertContainsAll(t, output,
		"source=",
		"granted_state_observable=",
		"accessibility_probe_supported=",
		"screen_recording_probe_supported=",
		"automation_probe_supported=",
		"full_disk_access_probe_supported=",
		"permission_prompt_supported=",
		"permission_write_supported=",
		"os_policy_failures_surface=",
	)
}

func TestTabsAgentPlanContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas task page",
			CapturedAt: "2026-04-08T13:00:00Z",
		},
		targets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "agent-plan", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs agent-plan failed: %v", err)
	}

	assertContainsAll(t, output,
		"goal=",
		"returned=",
		"read_only=",
		"executed=",
		"suggestion_returned=",
		"recommendation_returned=",
		"rollback=",
		"step_id=",
		"executable=",
		"execution_path=",
		"requires_provider=",
	)
}

func TestTabsAuthModeContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		authMode: tabs.AuthModeView{
			ID:                     "tab-1",
			Title:                  "ChatGPT",
			URL:                    "https://chatgpt.com/c/abc123",
			CapturedAt:             "2026-04-09T10:00:00Z",
			Host:                   "chatgpt.com",
			Path:                   "/c/abc123",
			Mode:                   "logged_in",
			Inferred:               true,
			Reason:                 "workspace_signals_observed",
			LoginPromptPresent:     false,
			WorkspaceSignalPresent: true,
			CookieCount:            1,
			CookieNames:            []string{"oai-session"},
			LocalStorageCount:      1,
			LocalStorageKeys:       []string{"atlas:last-project"},
			SessionStorageCount:    0,
			SessionStorageKeys:     []string{},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "auth-mode", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs auth-mode failed: %v", err)
	}

	assertContainsAll(t, output,
		"id=tab-1",
		"mode=logged_in",
		"inferred=true",
		"reason=workspace_signals_observed",
		"host=chatgpt.com",
		"path=/c/abc123",
		"cookie_count=1",
		"local_storage_count=1",
		"session_storage_count=0",
	)
}

func TestTabsGroupsContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "groups"})
	})
	if err != nil {
		t.Fatalf("run tabs groups failed: %v", err)
	}

	assertContainsAll(t, output,
		"inferred=",
		"returned=",
		"group_id=",
		"label=",
		"window_returned=",
		"window_id=",
	)
}

func TestTabsAgentExecuteContract(t *testing.T) {
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
			Text:       "Atlas task page",
			CapturedAt: "2026-04-08T13:00:00Z",
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

	assertContainsAll(t, output,
		"tab_id=tab-1",
		"step_id=suggest-summarize_page",
		"step_kind=sidebar_summarize",
		"executed=true",
		"confirmed=true",
		"trace_id=",
		"provider=openai",
		"model=gpt-5.4",
		"memory_persisted=false",
		"rollback=not_required_no_memory_persisted",
		`result="Atlas summary"`,
		`context_summary=`,
	)
}

func TestTabsAgentExecuteRelatedTabContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	client := &stubCommandTabsClient{
		context: tabs.PageContext{
			ID:         "tab-1",
			Title:      "Atlas",
			URL:        "https://chatgpt.com/atlas",
			Text:       "Atlas task page",
			CapturedAt: "2026-04-08T13:05:00Z",
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

	assertContainsAll(t, output,
		"tab_id=tab-1",
		"step_id=recommend-related-tab-tab-2",
		"step_kind=related_tab",
		"activated_tab_id=tab-2",
		"executed=true",
		"confirmed=true",
		"trace_id=",
		"memory_persisted=false",
		"rollback=manual_reactivate_previous_tab_if_needed",
	)
	if client.activatedTargetID != "tab-2" {
		t.Fatalf("unexpected activated target id: %q", client.activatedTargetID)
	}
}

func TestTabsAgentExecuteMemorySnippetContract(t *testing.T) {
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
		OccurredAt: "2026-04-08T15:40:00Z",
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
			Text:       "Atlas task page",
			CapturedAt: "2026-04-08T15:41:00Z",
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

	assertContainsAll(t, output,
		"tab_id=tab-1",
		"step_id=recommend-memory-relevant-page-capture",
		"step_kind=memory_snippet",
		"executed=true",
		"confirmed=true",
		"trace_id=",
		"provider=openai",
		"model=gpt-5.4",
		"memory_persisted=false",
		"rollback=not_required_no_memory_persisted",
		`result="Memory relevance answer"`,
	)
}

func TestTabsAgentExecuteBatchContract(t *testing.T) {
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
		OccurredAt: "2026-04-08T16:50:00Z",
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
			Text:       "Atlas task page",
			CapturedAt: "2026-04-08T16:51:00Z",
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
			"tab-1", "recommend-related-tab-tab-2", "recommend-memory-relevant-page-capture",
		})
	})
	if err != nil {
		t.Fatalf("run tabs agent-execute batch failed: %v", err)
	}

	assertContainsAll(t, output,
		"tab_id=tab-1",
		"requested=2",
		"executed=2",
		"stopped=false",
		"max_steps=2",
		"batch_index=0",
		"step_kind=related_tab",
		"batch_index=1",
		"step_kind=memory_snippet",
	)
}

func TestTabsOpenDevToolsPanelInWindowContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsPanelInWindow: tabs.WindowOpenResult{
			WindowID:          7,
			ActivatedTargetID: "tab-7",
			Target: tabs.Target{
				ID:    "devtools-window-tab",
				Type:  "page",
				Title: "DevTools",
				URL:   "http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Ftab-1",
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel-in-window", "tab-1", "network", "7"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools-panel-in-window failed: %v", err)
	}

	assertContainsAll(t, output,
		"window_id=7",
		"activated_target_id=tab-7",
		"id=devtools-window-tab",
		"title=\"DevTools\"",
		"panel=network",
	)
}

func TestTabsOpenDevToolsInWindowContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsInWindow: tabs.WindowOpenResult{
			WindowID:          7,
			ActivatedTargetID: "tab-7",
			Target: tabs.Target{
				ID:    "devtools-window-tab",
				Type:  "page",
				Title: "DevTools",
				URL:   "http://127.0.0.1/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Ftab-1",
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-in-window", "tab-1", "7"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools-in-window failed: %v", err)
	}

	assertContainsAll(t, output,
		"window_id=7",
		"activated_target_id=tab-7",
		"id=devtools-window-tab",
		"title=\"DevTools\"",
	)
}

func TestTabsOpenDevToolsWindowIntoWindowContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsWindowIntoWindow: tabs.DevToolsWindowOpenResult{
			SourceWindowID: 11,
			TargetWindowID: 21,
			Returned:       1,
			OpenedTargets: []tabs.DevToolsWindowOpenTarget{
				{
					SourceTargetID:    "src-1",
					ActivatedTargetID: "dst-1",
					Target: tabs.Target{
						ID:    "devtools-open-1",
						Type:  "page",
						Title: "DevTools",
						URL:   "http://127.0.0.1/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1",
					},
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-window-into-window", "11", "21"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools-window-into-window failed: %v", err)
	}

	assertContainsAll(t, output,
		"source_window_id=11",
		"target_window_id=21",
		"returned=1",
		"source_target_id=src-1",
		"activated_target_id=dst-1",
		"id=devtools-open-1",
	)
}

func TestTabsOpenDevToolsWindowToWindowsContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsWindowToWindows: tabs.DevToolsWindowToWindowsResult{
			SourceWindowID: 11,
			Returned:       1,
			OpenedTargets: []tabs.DevToolsWindowToWindowsTarget{
				{
					SourceTargetID: "src-1",
					Target: tabs.Target{
						ID:    "devtools-open-1",
						Type:  "page",
						Title: "DevTools",
						URL:   "http://127.0.0.1/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1",
					},
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-window-to-windows", "11"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools-window-to-windows failed: %v", err)
	}

	assertContainsAll(t, output,
		"source_window_id=11",
		"returned=1",
		"source_target_id=src-1",
		"id=devtools-open-1",
	)
}

func TestTabsOpenDevToolsPanelWindowToWindowsContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsPanelWindowToWindows: tabs.DevToolsPanelWindowToWindowsResult{
			SourceWindowID: 11,
			Panel:          "network",
			Returned:       1,
			OpenedTargets: []tabs.DevToolsWindowToWindowsTarget{
				{
					SourceTargetID: "src-1",
					Target: tabs.Target{
						ID:    "devtools-open-1",
						Type:  "page",
						Title: "DevTools",
						URL:   "http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1",
					},
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel-window-to-windows", "11", "network"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools-panel-window-to-windows failed: %v", err)
	}

	assertContainsAll(t, output,
		"source_window_id=11",
		"panel=network",
		"returned=1",
		"source_target_id=src-1",
		"id=devtools-open-1",
	)
}

func TestTabsSetTitleContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		titleUpdate: tabs.TitleUpdateResult{
			ID:    "tab-1",
			Title: "Atlas Workbench",
			URL:   "https://openai.com/work",
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "set-title", "tab-1", "Atlas Workbench"})
	})
	if err != nil {
		t.Fatalf("run tabs set-title failed: %v", err)
	}

	assertContainsAll(t, output,
		"id=tab-1",
		`title="Atlas Workbench"`,
		"url=https://openai.com/work",
	)
}

func TestTabsOpenDevToolsPanelWindowIntoWindowContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		openDevToolsPanelWindowIntoWindow: tabs.DevToolsPanelWindowOpenResult{
			SourceWindowID: 11,
			Panel:          "network",
			TargetWindowID: 21,
			Returned:       1,
			OpenedTargets: []tabs.DevToolsWindowOpenTarget{
				{
					SourceTargetID:    "src-1",
					ActivatedTargetID: "dst-1",
					Target: tabs.Target{
						ID:    "devtools-open-1",
						Type:  "page",
						Title: "DevTools",
						URL:   "http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1",
					},
				},
			},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-devtools-panel-window-into-window", "11", "network", "21"})
	})
	if err != nil {
		t.Fatalf("run tabs open-devtools-panel-window-into-window failed: %v", err)
	}

	assertContainsAll(t, output,
		"source_window_id=11",
		"panel=network",
		"target_window_id=21",
		"returned=1",
		"source_target_id=src-1",
		"activated_target_id=dst-1",
		"id=devtools-open-1",
	)
}

func TestRuntimeStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"runtime", "status"})
	})
	if err != nil {
		t.Fatalf("run runtime status failed: %v", err)
	}

	assertContainsAll(t, output,
		"runtime_root=",
		"manifest_present=",
		"install_plan_present=",
		"binary_executable=",
	)
}

func TestRuntimePlanStatusContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	plan, err := managedruntime.NewInstallPlan(managedruntime.InstallPlanOptions{
		Version:          "123.0.0",
		Channel:          "stable",
		SourceURL:        "https://example.com/chromium.zip",
		ExpectedSHA256:   "deadbeef",
		ArchivePath:      "/tmp/chromium.zip",
		StagedBundlePath: "/tmp/Chromium.app",
	})
	if err != nil {
		t.Fatalf("new install plan failed: %v", err)
	}
	plan.CurrentPhase = managedruntime.InstallPhaseVerifying
	if err := managedruntime.SaveInstallPlan(paths, plan); err != nil {
		t.Fatalf("save install plan failed: %v", err)
	}

	output, err := captureStdout(t, func() error {
		return run([]string{"runtime", "plan", "status"})
	})
	if err != nil {
		t.Fatalf("run runtime plan status failed: %v", err)
	}

	assertContainsAll(t, output,
		"install_plan_present=true",
		"install_plan_source_url=https://example.com/chromium.zip",
		"install_plan_phase=verifying",
	)
}

func TestRuntimeVerifyContractWithoutManifest(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"runtime", "verify"})
	})
	if err == nil {
		t.Fatal("expected runtime verify to fail without manifest")
	}
	if !strings.Contains(err.Error(), "managed runtime manifest is not present") {
		t.Fatalf("unexpected runtime verify error: %v", err)
	}

	assertContainsAll(t, output,
		"manifest_present=false",
		"verified=false",
	)
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	previousStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe failed: %v", err)
	}

	os.Stdout = writer
	runErr := fn()
	_ = writer.Close()
	os.Stdout = previousStdout

	data, readErr := io.ReadAll(reader)
	_ = reader.Close()
	if readErr != nil {
		t.Fatalf("read captured stdout failed: %v", readErr)
	}
	return string(data), runErr
}

func assertContainsAll(t *testing.T, output string, fragments ...string) {
	t.Helper()

	for _, fragment := range fragments {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, output=%s", fragment, output)
		}
	}
}
