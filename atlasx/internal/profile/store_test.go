package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveIsolatedCreatesManagedDirectory(t *testing.T) {
	root := filepath.Join(t.TempDir(), "profiles")

	selection, err := NewStore(root).Resolve(ModeIsolated)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if selection.UserDataDir == "" {
		t.Fatal("expected isolated user data dir")
	}

	if _, err := os.Stat(selection.UserDataDir); err != nil {
		t.Fatalf("isolated directory not created: %v", err)
	}
}

func TestResolveSharedLeavesUserDataDirEmpty(t *testing.T) {
	selection, err := NewStore(t.TempDir()).Resolve(ModeShared)
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}

	if selection.UserDataDir != "" {
		t.Fatalf("expected empty user data dir, got %s", selection.UserDataDir)
	}
}
