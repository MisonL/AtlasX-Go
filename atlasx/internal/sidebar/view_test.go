package sidebar

import (
	"strings"
	"testing"
)

func TestStatusRenderIncludesProviderRegistry(t *testing.T) {
	status := Status{
		Configured:      true,
		Ready:           true,
		DefaultProvider: "primary",
		Provider:        "openai",
		Model:           "gpt-5.4",
		BaseURL:         "https://api.openai.com/v1",
		APIKeyEnv:       "OPENAI_API_KEY",
		Providers: []ProviderStatus{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   "https://api.openai.com/v1",
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
		TimeoutMS:     30000,
		RetryAttempts: 1,
		TokenBudget:   12000,
		LastTraceID:   "trace-1",
		LastError:     "boom",
		LastErrorAt:   "2026-04-07T10:00:00Z",
	}

	rendered := status.Render()
	for _, fragment := range []string{
		"configured=true",
		"ready=true",
		"default_provider=primary",
		"provider_count=1",
		"provider[0].id=primary",
		"last_trace_id=trace-1",
		"last_error=boom",
	} {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("expected rendered status to contain %q, got %s", fragment, rendered)
		}
	}
}
