package settings

import (
	"encoding/json"
	"os"

	"atlasx/internal/platform/macos"
)

const (
	DefaultListenAddr = "127.0.0.1:17537"
	DefaultProfile    = "isolated"
	DefaultWebAppURL  = "https://chatgpt.com/atlas?get-started"
)

type Config struct {
	ChromeBinary           string                  `json:"chrome_binary"`
	DefaultProfile         string                  `json:"default_profile"`
	ListenAddr             string                  `json:"listen_addr"`
	WebAppURL              string                  `json:"web_app_url"`
	MemoryPersistEnabled   *bool                   `json:"memory_persist_enabled,omitempty"`
	MemoryPageVisibility   *bool                   `json:"memory_page_visibility_enabled,omitempty"`
	SidebarProvider        string                  `json:"sidebar_provider"`
	SidebarModel           string                  `json:"sidebar_model"`
	SidebarBaseURL         string                  `json:"sidebar_base_url"`
	SidebarDefaultProvider string                  `json:"sidebar_default_provider"`
	SidebarProviders       []SidebarProviderConfig `json:"sidebar_providers"`
}

type SidebarProviderConfig struct {
	ID        string `json:"id"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
	APIKeyEnv string `json:"api_key_env"`
}

type Store struct {
	path string
}

func NewStore(path string) Store {
	return Store{path: path}
}

func (s Store) Bootstrap() (Config, error) {
	if _, err := os.Stat(s.path); err == nil {
		return s.Load()
	}

	cfg := DefaultConfig()
	if err := s.Save(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (s Store) Load() (Config, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg.withDefaults(), nil
}

func (s Store) Save(cfg Config) error {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return err
	}
	if err := macos.EnsureDir(paths.SupportRoot); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg.withDefaults(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, append(data, '\n'), 0o644)
}

func DefaultConfig() Config {
	return Config{
		DefaultProfile:       DefaultProfile,
		ListenAddr:           DefaultListenAddr,
		WebAppURL:            DefaultWebAppURL,
		MemoryPersistEnabled: Bool(true),
		MemoryPageVisibility: Bool(true),
	}
}

func (c Config) withDefaults() Config {
	if c.DefaultProfile == "" {
		c.DefaultProfile = DefaultProfile
	}
	if c.ListenAddr == "" {
		c.ListenAddr = DefaultListenAddr
	}
	if c.WebAppURL == "" {
		c.WebAppURL = DefaultWebAppURL
	}
	if c.MemoryPersistEnabled == nil {
		c.MemoryPersistEnabled = Bool(true)
	}
	if c.MemoryPageVisibility == nil {
		c.MemoryPageVisibility = Bool(true)
	}
	if len(c.SidebarProviders) > 0 && c.SidebarDefaultProvider == "" {
		c.SidebarDefaultProvider = c.SidebarProviders[0].ID
	}
	return c
}

func (c Config) MemoryPersistEnabledValue() bool {
	if c.MemoryPersistEnabled == nil {
		return true
	}
	return *c.MemoryPersistEnabled
}

func (c Config) MemoryPageVisibilityEnabledValue() bool {
	if c.MemoryPageVisibility == nil {
		return true
	}
	return *c.MemoryPageVisibility
}

func Bool(value bool) *bool {
	return &value
}
