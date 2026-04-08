package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabsSetTitleEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		titleUpdate: tabs.TitleUpdateResult{
			ID:    "tab-1",
			Title: "Atlas Workbench",
			URL:   "https://openai.com/work",
		},
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/tabs/set-title",
		bytes.NewBufferString(`{"id":"tab-1","title":"Atlas Workbench"}`),
	)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"id":"tab-1"`, `"title":"Atlas Workbench"`, `"url":"https://openai.com/work"`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("expected body to contain %q, got %s", fragment, recorder.Body.String())
		}
	}
}

func TestTabsSetTitleEndpointRejectsBlankTitle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(
		http.MethodPost,
		"/v1/tabs/set-title",
		bytes.NewBufferString(`{"id":"tab-1","title":" "}`),
	)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"error":"title is required"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
