package profile

import (
	"strings"
	"testing"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
)

func TestLoadViewBootstrapsIsolatedProfile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	view, err := LoadView(paths)
	if err != nil {
		t.Fatalf("load view failed: %v", err)
	}

	if view.ProfilesRoot != paths.ProfilesRoot {
		t.Fatalf("unexpected profiles root: %+v", view)
	}
	if view.DefaultProfile != settings.DefaultProfile || view.SelectedMode != ModeIsolated {
		t.Fatalf("unexpected selected profile: %+v", view)
	}
	if view.SelectedUserDataDir == "" || view.IsolatedUserDataDir == "" || !view.IsolatedPresent {
		t.Fatalf("expected isolated profile directory: %+v", view)
	}
	if view.SharedManaged {
		t.Fatalf("shared profile should not be managed: %+v", view)
	}
}

func TestViewRenderIncludesSelectedProfileFields(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	view, err := LoadView(paths)
	if err != nil {
		t.Fatalf("load view failed: %v", err)
	}

	rendered := view.Render()
	for _, fragment := range []string{
		"profiles_root=" + paths.ProfilesRoot,
		"default_profile=isolated",
		"selected_mode=isolated",
		"isolated_present=true",
		"shared_managed=false",
	} {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("expected rendered profile to contain %q, got %s", fragment, rendered)
		}
	}
}
