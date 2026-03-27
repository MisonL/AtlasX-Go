package profile

import (
	"fmt"
	"path/filepath"

	"atlasx/internal/platform/macos"
)

const (
	ModeIsolated = "isolated"
	ModeShared   = "shared"
)

type Selection struct {
	Mode        string
	UserDataDir string
}

type Store struct {
	root string
}

func NewStore(root string) Store {
	return Store{root: root}
}

func (s Store) Ensure() error {
	return macos.EnsureDir(s.root)
}

func (s Store) Resolve(mode string) (Selection, error) {
	switch mode {
	case ModeIsolated:
		if err := s.Ensure(); err != nil {
			return Selection{}, err
		}
		return Selection{
			Mode:        mode,
			UserDataDir: filepath.Join(s.root, "webapp-isolated"),
		}, macos.EnsureDir(filepath.Join(s.root, "webapp-isolated"))
	case ModeShared:
		return Selection{Mode: mode}, nil
	default:
		return Selection{}, fmt.Errorf("unsupported profile mode %q", mode)
	}
}
