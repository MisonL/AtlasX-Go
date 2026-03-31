package chrome

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"atlasx/internal/platform/macos"
)

var ErrBinaryNotFound = errors.New("chrome runtime not found")

type Detection struct {
	BinaryPath string
	Source     string
	Candidates []string
}

func Detect(preferred string) (Detection, error) {
	paths, err := macos.DiscoverPaths()
	if err != nil {
		return Detection{}, err
	}
	return DetectWithPaths(preferred, paths)
}

func DetectWithPaths(preferred string, paths macos.Paths) (Detection, error) {
	candidates := candidatePaths(paths)
	candidateList := flattenCandidatePaths(candidates)
	if preferred != "" {
		if isExecutable(preferred) {
			return Detection{BinaryPath: preferred, Source: "config", Candidates: candidateList}, nil
		}
		return Detection{Candidates: candidateList}, errors.New("configured chrome_binary is not executable")
	}

	for _, candidate := range candidates {
		if isExecutable(candidate.Path) {
			return Detection{BinaryPath: candidate.Path, Source: candidate.Source, Candidates: candidateList}, nil
		}
	}
	return Detection{Candidates: candidateList}, ErrBinaryNotFound
}

type candidatePath struct {
	Path   string
	Source string
}

func candidatePaths(paths macos.Paths) []candidatePath {
	return []candidatePath{
		{Path: ManagedBinaryPath(paths), Source: "managed_auto"},
		{Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", Source: "system_auto"},
		{Path: filepath.Join(paths.Home, "Applications", "Google Chrome.app", "Contents", "MacOS", "Google Chrome"), Source: "system_auto"},
		{Path: "/Applications/Google Chrome Beta.app/Contents/MacOS/Google Chrome Beta", Source: "system_auto"},
		{Path: "/Applications/Chromium.app/Contents/MacOS/Chromium", Source: "system_auto"},
		{Path: filepath.Join(paths.Home, "Applications", "Chromium.app", "Contents", "MacOS", "Chromium"), Source: "system_auto"},
	}
}

func flattenCandidatePaths(candidates []candidatePath) []string {
	paths := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		paths = append(paths, candidate.Path)
	}
	return paths
}

func ManagedBinaryPath(paths macos.Paths) string {
	return filepath.Join(paths.RuntimeRoot, "Chromium.app", "Contents", "MacOS", "Chromium")
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir() && info.Mode()&0o111 != 0
}

func AppBundlePath(binaryPath string) string {
	index := strings.Index(binaryPath, ".app/")
	if index == -1 {
		return ""
	}
	return binaryPath[:index+4]
}
