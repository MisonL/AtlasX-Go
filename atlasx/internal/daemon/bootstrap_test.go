package daemon

import (
	"testing"

	"atlasx/internal/managedruntime"
	"atlasx/internal/platform/macos"
)

func TestBootstrapIncludesRuntimeBinaryState(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	sourceBundle := createDaemonFakeChromiumBundle(t)
	if _, err := managedruntime.StageLocal(paths, managedruntime.StageOptions{
		BundlePath: sourceBundle,
		Version:    "123.0.0",
		Channel:    "local",
	}); err != nil {
		t.Fatalf("stage local failed: %v", err)
	}

	status, err := Bootstrap()
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	if !status.RuntimeBundlePresent || !status.RuntimeBinaryPresent || !status.RuntimeBinaryExecutable {
		t.Fatalf("unexpected runtime state: %+v", status)
	}
}
