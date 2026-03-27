package macos

import (
	"os"
	"path/filepath"
)

const appName = "AtlasX"

type Paths struct {
	Home         string
	SupportRoot  string
	ConfigFile   string
	ProfilesRoot string
	LogsRoot     string
}

func DiscoverPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	supportRoot := filepath.Join(home, "Library", "Application Support", appName)
	return Paths{
		Home:         home,
		SupportRoot:  supportRoot,
		ConfigFile:   filepath.Join(supportRoot, "config.json"),
		ProfilesRoot: filepath.Join(supportRoot, "profiles"),
		LogsRoot:     filepath.Join(supportRoot, "logs"),
	}, nil
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
