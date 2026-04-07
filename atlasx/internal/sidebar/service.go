package sidebar

import (
	"errors"
	"fmt"
	"os"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/tabs"
)

var ErrNotConfigured = errors.New("sidebar qa provider is not configured")
var ErrBackendNotImplemented = errors.New("sidebar qa backend is not implemented")
var ErrProviderFailed = errors.New("sidebar qa provider request failed")
var ErrProviderNotFound = errors.New("sidebar qa provider id is not configured")
var ErrTokenBudgetExceeded = errors.New("sidebar qa token budget exceeded")

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
	TimeoutMS       int              `json:"timeout_ms"`
	RetryAttempts   int              `json:"retry_attempts"`
	TokenBudget     int              `json:"token_budget"`
	LastTraceID     string           `json:"last_trace_id"`
	LastError       string           `json:"last_error"`
	LastErrorAt     string           `json:"last_error_at"`
	Reason          string           `json:"reason"`
}

type AskRequest struct {
	TabID      string `json:"tab_id"`
	Question   string `json:"question"`
	ProviderID string `json:"provider_id,omitempty"`
}

type AskResponse struct {
	Answer         string `json:"answer"`
	Provider       string `json:"provider"`
	Model          string `json:"model"`
	ContextSummary string `json:"context_summary"`
	TraceID        string `json:"trace_id"`
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
		TimeoutMS:       defaultProviderTimeoutMS(),
		RetryAttempts:   defaultProviderRetryAttempts(),
		TokenBudget:     defaultProviderTokenBudget(),
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
	if selected.APIKeyEnv == "" {
		status.Reason = "sidebar qa api key env is not configured"
		return status
	}
	if os.Getenv(selected.APIKeyEnv) == "" {
		status.Reason = fmt.Sprintf("sidebar qa api key env %s is not set", selected.APIKeyEnv)
		return status
	}
	if !providerSupported(selected.Provider) {
		status.Reason = ErrBackendNotImplemented.Error()
		return status
	}
	status.Ready = true
	return status
}

func (c Config) Ask(request AskRequest, context tabs.PageContext) (AskResponse, error) {
	return c.AskWithMemory(request, context, nil)
}

func (c Config) AskWithMemory(request AskRequest, context tabs.PageContext, memorySnippets []string) (AskResponse, error) {
	if request.Question == "" {
		return AskResponse{}, errors.New("question is required")
	}
	if request.TabID == "" {
		return AskResponse{}, errors.New("tab_id is required")
	}

	if err := c.Validate(request.ProviderID); err != nil {
		return AskResponse{}, err
	}

	selected, err := c.resolveProvider(request.ProviderID)
	if err != nil {
		return AskResponse{}, err
	}

	apiKey := os.Getenv(selected.APIKeyEnv)
	answer, model, err := c.askProvider(selected, apiKey, request.Question, context, memorySnippets)
	if err != nil {
		if errors.Is(err, ErrTokenBudgetExceeded) {
			return AskResponse{}, err
		}
		return AskResponse{}, fmt.Errorf("%w: %s", ErrProviderFailed, err)
	}
	return AskResponse{
		Answer:         answer,
		Provider:       selected.Provider,
		Model:          model,
		ContextSummary: buildContextSummary(context),
	}, nil
}

func (c Config) Validate(providerID string) error {
	selected, err := c.resolveProvider(providerID)
	if err != nil {
		return err
	}
	if selected.Provider == "" || selected.Model == "" || selected.BaseURL == "" {
		return fmt.Errorf("%w: sidebar qa config is incomplete", ErrNotConfigured)
	}
	if selected.APIKeyEnv == "" {
		return fmt.Errorf("%w: sidebar qa api key env is not configured", ErrNotConfigured)
	}
	if os.Getenv(selected.APIKeyEnv) == "" {
		return fmt.Errorf("%w: sidebar qa api key env %s is not set", ErrNotConfigured, selected.APIKeyEnv)
	}
	if !providerSupported(selected.Provider) {
		return ErrBackendNotImplemented
	}
	return nil
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

func (c Config) resolveProvider(providerID string) (settings.SidebarProviderConfig, error) {
	if len(c.Providers) == 0 {
		return settings.SidebarProviderConfig{}, ErrNotConfigured
	}

	selectedID := providerID
	if selectedID == "" {
		selectedID = c.resolvedDefaultProviderID()
	}
	provider, ok := c.providerByID(selectedID)
	if !ok {
		return settings.SidebarProviderConfig{}, fmt.Errorf("%w: %s", ErrProviderNotFound, selectedID)
	}
	return provider, nil
}

func (c Config) askProvider(provider settings.SidebarProviderConfig, apiKey string, question string, context tabs.PageContext, memorySnippets []string) (string, string, error) {
	switch provider.Provider {
	case "openai", "openai-compatible":
		return askOpenAICompatible(provider, apiKey, question, context, memorySnippets)
	case "openrouter":
		return askOpenRouter(provider, apiKey, question, context, memorySnippets)
	default:
		return "", "", ErrBackendNotImplemented
	}
}

func (c Config) StatusWithRuntime(paths macos.Paths) (Status, error) {
	status := c.Status()
	runtimeState, err := LoadRuntimeState(paths)
	if err != nil {
		return Status{}, err
	}
	status.TimeoutMS = runtimeState.TimeoutMS
	status.RetryAttempts = runtimeState.RetryAttempts
	status.TokenBudget = runtimeState.TokenBudget
	status.LastTraceID = runtimeState.LastTraceID
	status.LastError = runtimeState.LastError
	status.LastErrorAt = runtimeState.LastErrorAt
	return status, nil
}
