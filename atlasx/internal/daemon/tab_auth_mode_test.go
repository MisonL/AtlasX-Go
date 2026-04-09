package daemon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabAuthModeReturnsStructuredView(t *testing.T) {
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
	body := recorder.Body.String()
	for _, fragment := range []string{
		`"mode":"logged_in"`,
		`"inferred":true`,
		`"reason":"workspace_signals_observed"`,
		`"host":"chatgpt.com"`,
		`"path":"/c/abc123"`,
	} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("expected body to contain %q, got %s", fragment, body)
		}
	}
}

func TestTabAuthModeReturnsCaptureError(t *testing.T) {
	context := tabs.PageContext{
		ID:           "tab-1",
		Title:        "ChatGPT",
		URL:          "https://chatgpt.com/auth/login",
		CapturedAt:   "2026-04-09T10:00:00Z",
		TextLimit:    4096,
		CaptureError: "cdp error -32000: capture failed",
	}
	restoreDaemonHooks(t, &stubTabsClient{
		authModeErr: &tabs.CaptureError{
			Context: context,
			Cause:   errStringDaemon("cdp error -32000: capture failed"),
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/auth-mode?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"capture_error":"cdp error -32000: capture failed"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
