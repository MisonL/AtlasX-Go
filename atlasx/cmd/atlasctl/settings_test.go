package main

import (
	"strings"
	"testing"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
)

func TestSettingsShowOutputsDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	output, err := captureStdout(t, func() error {
		return run([]string{"settings"})
	})
	if err != nil {
		t.Fatalf("run settings failed: %v", err)
	}

	assertContainsAll(t, output,
		"AtlasX Settings",
		"config_file=",
		"default_profile=isolated",
		"listen_addr=127.0.0.1:17537",
		"web_app_url=https://chatgpt.com/atlas?get-started",
	)
}

func TestSettingsShowOutputsSidebarProviders(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

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
		return run([]string{"settings", "show"})
	})
	if err != nil {
		t.Fatalf("run settings show failed: %v", err)
	}

	for _, fragment := range []string{
		"sidebar_default_provider=primary",
		"sidebar_provider_count=1",
		"sidebar_provider[0].id=primary",
		"sidebar_provider[0].provider=openai",
		"sidebar_provider[0].model=gpt-5.4",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}
