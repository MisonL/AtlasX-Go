package sidebar

import (
	"errors"
	"testing"

	"atlasx/internal/settings"
)

func TestStatusWithoutProvider(t *testing.T) {
	status := FromSettings(settings.Config{}).Status()
	if status.Configured {
		t.Fatal("expected unconfigured sidebar status")
	}
	if status.Reason == "" {
		t.Fatal("expected explicit reason")
	}
}

func TestStatusWithIncompleteConfig(t *testing.T) {
	status := FromSettings(settings.Config{
		SidebarProvider: "openai",
	}).Status()
	if !status.Configured {
		t.Fatal("expected configured status")
	}
	if status.Reason != "sidebar qa config is incomplete" {
		t.Fatalf("unexpected reason: %s", status.Reason)
	}
}

func TestStatusWithProviderRegistry(t *testing.T) {
	status := FromSettings(settings.Config{
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
				ID:        "backup",
				Provider:  "openrouter",
				Model:     "openai/gpt-5",
				BaseURL:   "https://openrouter.ai/api/v1",
				APIKeyEnv: "OPENROUTER_API_KEY",
			},
		},
	}).Status()

	if status.DefaultProvider != "primary" {
		t.Fatalf("unexpected default provider: %s", status.DefaultProvider)
	}
	if len(status.Providers) != 2 {
		t.Fatalf("unexpected providers: %+v", status.Providers)
	}
	if status.APIKeyEnv != "OPENAI_API_KEY" {
		t.Fatalf("unexpected api key env: %s", status.APIKeyEnv)
	}
}

func TestStatusPrefersRegistryOverLegacyFields(t *testing.T) {
	status := FromSettings(settings.Config{
		SidebarProvider:        "legacy-openai",
		SidebarModel:           "legacy-model",
		SidebarBaseURL:         "https://legacy.example.com/v1",
		SidebarDefaultProvider: "registry-openrouter",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "registry-openrouter",
				Provider:  "openrouter",
				Model:     "openai/gpt-5",
				BaseURL:   "https://openrouter.ai/api/v1",
				APIKeyEnv: "OPENROUTER_API_KEY",
			},
		},
	}).Status()

	if status.DefaultProvider != "registry-openrouter" {
		t.Fatalf("unexpected default provider: %s", status.DefaultProvider)
	}
	if status.Provider != "openrouter" {
		t.Fatalf("unexpected provider: %s", status.Provider)
	}
	if len(status.Providers) != 1 || status.Providers[0].ID != "registry-openrouter" {
		t.Fatalf("unexpected providers: %+v", status.Providers)
	}
}

func TestAskRejectsUnconfiguredBackend(t *testing.T) {
	err := Config{}.Ask(AskRequest{
		TabID:    "tab-1",
		Question: "summarize this page",
	})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("unexpected error: %v", err)
	}
}
