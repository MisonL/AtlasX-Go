package launcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"atlasx/internal/platform/macos"
	"atlasx/internal/profile"
)

func TestBuildArgsIncludesAppModeAndManagedProfile(t *testing.T) {
	args := BuildArgs(
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		profile.Selection{Mode: profile.ModeIsolated, UserDataDir: "/tmp/atlasx"},
		"https://chatgpt.com/atlas?get-started",
	)

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--app=https://chatgpt.com/atlas?get-started") {
		t.Fatalf("missing app arg: %s", joined)
	}
	if !strings.Contains(joined, "--user-data-dir=/tmp/atlasx") {
		t.Fatalf("missing user data dir: %s", joined)
	}
}

func TestBuildArgsSkipsManagedProfileForSharedMode(t *testing.T) {
	args := BuildArgs(
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		profile.Selection{Mode: profile.ModeShared},
		"https://chatgpt.com/atlas?get-started",
	)

	joined := strings.Join(args, " ")
	if strings.Contains(joined, "--user-data-dir=") {
		t.Fatalf("unexpected user data dir for shared mode: %s", joined)
	}
}

func TestStatusAbsentWithoutStateFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	report, err := Status(paths)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if report.Present {
		t.Fatal("expected absent session")
	}
}

func TestSaveAndLoadState(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	state := State{
		Mode:        profile.ModeIsolated,
		Managed:     true,
		BinaryPath:  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		URL:         "https://chatgpt.com/atlas?get-started",
		UserDataDir: filepath.Join(t.TempDir(), "profile"),
	}
	if err := SaveState(paths, state); err != nil {
		t.Fatalf("save state failed: %v", err)
	}

	loaded, err := LoadState(paths)
	if err != nil {
		t.Fatalf("load state failed: %v", err)
	}
	if loaded.UserDataDir != state.UserDataDir {
		t.Fatalf("unexpected user data dir: %s", loaded.UserDataDir)
	}

	if _, err := os.Stat(paths.SessionFile); err != nil {
		t.Fatalf("state file not created: %v", err)
	}
}
