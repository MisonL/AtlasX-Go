package sourcepaths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"atlasx/internal/platform/macos"
)

func DefaultChromeProfilesRoot(paths macos.Paths) string {
	return filepath.Join(paths.Home, "Library", "Application Support", "Google", "Chrome")
}

func ValidateMirrorProfileDir(paths macos.Paths, profileDir string) error {
	return validateAllowedDir(profileDir, []string{
		paths.ProfilesRoot,
		DefaultChromeProfilesRoot(paths),
	})
}

func ValidateChromeImportSourceDir(paths macos.Paths, profileDir string) error {
	return validateAllowedDir(profileDir, []string{
		DefaultChromeProfilesRoot(paths),
	})
}

func validateAllowedDir(target string, allowedRoots []string) error {
	normalizedTarget, err := normalizePath(target)
	if err != nil {
		return err
	}

	for _, root := range allowedRoots {
		normalizedRoot, err := normalizePath(root)
		if err != nil {
			return err
		}
		if withinRoot(normalizedRoot, normalizedTarget) {
			return nil
		}
	}

	return fmt.Errorf("path %q is outside allowed profile roots", target)
}

func normalizePath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path is required")
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if _, err := os.Lstat(absolute); err == nil {
		resolved, err := filepath.EvalSymlinks(absolute)
		if err != nil {
			return "", err
		}
		absolute = resolved
	}

	return filepath.Clean(absolute), nil
}

func withinRoot(root string, target string) bool {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	if relative == "." {
		return true
	}
	return !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) && relative != ".."
}
