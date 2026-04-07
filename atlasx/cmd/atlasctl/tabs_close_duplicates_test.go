package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsCloseDuplicatesOutputsClosedTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		duplicateClose: tabs.CloseDuplicatesResult{
			Returned: 1,
			Groups: []tabs.DuplicateCloseGroup{
				{
					URL:             "https://openai.com/docs",
					KeptTargetID:    "tab-1",
					ClosedTargetIDs: []string{"tab-2"},
					Returned:        1,
				},
			},
			ClosedTargets: []string{"tab-2"},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "close-duplicates"})
	})
	if err != nil {
		t.Fatalf("run tabs close-duplicates failed: %v", err)
	}
	for _, fragment := range []string{"returned=1", "url=https://openai.com/docs", "kept_target_id=tab-1", "id=tab-2"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsCloseDuplicatesSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		duplicateCloseErr: errString("unexpected status 500"),
	})

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "close-duplicates"})
	})
	if err == nil {
		t.Fatal("expected tabs close-duplicates to fail")
	}
	if !strings.Contains(err.Error(), "unexpected status 500") {
		t.Fatalf("unexpected error: %v", err)
	}
}
