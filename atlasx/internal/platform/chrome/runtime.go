package chrome

import (
	"errors"
	"os"
	"path/filepath"

	"atlasx/internal/platform/macos"
)

var ErrBinaryNotFound = errors.New("chrome runtime not found")

type Detection struct {
	BinaryPath string
	Source     string
	Candidates []string
}

func Detect(preferred string) (Detection, error) {
	candidates, err := candidatePaths()
	if err != nil {
		return Detection{}, err
	}

	if preferred != "" {
		if isExecutable(preferred) {
			return Detection{BinaryPath: preferred, Source: "config", Candidates: candidates}, nil
		}
		return Detection{Candidates: candidates}, errors.New("configured chrome_binary is not executable")
	}

	for _, candidate := range candidates {
		if isExecutable(candidate) {
			return Detection{BinaryPath: candidate, Source: "auto", Candidates: candidates}, nil
		}
	}
	return Detection{Candidates: candidates}, ErrBinaryNotFound
}

func candidatePaths() ([]string, error) {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return nil, err
	}

	return []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		filepath.Join(paths.Home, "Applications", "Google Chrome.app", "Contents", "MacOS", "Google Chrome"),
		"/Applications/Google Chrome Beta.app/Contents/MacOS/Google Chrome Beta",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		filepath.Join(paths.Home, "Applications", "Chromium.app", "Contents", "MacOS", "Chromium"),
	}, nil
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Mode()&0o111 != 0
}
