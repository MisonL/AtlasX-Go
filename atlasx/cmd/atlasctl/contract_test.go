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
