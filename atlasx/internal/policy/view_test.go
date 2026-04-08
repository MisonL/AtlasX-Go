package policy

import (
	"strings"
	"testing"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
)

func TestLoadViewReturnsGuardrails(t *testing.T) {
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
			{
				ID:        "secondary",
				Provider:  "openrouter",
				Model:     "gpt-5.4-mini",
				BaseURL:   "https://openrouter.ai/api/v1",
				APIKeyEnv: "OPENROUTER_API_KEY",
			},
		},
	}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	view, err := LoadView(paths)
	if err != nil {
		t.Fatalf("load view failed: %v", err)
	}

	if !view.LoopbackOnlyDefault || !view.RemoteControlFlagRequired {
		t.Fatalf("unexpected remote-control policy: %+v", view)
	}
	if view.SharedProfileManaged || view.SidebarSecretsPersisted {
		t.Fatalf("unexpected policy booleans: %+v", view)
	}
	if view.SidebarProviderCount != 2 || len(view.SidebarProviderEnvKeys) != 2 {
		t.Fatalf("unexpected sidebar policy fields: %+v", view)
	}
	if view.RemoteControlFlag != RemoteControlFlag {
		t.Fatalf("unexpected remote control flag: %+v", view)
	}
}

func TestViewRenderIncludesGuardrailFields(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	view, err := LoadView(paths)
	if err != nil {
		t.Fatalf("load view failed: %v", err)
	}

	rendered := view.Render()
	for _, fragment := range []string{
		"default_listen_addr=127.0.0.1:17537",
		"loopback_only_default=true",
		"remote_control_flag=--allow-remote-control",
		"shared_profile_managed=false",
		"sidebar_secrets_persisted=false",
	} {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("expected rendered policy to contain %q, got %s", fragment, rendered)
		}
	}
}
