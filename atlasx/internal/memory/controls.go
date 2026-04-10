package memory

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
)

type Controls struct {
	ConfigFile            string   `json:"config_file"`
	PersistEnabled        bool     `json:"persist_enabled"`
	PageVisibilityEnabled bool     `json:"page_visibility_enabled"`
	HiddenHosts           []string `json:"hidden_hosts"`
}

type ControlsUpdate struct {
	PersistEnabled        *bool  `json:"persist_enabled,omitempty"`
	PageVisibilityEnabled *bool  `json:"page_visibility_enabled,omitempty"`
	SiteHost              string `json:"site_host,omitempty"`
	SiteVisibilityEnabled *bool  `json:"site_visibility_enabled,omitempty"`
}

func LoadControls(paths macos.Paths) (Controls, error) {
	config, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return Controls{}, err
	}
	hiddenHosts, err := normalizeHiddenHosts(config.MemoryHiddenHosts)
	if err != nil {
		return Controls{}, err
	}

	return Controls{
		ConfigFile:            paths.ConfigFile,
		PersistEnabled:        config.MemoryPersistEnabledValue(),
		PageVisibilityEnabled: config.MemoryPageVisibilityEnabledValue(),
		HiddenHosts:           hiddenHosts,
	}, nil
}

func SetPersistEnabled(paths macos.Paths, enabled bool) (Controls, error) {
	return UpdateControls(paths, ControlsUpdate{PersistEnabled: settings.Bool(enabled)})
}

func SetPageVisibilityEnabled(paths macos.Paths, enabled bool) (Controls, error) {
	return UpdateControls(paths, ControlsUpdate{PageVisibilityEnabled: settings.Bool(enabled)})
}

func SetSiteVisibility(paths macos.Paths, rawHost string, enabled bool) (Controls, error) {
	return UpdateControls(paths, ControlsUpdate{
		SiteHost:              rawHost,
		SiteVisibilityEnabled: settings.Bool(enabled),
	})
}

func UpdateControls(paths macos.Paths, update ControlsUpdate) (Controls, error) {
	store := settings.NewStore(paths.ConfigFile)
	config, err := store.Bootstrap()
	if err != nil {
		return Controls{}, err
	}

	if update.PersistEnabled != nil {
		config.MemoryPersistEnabled = settings.Bool(*update.PersistEnabled)
	}
	if update.PageVisibilityEnabled != nil {
		config.MemoryPageVisibility = settings.Bool(*update.PageVisibilityEnabled)
	}
	if update.SiteHost != "" || update.SiteVisibilityEnabled != nil {
		if update.SiteHost == "" || update.SiteVisibilityEnabled == nil {
			return Controls{}, fmt.Errorf("site_host and site_visibility_enabled must be provided together")
		}
		host, err := NormalizeSiteHost(update.SiteHost)
		if err != nil {
			return Controls{}, err
		}
		config.MemoryHiddenHosts, err = setHiddenHost(config.MemoryHiddenHosts, host, !*update.SiteVisibilityEnabled)
		if err != nil {
			return Controls{}, err
		}
	}
	if err := store.Save(config); err != nil {
		return Controls{}, err
	}
	return LoadControls(paths)
}

func AppendPageCaptureControlled(paths macos.Paths, input PageCaptureInput) (bool, error) {
	controls, err := LoadControls(paths)
	if err != nil {
		return false, err
	}
	if !controls.PersistEnabled {
		return false, nil
	}
	if err := AppendPageCapture(paths, input); err != nil {
		return false, err
	}
	return true, nil
}

func AppendQATurnControlled(paths macos.Paths, input QATurnInput) (bool, error) {
	controls, err := LoadControls(paths)
	if err != nil {
		return false, err
	}
	if !controls.PersistEnabled {
		return false, nil
	}
	if err := AppendQATurn(paths, input); err != nil {
		return false, err
	}
	return true, nil
}

func ParsePersistValue(raw string) (bool, error) {
	switch raw {
	case "enabled", "true", "on", "1":
		return true, nil
	case "disabled", "false", "off", "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid persist value %q", raw)
	}
}

func ParsePageVisibilityValue(raw string) (bool, error) {
	switch raw {
	case "visible", "enabled", "true", "on", "1":
		return true, nil
	case "hidden", "disabled", "false", "off", "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid page visibility value %q", raw)
	}
}

func NormalizeSiteHost(raw string) (string, error) {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return "", fmt.Errorf("invalid site host %q", raw)
	}

	candidate := trimmed
	if !strings.Contains(candidate, "://") {
		candidate = "https://" + candidate
	}

	parsed, err := url.Parse(candidate)
	if err != nil {
		return "", fmt.Errorf("invalid site host %q", raw)
	}
	host := strings.TrimSpace(strings.ToLower(strings.TrimSuffix(parsed.Hostname(), ".")))
	if host == "" {
		return "", fmt.Errorf("invalid site host %q", raw)
	}
	return host, nil
}

func normalizeHiddenHosts(rawHosts []string) ([]string, error) {
	normalized := make([]string, 0, len(rawHosts))
	seen := make(map[string]struct{}, len(rawHosts))
	for _, rawHost := range rawHosts {
		host, err := NormalizeSiteHost(rawHost)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[host]; ok {
			continue
		}
		seen[host] = struct{}{}
		normalized = append(normalized, host)
	}
	sort.Strings(normalized)
	if normalized == nil {
		return []string{}, nil
	}
	return normalized, nil
}

func setHiddenHost(rawHosts []string, host string, hidden bool) ([]string, error) {
	normalized, err := normalizeHiddenHosts(rawHosts)
	if err != nil {
		return nil, err
	}

	filtered := make([]string, 0, len(normalized))
	found := false
	for _, existing := range normalized {
		if existing == host {
			found = true
			if hidden {
				filtered = append(filtered, existing)
			}
			continue
		}
		filtered = append(filtered, existing)
	}
	if hidden && !found {
		filtered = append(filtered, host)
		sort.Strings(filtered)
	}
	if filtered == nil {
		return []string{}, nil
	}
	return filtered, nil
}
