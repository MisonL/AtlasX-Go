package launcher

import (
	"strings"
	"testing"

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
