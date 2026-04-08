package policy

import (
	"fmt"
	"strings"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
	"atlasx/internal/sourcepaths"
)

const (
	RemoteControlFlag = "--allow-remote-control"
)

type View struct {
	ConfigFile                string   `json:"config_file"`
	DefaultListenAddr         string   `json:"default_listen_addr"`
	LoopbackOnlyDefault       bool     `json:"loopback_only_default"`
	RemoteControlFlag         string   `json:"remote_control_flag"`
	RemoteControlFlagRequired bool     `json:"remote_control_flag_required"`
	SharedProfileManaged      bool     `json:"shared_profile_managed"`
	SidebarSecretsPersisted   bool     `json:"sidebar_secrets_persisted"`
	SidebarDefaultProvider    string   `json:"sidebar_default_provider"`
	SidebarProviderCount      int      `json:"sidebar_provider_count"`
	SidebarProviderEnvKeys    []string `json:"sidebar_provider_env_keys"`
	MirrorAllowedRoots        []string `json:"mirror_allowed_roots"`
	ChromeImportAllowedRoots  []string `json:"chrome_import_allowed_roots"`
}

func LoadView(paths macos.Paths) (View, error) {
	config, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return View{}, err
	}

	envKeys := make([]string, 0, len(config.SidebarProviders))
	seen := map[string]struct{}{}
	for _, provider := range config.SidebarProviders {
		if provider.APIKeyEnv == "" {
			continue
		}
		if _, ok := seen[provider.APIKeyEnv]; ok {
			continue
		}
		seen[provider.APIKeyEnv] = struct{}{}
		envKeys = append(envKeys, provider.APIKeyEnv)
	}

	return View{
		ConfigFile:                paths.ConfigFile,
		DefaultListenAddr:         settings.DefaultListenAddr,
		LoopbackOnlyDefault:       true,
		RemoteControlFlag:         RemoteControlFlag,
		RemoteControlFlagRequired: true,
		SharedProfileManaged:      false,
		SidebarSecretsPersisted:   false,
		SidebarDefaultProvider:    config.SidebarDefaultProvider,
		SidebarProviderCount:      len(config.SidebarProviders),
		SidebarProviderEnvKeys:    envKeys,
		MirrorAllowedRoots: []string{
			paths.ProfilesRoot,
			sourcepaths.DefaultChromeProfilesRoot(paths),
		},
		ChromeImportAllowedRoots: []string{
			sourcepaths.DefaultChromeProfilesRoot(paths),
		},
	}, nil
}

func (v View) Render() string {
	lines := []string{
		fmt.Sprintf("config_file=%s", v.ConfigFile),
		fmt.Sprintf("default_listen_addr=%s", v.DefaultListenAddr),
		fmt.Sprintf("loopback_only_default=%t", v.LoopbackOnlyDefault),
		fmt.Sprintf("remote_control_flag=%s", v.RemoteControlFlag),
		fmt.Sprintf("remote_control_flag_required=%t", v.RemoteControlFlagRequired),
		fmt.Sprintf("shared_profile_managed=%t", v.SharedProfileManaged),
		fmt.Sprintf("sidebar_secrets_persisted=%t", v.SidebarSecretsPersisted),
		fmt.Sprintf("sidebar_default_provider=%s", v.SidebarDefaultProvider),
		fmt.Sprintf("sidebar_provider_count=%d", v.SidebarProviderCount),
		fmt.Sprintf("sidebar_provider_env_key_count=%d", len(v.SidebarProviderEnvKeys)),
		fmt.Sprintf("mirror_allowed_root_count=%d", len(v.MirrorAllowedRoots)),
		fmt.Sprintf("chrome_import_allowed_root_count=%d", len(v.ChromeImportAllowedRoots)),
	}
	for index, key := range v.SidebarProviderEnvKeys {
		lines = append(lines, fmt.Sprintf("sidebar_provider_env_keys[%d]=%s", index, key))
	}
	for index, root := range v.MirrorAllowedRoots {
		lines = append(lines, fmt.Sprintf("mirror_allowed_roots[%d]=%s", index, root))
	}
	for index, root := range v.ChromeImportAllowedRoots {
		lines = append(lines, fmt.Sprintf("chrome_import_allowed_roots[%d]=%s", index, root))
	}
	return strings.Join(lines, "\n") + "\n"
}
