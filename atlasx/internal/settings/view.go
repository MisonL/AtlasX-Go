package settings

import (
	"fmt"
	"strings"

	"atlasx/internal/platform/macos"
)

type View struct {
	ConfigFile             string                  `json:"config_file"`
	ChromeBinary           string                  `json:"chrome_binary"`
	DefaultProfile         string                  `json:"default_profile"`
	ListenAddr             string                  `json:"listen_addr"`
	WebAppURL              string                  `json:"web_app_url"`
	SidebarProvider        string                  `json:"sidebar_provider"`
	SidebarModel           string                  `json:"sidebar_model"`
	SidebarBaseURL         string                  `json:"sidebar_base_url"`
	SidebarDefaultProvider string                  `json:"sidebar_default_provider"`
	SidebarProviders       []SidebarProviderConfig `json:"sidebar_providers"`
}

func LoadView(paths macos.Paths) (View, error) {
	config, err := NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return View{}, err
	}

	return View{
		ConfigFile:             paths.ConfigFile,
		ChromeBinary:           config.ChromeBinary,
		DefaultProfile:         config.DefaultProfile,
		ListenAddr:             config.ListenAddr,
		WebAppURL:              config.WebAppURL,
		SidebarProvider:        config.SidebarProvider,
		SidebarModel:           config.SidebarModel,
		SidebarBaseURL:         config.SidebarBaseURL,
		SidebarDefaultProvider: config.SidebarDefaultProvider,
		SidebarProviders:       append([]SidebarProviderConfig(nil), config.SidebarProviders...),
	}, nil
}

func (v View) Render() string {
	lines := []string{
		"AtlasX Settings",
		fmt.Sprintf("config_file=%s", v.ConfigFile),
		fmt.Sprintf("chrome_binary=%s", v.ChromeBinary),
		fmt.Sprintf("default_profile=%s", v.DefaultProfile),
		fmt.Sprintf("listen_addr=%s", v.ListenAddr),
		fmt.Sprintf("web_app_url=%s", v.WebAppURL),
		fmt.Sprintf("sidebar_provider=%s", v.SidebarProvider),
		fmt.Sprintf("sidebar_model=%s", v.SidebarModel),
		fmt.Sprintf("sidebar_base_url=%s", v.SidebarBaseURL),
		fmt.Sprintf("sidebar_default_provider=%s", v.SidebarDefaultProvider),
		fmt.Sprintf("sidebar_provider_count=%d", len(v.SidebarProviders)),
	}
	for index, provider := range v.SidebarProviders {
		lines = append(lines,
			fmt.Sprintf("sidebar_provider[%d].id=%s", index, provider.ID),
			fmt.Sprintf("sidebar_provider[%d].provider=%s", index, provider.Provider),
			fmt.Sprintf("sidebar_provider[%d].model=%s", index, provider.Model),
			fmt.Sprintf("sidebar_provider[%d].base_url=%s", index, provider.BaseURL),
			fmt.Sprintf("sidebar_provider[%d].api_key_env=%s", index, provider.APIKeyEnv),
		)
	}
	return strings.Join(lines, "\n") + "\n"
}
