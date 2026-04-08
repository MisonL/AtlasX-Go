package main

import (
	"strings"
	"testing"

	"atlasx/internal/defaultbrowser"
)

func TestDefaultBrowserStatusCommandRendersStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	previous := readDefaultBrowserStatus
	readDefaultBrowserStatus = func() (defaultbrowser.Status, error) {
		return defaultbrowser.Status{
			Source:        "launchservices",
			HTTPBundleID:  "org.mozilla.firefox",
			HTTPRole:      "all",
			HTTPKnown:     true,
			HTTPSBundleID: "org.mozilla.firefox",
			HTTPSRole:     "all",
			HTTPSKnown:    true,
			Consistent:    true,
		}, nil
	}
	t.Cleanup(func() {
		readDefaultBrowserStatus = previous
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"default-browser", "status"})
	})
	if err != nil {
		t.Fatalf("run default-browser status failed: %v", err)
	}

	assertContainsAll(t, output,
		"http_bundle_id=org.mozilla.firefox",
		"https_bundle_id=org.mozilla.firefox",
		"consistent=true",
	)
}

func TestDefaultBrowserStatusCommandRejectsUnknownSubcommand(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"default-browser", "inspect"})
	})
	if err == nil {
		t.Fatal("expected default-browser inspect to fail")
	}
	if !strings.Contains(err.Error(), `unknown default-browser subcommand "inspect"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
