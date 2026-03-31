package sidebar

import (
	"errors"
	"testing"

	"atlasx/internal/settings"
)

func TestStatusWithoutProvider(t *testing.T) {
	status := FromSettings(settings.Config{}).Status()
	if status.Configured {
		t.Fatal("expected unconfigured sidebar status")
	}
	if status.Reason == "" {
		t.Fatal("expected explicit reason")
	}
}

func TestStatusWithIncompleteConfig(t *testing.T) {
	status := FromSettings(settings.Config{
		SidebarProvider: "openai",
	}).Status()
	if !status.Configured {
		t.Fatal("expected configured status")
	}
	if status.Reason != "sidebar qa config is incomplete" {
		t.Fatalf("unexpected reason: %s", status.Reason)
	}
}

func TestAskRejectsUnconfiguredBackend(t *testing.T) {
	err := Config{}.Ask(AskRequest{
		TabID:    "tab-1",
		Question: "summarize this page",
	})
	if !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("unexpected error: %v", err)
	}
}
