package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabCloseWindowEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowClose: tabs.WindowCloseResult{
			WindowID:      7,
			Returned:      1,
			ClosedTargets: []string{"tab-1"},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/close-window", bytes.NewBufferString(`{"window_id":7}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "window_id", "returned", "closed_targets")
}
