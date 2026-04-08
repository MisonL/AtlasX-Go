package memory

import (
	"os"
	"testing"

	"atlasx/internal/platform/macos"
	"atlasx/internal/settings"
)

func TestLoadControlsDefaultsToEnabled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	controls, err := LoadControls(paths)
	if err != nil {
		t.Fatalf("load controls failed: %v", err)
	}
	if !controls.PersistEnabled {
		t.Fatalf("expected persist enabled by default: %+v", controls)
	}
	if !controls.PageVisibilityEnabled {
		t.Fatalf("expected page visibility enabled by default: %+v", controls)
	}
}

func TestSetPersistEnabledUpdatesConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	controls, err := SetPersistEnabled(paths, false)
	if err != nil {
		t.Fatalf("set persist enabled failed: %v", err)
	}
	if controls.PersistEnabled {
		t.Fatalf("expected persist disabled: %+v", controls)
	}

	config, err := settings.NewStore(paths.ConfigFile).Load()
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if config.MemoryPersistEnabledValue() {
		t.Fatalf("expected config persist disabled: %+v", config)
	}
	if !config.MemoryPageVisibilityEnabledValue() {
		t.Fatalf("expected page visibility to remain enabled: %+v", config)
	}
}

func TestSetPageVisibilityEnabledUpdatesConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	controls, err := SetPageVisibilityEnabled(paths, false)
	if err != nil {
		t.Fatalf("set page visibility failed: %v", err)
	}
	if controls.PageVisibilityEnabled {
		t.Fatalf("expected page visibility disabled: %+v", controls)
	}

	config, err := settings.NewStore(paths.ConfigFile).Load()
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if config.MemoryPageVisibilityEnabledValue() {
		t.Fatalf("expected config page visibility disabled: %+v", config)
	}
	if !config.MemoryPersistEnabledValue() {
		t.Fatalf("expected persist to remain enabled: %+v", config)
	}
}

func TestAppendPageCaptureControlledSkipsWriteWhenDisabled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := settings.NewStore(paths.ConfigFile).Save(settings.Config{MemoryPersistEnabled: settings.Bool(false)}); err != nil {
		t.Fatalf("save config failed: %v", err)
	}

	persisted, err := AppendPageCaptureControlled(paths, PageCaptureInput{
		OccurredAt: "2026-04-08T10:00:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
	})
	if err != nil {
		t.Fatalf("append page capture controlled failed: %v", err)
	}
	if persisted {
		t.Fatalf("expected no persistence when disabled")
	}
	if _, err := os.Stat(paths.MemoryEventsFile); !os.IsNotExist(err) {
		t.Fatalf("expected no events file, got err=%v", err)
	}
}

func TestAppendQATurnControlledPersistsWhenEnabled(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	persisted, err := AppendQATurnControlled(paths, QATurnInput{
		OccurredAt: "2026-04-08T10:00:00Z",
		TabID:      "tab-1",
		Title:      "Atlas",
		URL:        "https://chatgpt.com/atlas",
		Question:   "q",
		Answer:     "a",
	})
	if err != nil {
		t.Fatalf("append qa turn controlled failed: %v", err)
	}
	if !persisted {
		t.Fatalf("expected persistence when enabled")
	}

	events, err := Load(paths)
	if err != nil {
		t.Fatalf("load events failed: %v", err)
	}
	if len(events) != 1 || events[0].Kind != EventKindQATurn {
		t.Fatalf("unexpected events: %+v", events)
	}
}

func TestParsePersistValueSupportsCommonWords(t *testing.T) {
	cases := map[string]bool{
		"enabled":  true,
		"true":     true,
		"on":       true,
		"1":        true,
		"disabled": false,
		"false":    false,
		"off":      false,
		"0":        false,
	}

	for raw, want := range cases {
		got, err := ParsePersistValue(raw)
		if err != nil {
			t.Fatalf("parse persist value %q failed: %v", raw, err)
		}
		if got != want {
			t.Fatalf("parse persist value %q = %t, want %t", raw, got, want)
		}
	}
}

func TestParsePageVisibilityValueSupportsCommonWords(t *testing.T) {
	cases := map[string]bool{
		"visible":  true,
		"enabled":  true,
		"true":     true,
		"on":       true,
		"1":        true,
		"hidden":   false,
		"disabled": false,
		"false":    false,
		"off":      false,
		"0":        false,
	}

	for raw, want := range cases {
		got, err := ParsePageVisibilityValue(raw)
		if err != nil {
			t.Fatalf("parse page visibility value %q failed: %v", raw, err)
		}
		if got != want {
			t.Fatalf("parse page visibility value %q = %t, want %t", raw, got, want)
		}
	}
}
