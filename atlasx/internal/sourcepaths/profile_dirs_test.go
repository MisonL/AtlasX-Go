package sourcepaths

import (
	"os"
	"path/filepath"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestValidateMirrorProfileDirAllowsAtlasXProfilesRoot(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	target := filepath.Join(paths.ProfilesRoot, "webapp-isolated")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	if err := ValidateMirrorProfileDir(paths, target); err != nil {
		t.Fatalf("expected atlasx profile dir to pass: %v", err)
	}
}

func TestValidateMirrorProfileDirAllowsChromeProfilesRoot(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	target := filepath.Join(DefaultChromeProfilesRoot(paths), "Default")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	if err := ValidateMirrorProfileDir(paths, target); err != nil {
		t.Fatalf("expected chrome profile dir to pass: %v", err)
	}
}

func TestValidateMirrorProfileDirRejectsOutsideRoots(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	target := filepath.Join(t.TempDir(), "outside", "profile")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	if err := ValidateMirrorProfileDir(paths, target); err == nil {
		t.Fatal("expected outside mirror dir to fail")
	}
}

func TestValidateChromeImportSourceDirRejectsAtlasXProfilesRoot(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	target := filepath.Join(paths.ProfilesRoot, "webapp-isolated")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	if err := ValidateChromeImportSourceDir(paths, target); err == nil {
		t.Fatal("expected atlasx profile dir to fail for chrome import")
	}
}
