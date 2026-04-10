package managedruntime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocateExtractedBundleIgnoresNestedHelperApps(t *testing.T) {
	extractRoot := t.TempDir()
	mainBundle := createFakeBundle(t, "Google Chrome for Testing.app", "Google Chrome for Testing")
	targetMainBundle := filepath.Join(extractRoot, filepath.Base(mainBundle))
	if err := os.Rename(mainBundle, targetMainBundle); err != nil {
		t.Fatalf("move main bundle failed: %v", err)
	}

	helperBundle := filepath.Join(
		targetMainBundle,
		"Contents",
		"Frameworks",
		"Google Chrome for Testing Framework.framework",
		"Versions",
		"146.0.7680.178",
		"Helpers",
		"Google Chrome for Testing Helper.app",
	)
	helperBinary := filepath.Join(helperBundle, "Contents", "MacOS", "Google Chrome for Testing Helper")
	if err := os.MkdirAll(filepath.Dir(helperBinary), 0o755); err != nil {
		t.Fatalf("mkdir helper bundle failed: %v", err)
	}
	if err := os.WriteFile(helperBinary, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write helper binary failed: %v", err)
	}

	bundlePath, err := locateExtractedBundle(extractRoot)
	if err != nil {
		t.Fatalf("locate extracted bundle failed: %v", err)
	}
	if bundlePath != targetMainBundle {
		t.Fatalf("unexpected bundle path: got=%s want=%s", bundlePath, targetMainBundle)
	}
}
