package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabAuthModeEndpointContract(t *testing.T) {
	restoreDaemonHooks(t, &stubTabsClient{
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

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/auth-mode?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"id",
		"title",
		"url",
		"captured_at",
		"host",
		"path",
		"mode",
		"inferred",
		"reason",
		"login_prompt_present",
		"workspace_signal_present",
		"cookie_count",
		"cookie_names",
		"local_storage_count",
		"local_storage_keys",
		"session_storage_count",
		"session_storage_keys",
	)
}
