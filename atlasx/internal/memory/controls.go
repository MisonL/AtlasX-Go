package memory

import (
	"fmt"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
)

type Controls struct {
	ConfigFile     string `json:"config_file"`
	PersistEnabled bool   `json:"persist_enabled"`
}

func LoadControls(paths macos.Paths) (Controls, error) {
	config, err := settings.NewStore(paths.ConfigFile).Bootstrap()
	if err != nil {
		return Controls{}, err
	}

	return Controls{
		ConfigFile:     paths.ConfigFile,
		PersistEnabled: config.MemoryPersistEnabledValue(),
	}, nil
}

func SetPersistEnabled(paths macos.Paths, enabled bool) (Controls, error) {
	store := settings.NewStore(paths.ConfigFile)
	config, err := store.Bootstrap()
	if err != nil {
		return Controls{}, err
	}

	config.MemoryPersistEnabled = settings.Bool(enabled)
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
