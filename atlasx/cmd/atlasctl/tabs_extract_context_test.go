package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsExtractContextOutputsStructuredSemanticContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		semanticContext: tabs.SemanticContext{
			ID:               "tab-1",
			Title:            "Atlas",
			URL:              "https://chatgpt.com/atlas",
			CapturedAt:       "2026-04-07T13:00:00Z",
			Returned:         3,
			HeadingsReturned: 1,
			LinksReturned:    1,
			FormsReturned:    1,
			Headings:         []tabs.SemanticHeading{{Level: 1, Text: "Atlas"}},
			Links:            []tabs.SemanticLink{{Text: "OpenAI Docs", URL: "https://platform.openai.com/docs"}},
			Forms:            []tabs.SemanticForm{{Action: "https://chatgpt.com/search", Method: "GET", InputCount: 2}},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "extract-context", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs extract-context failed: %v", err)
	}
	for _, fragment := range []string{
		"returned=3",
		"headings_returned=1",
		"links_returned=1",
		"forms_returned=1",
		`heading_index=0 level=1 text="Atlas"`,
		`link_index=0 text="OpenAI Docs" url=https://platform.openai.com/docs`,
		`form_index=0 action=https://chatgpt.com/search method=GET input_count=2`,
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsExtractContextReturnsEmptyCollections(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreCommandTabsClient(t, &stubCommandTabsClient{
		semanticContext: tabs.SemanticContext{
			ID:               "tab-1",
			Title:            "Atlas",
			URL:              "about:blank",
			CapturedAt:       "2026-04-07T13:00:00Z",
			Returned:         0,
			HeadingsReturned: 0,
			LinksReturned:    0,
			FormsReturned:    0,
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "extract-context", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs extract-context failed: %v", err)
	}
	if !strings.Contains(output, "returned=0") {
		t.Fatalf("unexpected output: %s", output)
	}
	if strings.Contains(output, "heading_index=") || strings.Contains(output, "link_index=") || strings.Contains(output, "form_index=") {
		t.Fatalf("expected no semantic entries, got %s", output)
	}
}

func TestTabsExtractContextFailurePrintsCaptureError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	context := tabs.SemanticContext{
		ID:           "tab-1",
		Title:        "Atlas",
		URL:          "https://chatgpt.com/atlas",
		CapturedAt:   "2026-04-07T13:00:00Z",
		CaptureError: "cdp error -32000: semantic capture failed",
	}
	restoreCommandTabsClient(t, &stubCommandTabsClient{
		semanticContext: context,
		semanticErr: &tabs.SemanticCaptureError{
			Context: context,
			Cause:   errString("cdp error -32000: semantic capture failed"),
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "extract-context", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs extract-context to fail")
	}
	if !strings.Contains(output, `capture_error="cdp error -32000: semantic capture failed"`) {
		t.Fatalf("unexpected output: %s", output)
	}
}
