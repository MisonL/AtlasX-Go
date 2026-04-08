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

func TestDefaultBrowserSetCommandRendersStatus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	previous := setDefaultBrowserBundleID
	setDefaultBrowserBundleID = func(bundleID string) (defaultbrowser.Status, error) {
		if bundleID != "com.openai.atlasx" {
			t.Fatalf("unexpected bundle id: %s", bundleID)
		}
		return defaultbrowser.Status{
			Source:        "launchservices",
			HTTPBundleID:  "com.openai.atlasx",
			HTTPRole:      "all",
			HTTPKnown:     true,
			HTTPSBundleID: "com.openai.atlasx",
			HTTPSRole:     "all",
			HTTPSKnown:    true,
			Consistent:    true,
		}, nil
	}
	t.Cleanup(func() {
		setDefaultBrowserBundleID = previous
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"default-browser", "set", "com.openai.atlasx"})
	})
	if err != nil {
		t.Fatalf("run default-browser set failed: %v", err)
	}

	assertContainsAll(t, output,
		"http_bundle_id=com.openai.atlasx",
		"https_bundle_id=com.openai.atlasx",
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

func TestDefaultBrowserSetCommandRejectsMissingBundleID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"default-browser", "set"})
	})
	if err == nil {
		t.Fatal("expected default-browser set to fail")
	}
	if !strings.Contains(err.Error(), "default-browser set requires a bundle id") {
		t.Fatalf("unexpected error: %v", err)
	}
}
