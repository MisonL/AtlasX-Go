package sidebar

import (
	"errors"
	"testing"

	"atlasx/internal/platform/macos"
)

func TestSaveRuntimeResultAndLoadRuntimeState(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}
	if err := SaveRuntimeResult(paths, "trace-123", errors.New("upstream failed")); err != nil {
		t.Fatalf("save runtime result failed: %v", err)
	}

	state, err := LoadRuntimeState(paths)
	if err != nil {
		t.Fatalf("load runtime state failed: %v", err)
	}
	if state.LastTraceID != "trace-123" {
		t.Fatalf("unexpected runtime state: %+v", state)
	}
	if state.LastError != "upstream failed" || state.LastErrorAt == "" {
		t.Fatalf("unexpected runtime state: %+v", state)
	}
	if state.TimeoutMS == 0 || state.TokenBudget == 0 {
		t.Fatalf("unexpected runtime defaults: %+v", state)
	}
}
