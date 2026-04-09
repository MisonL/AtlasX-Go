package main

import (
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsAuthModeOutputsStructuredView(t *testing.T) {
	restoreCommandTabsClient(t, &stubCommandTabsClient{
		authMode: tabs.AuthModeView{
			ID:                     "tab-1",
			Title:                  "ChatGPT",
			URL:                    "https://chatgpt.com/c/abc123",
			CapturedAt:             "2026-04-09T10:00:00Z",
			Host:                   "chatgpt.com",
			Path:                   "/c/abc123",
			Mode:                   "logged_in",
			Inferred:               true,
			Reason:                 "workspace_signals_observed",
			LoginPromptPresent:     false,
			WorkspaceSignalPresent: true,
			CookieCount:            1,
			CookieNames:            []string{"oai-session"},
			LocalStorageCount:      1,
			LocalStorageKeys:       []string{"atlas:last-project"},
			SessionStorageCount:    0,
			SessionStorageKeys:     []string{},
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "auth-mode", "tab-1"})
	})
	if err != nil {
		t.Fatalf("run tabs auth-mode failed: %v", err)
	}

	for _, fragment := range []string{
		"mode=logged_in",
		"inferred=true",
		"reason=workspace_signals_observed",
		"host=chatgpt.com",
		"path=/c/abc123",
		"cookie_count=1",
		"local_storage_count=1",
		"session_storage_count=0",
		"cookie_names[0]=oai-session",
		"local_storage_keys[0]=atlas:last-project",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsAuthModePrintsCaptureFailure(t *testing.T) {
	context := tabs.PageContext{
		ID:           "tab-1",
		Title:        "ChatGPT",
		URL:          "https://chatgpt.com/auth/login",
		CapturedAt:   "2026-04-09T10:00:00Z",
		TextLimit:    4096,
		CaptureError: "cdp error -32000: capture failed",
	}
	restoreCommandTabsClient(t, &stubCommandTabsClient{
		authModeErr: &tabs.CaptureError{
			Context: context,
			Cause:   errString("cdp error -32000: capture failed"),
		},
	})

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "auth-mode", "tab-1"})
	})
	if err == nil {
		t.Fatal("expected tabs auth-mode to fail")
	}
	if !strings.Contains(output, `capture_error="cdp error -32000: capture failed"`) {
		t.Fatalf("unexpected output: %s", output)
	}
}
