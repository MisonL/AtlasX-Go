package main

import (
	"errors"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

type stubCommandTabsWindowClient struct {
	stubCommandTabsClient
	openWindowTargetURL string
	openWindowTarget    string
	openWindowErr       error
}

func (s *stubCommandTabsWindowClient) OpenWindow(url string) (tabs.Target, error) {
	s.openWindowTargetURL = url
	if s.openWindowErr != nil {
		return tabs.Target{}, s.openWindowErr
	}
	return tabs.Target{ID: "tab-window", Type: "page", Title: "OpenAI", URL: s.openWindowTarget}, nil
}

func TestTabsOpenWindowOutputsTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	client := &stubCommandTabsWindowClient{openWindowTarget: "https://openai.com"}
	restoreCommandTabsClient(t, client)

	output, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-window", "https://openai.com"})
	})
	if err != nil {
		t.Fatalf("run tabs open-window failed: %v", err)
	}
	if client.openWindowTargetURL != "https://openai.com" {
		t.Fatalf("unexpected open-window url: %s", client.openWindowTargetURL)
	}
	for _, fragment := range []string{"id=tab-window", `title="OpenAI"`, "url=https://openai.com"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %q, got %s", fragment, output)
		}
	}
}

func TestTabsOpenWindowRejectsDangerousScheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	client := &stubCommandTabsWindowClient{
		openWindowTarget: "https://openai.com",
		openWindowErr:    errors.New(`unsupported url scheme "javascript"`),
	}
	restoreCommandTabsClient(t, client)

	_, err := captureStdout(t, func() error {
		return run([]string{"tabs", "open-window", "javascript:alert(1)"})
	})
	if err == nil {
		t.Fatal("expected tabs open-window to fail")
	}
	if !strings.Contains(err.Error(), "unsupported url scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
}
