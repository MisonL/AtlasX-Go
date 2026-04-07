package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabWindowBoundsEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowBounds: tabs.WindowBounds{
			WindowID: 7,
			State:    "normal",
			Left:     10,
			Top:      20,
			Width:    1280,
			Height:   720,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/window-bounds", bytes.NewBufferString(`{"window_id":7,"left":10,"top":20,"width":1280,"height":720}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "window_id", "state", "left", "top", "width", "height")
}
