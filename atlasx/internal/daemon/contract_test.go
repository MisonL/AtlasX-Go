package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/managedruntime"
	"atlasx/internal/mirror"
	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/tabs"
)

func TestStatusEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/status", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"ready",
		"chrome_status",
		"memory_root",
		"memory_events_file",
		"memory_present",
		"memory_event_count",
		"memory_last_event_at",
		"memory_last_event_kind",
		"runtime_manifest_present",
		"runtime_bundle_present",
		"sidebar_qa_configured",
		"sidebar_qa_default_provider",
		"sidebar_qa_providers",
		"sidebar_qa_timeout_ms",
		"sidebar_qa_last_error",
		"sidebar_qa_last_trace_id",
	)
}

func TestSettingsEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/settings", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"config_file",
		"default_profile",
		"listen_addr",
		"web_app_url",
		"sidebar_default_provider",
		"sidebar_providers",
	)
}

func TestProfileEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/profile", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"profiles_root",
		"default_profile",
		"selected_mode",
		"selected_user_data_dir",
		"isolated_user_data_dir",
		"isolated_present",
		"shared_managed",
	)
}

func TestPolicyEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/policy", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"config_file",
		"default_listen_addr",
		"loopback_only_default",
		"remote_control_flag",
		"remote_control_flag_required",
		"shared_profile_managed",
		"sidebar_secrets_persisted",
		"sidebar_default_provider",
		"sidebar_provider_count",
		"sidebar_provider_env_keys",
		"mirror_allowed_roots",
		"chrome_import_allowed_roots",
	)
}

func TestPermissionsEndpointContract(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/v1/permissions", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"source",
		"native_bridge_present",
		"granted_state_observable",
		"accessibility_probe_supported",
		"screen_recording_probe_supported",
		"automation_probe_supported",
		"full_disk_access_probe_supported",
		"permission_prompt_supported",
		"permission_write_supported",
		"permission_state_persisted",
		"os_policy_failures_surface",
		"notes",
	)
}

func TestTabAgentPlanEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
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

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/agent-plan?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"id",
		"title",
		"url",
		"captured_at",
		"goal",
		"read_only",
		"executed",
		"returned",
		"memory_returned",
		"suggestion_returned",
		"recommendation_returned",
		"rollback",
		"guardrails",
		"steps",
	)
}

func TestMemoryEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/memory", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"root",
		"events_file",
		"present",
		"event_count",
		"last_event_at",
		"last_event_kind",
		"returned",
		"events",
	)
}

func TestMemorySearchEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/memory/search?question=atlas", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"question",
		"returned",
		"snippets",
	)
}

func TestRuntimeStatusEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/runtime/status", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"runtime_root",
		"manifest_present",
		"install_plan_present",
		"binary_present",
	)
}

func TestRuntimePlanEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/runtime/plan", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"path",
		"present",
		"source_url",
		"current_phase",
	)
}

func TestSidebarStatusEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	bootstrapConfig(t)

	request := httptest.NewRequest(http.MethodGet, "/v1/sidebar/status", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"configured",
		"ready",
		"default_provider",
		"provider",
		"model",
		"api_key_env",
		"providers",
		"timeout_ms",
		"retry_attempts",
		"token_budget",
		"last_trace_id",
		"last_error",
		"last_error_at",
		"reason",
	)
}

func TestSidebarAskEndpointContract(t *testing.T) {
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

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:    "tab-1",
			Title: "Atlas",
			URL:   "https://chatgpt.com/atlas",
			Text:  "Atlas context",
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/ask", bytes.NewBufferString(`{"tab_id":"tab-1","question":"summarize this page"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "answer", "provider", "model", "context_summary", "trace_id")
}

func TestSidebarSummarizeEndpointContract(t *testing.T) {
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

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:    "tab-1",
			Title: "Atlas",
			URL:   "https://chatgpt.com/atlas",
			Text:  "Atlas context",
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/sidebar/summarize", bytes.NewBufferString(`{"tab_id":"tab-1"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "summary", "provider", "model", "context_summary", "trace_id")
}

func TestTabContextEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		context: tabs.PageContext{
			ID:            "tab-1",
			Title:         "Atlas",
			URL:           "https://chatgpt.com/atlas",
			Text:          "Atlas context",
			CapturedAt:    "2026-04-06T10:00:00Z",
			TextLength:    13,
			TextLimit:     4096,
			TextTruncated: false,
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/context?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"id",
		"title",
		"url",
		"text",
		"captured_at",
		"text_truncated",
		"text_length",
		"text_limit",
		"capture_error",
	)
	if payload["id"] != "tab-1" || payload["url"] != "https://chatgpt.com/atlas" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestHistoryEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := mirror.Save(paths, mirror.Snapshot{
		HistoryRows: []mirror.HistoryEntry{
			{
				URL:           "https://example.com/history",
				Title:         "Example",
				VisitCount:    3,
				LastVisitTime: "2026-04-06T00:00:00Z",
			},
		},
	}); err != nil {
		t.Fatalf("save mirror failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/v1/history", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeArrayResponse(t, recorder)
	if len(payload) != 1 {
		t.Fatalf("unexpected history rows: %+v", payload)
	}
	assertMapKeys(t, payload[0], "url", "title", "visit_count", "last_visit_time")
}

func TestRuntimeVerifyEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if _, err := managedruntime.StageLocal(paths, managedruntime.StageOptions{
		BundlePath: createDaemonFakeChromiumBundle(t),
		Version:    "123.0.0",
		Channel:    "local",
	}); err != nil {
		t.Fatalf("stage local failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/v1/runtime/verify", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"runtime_root",
		"manifest_present",
		"binary_executable",
		"verified",
	)
}

func decodeObjectResponse(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode object response failed: %v", err)
	}
	return payload
}

func decodeArrayResponse(t *testing.T, recorder *httptest.ResponseRecorder) []map[string]any {
	t.Helper()

	var payload []map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode array response failed: %v", err)
	}
	return payload
}

func assertMapKeys(t *testing.T, payload map[string]any, keys ...string) {
	t.Helper()

	for _, key := range keys {
		if _, ok := payload[key]; !ok {
			t.Fatalf("missing key %q in payload %+v", key, payload)
		}
	}
}
