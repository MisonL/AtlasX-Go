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
	ChromeBinary   string `json:"chrome_binary"`
	DefaultProfile string `json:"default_profile"`
	ListenAddr     string `json:"listen_addr"`
	WebAppURL      string `json:"web_app_url"`
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
		DefaultProfile: DefaultProfile,
		ListenAddr:     DefaultListenAddr,
		WebAppURL:      DefaultWebAppURL,
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
	return c
}
