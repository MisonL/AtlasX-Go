package settings

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBootstrapCreatesDefaultConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "AtlasX", "config.json")
	cfg, err := NewStore(path).Bootstrap()
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}

	if cfg.WebAppURL != DefaultWebAppURL {
		t.Fatalf("unexpected web app url: %s", cfg.WebAppURL)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}

func TestSaveLoadSidebarProviderRegistry(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	path := filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "AtlasX", "config.json")
	store := NewStore(path)
	input := Config{
		SidebarProviders: []SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   "https://api.openai.com/v1",
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	}
	if err := store.Save(input); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.SidebarDefaultProvider != "primary" {
		t.Fatalf("unexpected default provider: %s", loaded.SidebarDefaultProvider)
	}
	if len(loaded.SidebarProviders) != 1 {
		t.Fatalf("unexpected providers: %+v", loaded.SidebarProviders)
	}
	if loaded.SidebarProviders[0].APIKeyEnv != "OPENAI_API_KEY" {
		t.Fatalf("unexpected api key env: %+v", loaded.SidebarProviders[0])
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config failed: %v", err)
	}
	text := string(data)
	if !strings.Contains(text, "\"api_key_env\": \"OPENAI_API_KEY\"") {
		t.Fatalf("missing api_key_env in config: %s", text)
	}
	if strings.Contains(text, "sk-live-") {
		t.Fatalf("unexpected raw secret in config: %s", text)
	}
}
