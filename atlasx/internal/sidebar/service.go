package sidebar

import (
	"errors"
	"fmt"

	"atlasx/internal/settings"
)

var ErrNotConfigured = errors.New("sidebar qa provider is not configured")
var ErrBackendNotImplemented = errors.New("sidebar qa backend is not implemented")

type Config struct {
	Provider string
	Model    string
	BaseURL  string
}

type Status struct {
	Configured bool   `json:"configured"`
	Ready      bool   `json:"ready"`
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	BaseURL    string `json:"base_url"`
	Reason     string `json:"reason"`
}

type AskRequest struct {
	TabID    string `json:"tab_id"`
	Question string `json:"question"`
}

func FromSettings(cfg settings.Config) Config {
	return Config{
		Provider: cfg.SidebarProvider,
		Model:    cfg.SidebarModel,
		BaseURL:  cfg.SidebarBaseURL,
	}
}

func (c Config) Status() Status {
	status := Status{
		Provider: c.Provider,
		Model:    c.Model,
		BaseURL:  c.BaseURL,
	}
	if c.Provider == "" {
		status.Reason = ErrNotConfigured.Error()
		return status
	}
	if c.Model == "" || c.BaseURL == "" {
		status.Configured = true
		status.Reason = "sidebar qa config is incomplete"
		return status
	}
	status.Configured = true
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
