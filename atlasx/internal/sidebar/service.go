package sidebar

import (
	"errors"
	"fmt"

	"atlasx/internal/settings"
)

var ErrNotConfigured = errors.New("sidebar qa provider is not configured")
var ErrBackendNotImplemented = errors.New("sidebar qa backend is not implemented")

type Config struct {
	DefaultProvider string
	Providers       []settings.SidebarProviderConfig
}

type ProviderStatus struct {
	ID        string `json:"id"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	BaseURL   string `json:"base_url"`
	APIKeyEnv string `json:"api_key_env"`
}

type Status struct {
	Configured      bool             `json:"configured"`
	Ready           bool             `json:"ready"`
	DefaultProvider string           `json:"default_provider"`
	Provider        string           `json:"provider"`
	Model           string           `json:"model"`
	BaseURL         string           `json:"base_url"`
	APIKeyEnv       string           `json:"api_key_env"`
	Providers       []ProviderStatus `json:"providers"`
	Reason          string           `json:"reason"`
}

type AskRequest struct {
	TabID    string `json:"tab_id"`
	Question string `json:"question"`
}

func FromSettings(cfg settings.Config) Config {
	if len(cfg.SidebarProviders) > 0 {
		return Config{
			DefaultProvider: cfg.SidebarDefaultProvider,
			Providers:       cloneProviderConfigs(cfg.SidebarProviders),
		}
	}

	legacyProvider := legacyProviderConfig(cfg)
	if legacyProvider.ID == "" {
		return Config{}
	}
	return Config{
		DefaultProvider: legacyProvider.ID,
		Providers:       []settings.SidebarProviderConfig{legacyProvider},
	}
}

func (c Config) Status() Status {
	status := Status{
		DefaultProvider: c.resolvedDefaultProviderID(),
		Providers:       providerStatuses(c.Providers),
	}
	if len(c.Providers) == 0 {
		status.Reason = ErrNotConfigured.Error()
		return status
	}

	selected, ok := c.providerByID(status.DefaultProvider)
	if !ok {
		status.Configured = true
		status.Reason = "sidebar qa default provider is missing from registry"
		return status
	}

	status.Configured = true
	status.Provider = selected.Provider
	status.Model = selected.Model
	status.BaseURL = selected.BaseURL
	status.APIKeyEnv = selected.APIKeyEnv
	if selected.Provider == "" || selected.Model == "" || selected.BaseURL == "" {
		status.Reason = "sidebar qa config is incomplete"
		return status
	}
	status.Reason = ErrBackendNotImplemented.Error()
	return status
}

func (c Config) Ask(request AskRequest) error {
	if request.Question == "" {
		return errors.New("question is required")
	}
	if request.TabID == "" {
		return errors.New("tab_id is required")
	}

	status := c.Status()
	if !status.Configured {
		return ErrNotConfigured
	}
	if status.Ready {
		return nil
	}
	if status.Reason != "" {
		return fmt.Errorf("%w: %s", ErrBackendNotImplemented, status.Reason)
	}
	return ErrBackendNotImplemented
}

func legacyProviderConfig(cfg settings.Config) settings.SidebarProviderConfig {
	if cfg.SidebarProvider == "" {
		return settings.SidebarProviderConfig{}
	}
	return settings.SidebarProviderConfig{
		ID:       cfg.SidebarProvider,
		Provider: cfg.SidebarProvider,
		Model:    cfg.SidebarModel,
		BaseURL:  cfg.SidebarBaseURL,
	}
}

func cloneProviderConfigs(providers []settings.SidebarProviderConfig) []settings.SidebarProviderConfig {
	cloned := make([]settings.SidebarProviderConfig, len(providers))
	copy(cloned, providers)
	return cloned
}

func providerStatuses(providers []settings.SidebarProviderConfig) []ProviderStatus {
	statuses := make([]ProviderStatus, 0, len(providers))
	for _, provider := range providers {
		statuses = append(statuses, ProviderStatus{
			ID:        provider.ID,
			Provider:  provider.Provider,
			Model:     provider.Model,
			BaseURL:   provider.BaseURL,
			APIKeyEnv: provider.APIKeyEnv,
		})
	}
	return statuses
}

func (c Config) resolvedDefaultProviderID() string {
	if c.DefaultProvider != "" {
		return c.DefaultProvider
	}
	if len(c.Providers) == 0 {
		return ""
	}
	return c.Providers[0].ID
}

func (c Config) providerByID(id string) (settings.SidebarProviderConfig, bool) {
	for _, provider := range c.Providers {
		if provider.ID == id {
			return provider, true
		}
	}
	return settings.SidebarProviderConfig{}, false
}
