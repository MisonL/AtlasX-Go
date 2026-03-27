package settings

import (
	"os"
	"path/filepath"
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
