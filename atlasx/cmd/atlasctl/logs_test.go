package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogsStatusCommandRendersSummary(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("user home dir failed: %v", err)
	}
	logsRoot := filepath.Join(home, "Library", "Application Support", "AtlasX", "logs")
	if err := os.MkdirAll(logsRoot, 0o755); err != nil {
		t.Fatalf("mkdir logs root failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(logsRoot, "atlas.log"), []byte("atlas-data"), 0o644); err != nil {
		t.Fatalf("write log file failed: %v", err)
	}

	output, err := captureStdout(t, func() error {
		return run([]string{"logs", "status"})
	})
	if err != nil {
		t.Fatalf("run logs status failed: %v", err)
	}

	assertContainsAll(t, output,
		"logs_root=",
		"present=true",
		"file_count=1",
		"total_bytes=10",
	)
}

func TestLogsStatusCommandRejectsNegativeLimit(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	_, err := captureStdout(t, func() error {
		return run([]string{"logs", "status", "--limit", "-1"})
	})
	if err == nil {
		t.Fatal("expected logs status with negative limit to fail")
	}
	if !strings.Contains(err.Error(), "limit must be >= 0") {
		t.Fatalf("unexpected error: %v", err)
	}
}
