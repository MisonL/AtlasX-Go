package profile

import (
	"fmt"
	"os"
	"strings"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
)

type View struct {
	ProfilesRoot        string `json:"profiles_root"`
	DefaultProfile      string `json:"default_profile"`
	SelectedMode        string `json:"selected_mode"`
	SelectedUserDataDir string `json:"selected_user_data_dir"`
	IsolatedUserDataDir string `json:"isolated_user_data_dir"`
	IsolatedPresent     bool   `json:"isolated_present"`
	SharedManaged       bool   `json:"shared_managed"`
}

func LoadView(paths macos.Paths) (View, error) {
	config, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return View{}, err
	}

	store := NewStore(paths.ProfilesRoot)
	selected, err := store.Resolve(config.DefaultProfile)
	if err != nil {
		return View{}, err
	}
	isolated, err := store.Resolve(ModeIsolated)
	if err != nil {
		return View{}, err
	}

	view := View{
		ProfilesRoot:        paths.ProfilesRoot,
		DefaultProfile:      config.DefaultProfile,
		SelectedMode:        selected.Mode,
		SelectedUserDataDir: selected.UserDataDir,
		IsolatedUserDataDir: isolated.UserDataDir,
		SharedManaged:       false,
	}
	if view.IsolatedUserDataDir != "" {
		if _, err := os.Stat(view.IsolatedUserDataDir); err == nil {
			view.IsolatedPresent = true
		} else if err != nil && !os.IsNotExist(err) {
			return View{}, err
		}
	}
	return view, nil
}

func (v View) Render() string {
	return strings.Join([]string{
		fmt.Sprintf("profiles_root=%s", v.ProfilesRoot),
		fmt.Sprintf("default_profile=%s", v.DefaultProfile),
		fmt.Sprintf("selected_mode=%s", v.SelectedMode),
		fmt.Sprintf("selected_user_data_dir=%s", v.SelectedUserDataDir),
		fmt.Sprintf("isolated_user_data_dir=%s", v.IsolatedUserDataDir),
		fmt.Sprintf("isolated_present=%t", v.IsolatedPresent),
		fmt.Sprintf("shared_managed=%t", v.SharedManaged),
	}, "\n") + "\n"
}
