package sidebar

import (
	"fmt"
	"strings"
)

func (s Status) Render() string {
	lines := []string{
		"AtlasX Sidebar",
		fmt.Sprintf("configured=%t", s.Configured),
		fmt.Sprintf("ready=%t", s.Ready),
		fmt.Sprintf("default_provider=%s", s.DefaultProvider),
		fmt.Sprintf("provider=%s", s.Provider),
		fmt.Sprintf("model=%s", s.Model),
		fmt.Sprintf("base_url=%s", s.BaseURL),
		fmt.Sprintf("api_key_env=%s", s.APIKeyEnv),
		fmt.Sprintf("provider_count=%d", len(s.Providers)),
		fmt.Sprintf("timeout_ms=%d", s.TimeoutMS),
		fmt.Sprintf("retry_attempts=%d", s.RetryAttempts),
		fmt.Sprintf("token_budget=%d", s.TokenBudget),
		fmt.Sprintf("last_trace_id=%s", s.LastTraceID),
		fmt.Sprintf("last_error=%s", s.LastError),
		fmt.Sprintf("last_error_at=%s", s.LastErrorAt),
		fmt.Sprintf("reason=%s", s.Reason),
	}
	for index, provider := range s.Providers {
		lines = append(lines,
			fmt.Sprintf("provider[%d].id=%s", index, provider.ID),
			fmt.Sprintf("provider[%d].provider=%s", index, provider.Provider),
			fmt.Sprintf("provider[%d].model=%s", index, provider.Model),
			fmt.Sprintf("provider[%d].base_url=%s", index, provider.BaseURL),
			fmt.Sprintf("provider[%d].api_key_env=%s", index, provider.APIKeyEnv),
		)
	}
	return strings.Join(lines, "\n") + "\n"
}
