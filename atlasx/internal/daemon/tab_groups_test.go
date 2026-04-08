package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabGroupsReturnsStructuredInferredGroups(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 11,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas A", URL: "https://chatgpt.com/atlas/a"},
					{ID: "tab-2", Type: "page", Title: "Atlas B", URL: "https://chatgpt.com/atlas/b"},
				},
			},
			{
				WindowID: 22,
				Targets: []tabs.Target{
					{ID: "tab-3", Type: "page", Title: "Atlas C", URL: "https://chatgpt.com/atlas/c"},
					{ID: "tab-4", Type: "page", Title: "Elsewhere", URL: "https://example.com/other"},
				},
			},
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/groups", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	if payload["inferred"] != true || payload["returned"].(float64) != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	groups, ok := payload["groups"].([]any)
	if !ok || len(groups) != 1 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}
