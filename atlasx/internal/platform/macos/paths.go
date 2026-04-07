package macos

import (
	"os"
	"path/filepath"
)

const appName = "AtlasX"

type Paths struct {
	Home                   string
	SupportRoot            string
	ConfigFile             string
	ProfilesRoot           string
	LogsRoot               string
	MemoryRoot             string
	MemoryEventsFile       string
	StateRoot              string
	SessionFile            string
	MirrorsRoot            string
	MirrorFile             string
	ImportsRoot            string
	RuntimeRoot            string
	RuntimeManifestFile    string
	RuntimeInstallPlanFile string
}

func DiscoverPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	supportRoot := filepath.Join(home, "Library", "Application Support", appName)
	memoryRoot := filepath.Join(supportRoot, "memory")
	stateRoot := filepath.Join(supportRoot, "state")
	mirrorsRoot := filepath.Join(supportRoot, "mirrors")
	importsRoot := filepath.Join(supportRoot, "imports")
	runtimeRoot := filepath.Join(supportRoot, "runtime")
	return Paths{
		Home:                   home,
		SupportRoot:            supportRoot,
		ConfigFile:             filepath.Join(supportRoot, "config.json"),
		ProfilesRoot:           filepath.Join(supportRoot, "profiles"),
		LogsRoot:               filepath.Join(supportRoot, "logs"),
		MemoryRoot:             memoryRoot,
		MemoryEventsFile:       filepath.Join(memoryRoot, "events.jsonl"),
		StateRoot:              stateRoot,
		SessionFile:            filepath.Join(stateRoot, "webapp-session.json"),
		MirrorsRoot:            mirrorsRoot,
		MirrorFile:             filepath.Join(mirrorsRoot, "browser-data.json"),
		ImportsRoot:            importsRoot,
		RuntimeRoot:            runtimeRoot,
		RuntimeManifestFile:    filepath.Join(runtimeRoot, "manifest.json"),
		RuntimeInstallPlanFile: filepath.Join(runtimeRoot, "install-plan.json"),
	}, nil
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
