package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestLoadViewBootstrapsDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	view, err := LoadView(paths)
	if err != nil {
		t.Fatalf("load view failed: %v", err)
	}

	if view.ConfigFile != paths.ConfigFile {
		t.Fatalf("unexpected config file: %+v", view)
	}
	if view.DefaultProfile != DefaultProfile {
		t.Fatalf("unexpected default profile: %+v", view)
	}
	if view.ListenAddr != DefaultListenAddr {
		t.Fatalf("unexpected listen addr: %+v", view)
	}
	if view.WebAppURL != DefaultWebAppURL {
		t.Fatalf("unexpected web app url: %+v", view)
	}
	if _, err := os.Stat(paths.ConfigFile); err != nil {
		t.Fatalf("expected bootstrapped config file: %v", err)
	}
}

func TestViewRenderIncludesSidebarProviders(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := NewStore(paths.ConfigFile).Save(Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []SidebarProviderConfig{
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

	view, err := LoadView(paths)
	if err != nil {
		t.Fatalf("load view failed: %v", err)
	}

	rendered := view.Render()
	expectedConfigFile := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "AtlasX", "config.json")
	for _, fragment := range []string{
		"config_file=" + expectedConfigFile,
		"memory_persist_enabled=true",
		"sidebar_default_provider=primary",
		"sidebar_provider_count=1",
		"sidebar_provider[0].id=primary",
		"sidebar_provider[0].api_key_env=OPENAI_API_KEY",
	} {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("expected rendered settings to contain %q, got %s", fragment, rendered)
		}
	}
}
