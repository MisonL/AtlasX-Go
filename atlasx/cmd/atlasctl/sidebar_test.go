package main

import (
	"strings"
	"testing"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
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
