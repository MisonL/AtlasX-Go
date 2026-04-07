package daemon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabSearchEndpointReturnsMatches(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		searchTargets: []tabs.Target{
			{ID: "tab-1", Type: "page", Title: "Atlas Docs", URL: "https://openai.com/docs/atlas"},
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/search?q=atlas", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"returned":1`, `"id":"tab-1"`, `"title":"Atlas Docs"`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabSearchEndpointRejectsMissingQuery(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/search", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "missing query parameter q") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestTabSearchEndpointSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		searchErr: errStringDaemon("search failed"),
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/search?q=atlas", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "search failed") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
