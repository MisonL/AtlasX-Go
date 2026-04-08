package memory

import (
	"fmt"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
)

type Controls struct {
	ConfigFile            string `json:"config_file"`
	PersistEnabled        bool   `json:"persist_enabled"`
	PageVisibilityEnabled bool   `json:"page_visibility_enabled"`
}

type ControlsUpdate struct {
	PersistEnabled        *bool `json:"persist_enabled,omitempty"`
	PageVisibilityEnabled *bool `json:"page_visibility_enabled,omitempty"`
}

func LoadControls(paths macos.Paths) (Controls, error) {
	config, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return Controls{}, err
	}

	return Controls{
		ConfigFile:            paths.ConfigFile,
		PersistEnabled:        config.MemoryPersistEnabledValue(),
		PageVisibilityEnabled: config.MemoryPageVisibilityEnabledValue(),
	}, nil
}

func SetPersistEnabled(paths macos.Paths, enabled bool) (Controls, error) {
	return UpdateControls(paths, ControlsUpdate{PersistEnabled: settings.Bool(enabled)})
}

func SetPageVisibilityEnabled(paths macos.Paths, enabled bool) (Controls, error) {
	return UpdateControls(paths, ControlsUpdate{PageVisibilityEnabled: settings.Bool(enabled)})
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
