package daemon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabCloseDuplicatesEndpointReturnsClosedTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		duplicateClose: tabs.CloseDuplicatesResult{
			Returned: 1,
			Groups: []tabs.DuplicateCloseGroup{
				{
					URL:             "https://openai.com/docs",
					KeptTargetID:    "tab-1",
					ClosedTargetIDs: []string{"tab-2"},
					Returned:        1,
				},
			},
			ClosedTargets: []string{"tab-2"},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/close-duplicates", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"returned":1`, `"url":"https://openai.com/docs"`, `"closed_targets":["tab-2"]`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabCloseDuplicatesEndpointSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		duplicateCloseErr: errStringDaemon("unexpected status 500"),
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/close-duplicates", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "unexpected status 500") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
