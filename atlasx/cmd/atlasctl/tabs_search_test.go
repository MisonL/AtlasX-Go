package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsSearchOutputsMatches(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		searchTargets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas Docs", URL: "https://openai.com/docs/atlas"},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "search", "atlas"})
	})
	if err != nil {
		t.Fatalf("run tabs search failed: %v", err)
	}
	for _, fragment := range []string{"returned=1", "id=tab-1", `title="Atlas Docs"`} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsSearchRejectsMissingQuery(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "search"})
	})
	if err == nil {
		t.Fatal("expected tabs search to fail")
	}
	if !strings.Contains(err.Error(), "missing query for tabs search") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTabsSearchSurfacesSearchFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		searchErr: errString("search failed"),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "search", "atlas"})
	})
	if err == nil {
		t.Fatal("expected tabs search to fail")
	}
	if !strings.Contains(err.Error(), "search failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
