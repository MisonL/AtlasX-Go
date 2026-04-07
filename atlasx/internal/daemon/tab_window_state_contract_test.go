package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabWindowStateEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowState: tabs.WindowBounds{
			WindowID: 7,
			State:    "normal",
			Left:     20,
			Top:      30,
			Width:    1440,
			Height:   900,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/window-state", bytes.NewBufferString(`{"window_id":7,"state":"normal"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "window_id", "state", "left", "top", "width", "height")
}
