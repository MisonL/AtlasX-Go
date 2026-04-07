package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabWindowBoundsEndpointReturnsStructuredBounds(t *testing.T) {
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

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["window_id"].(float64) != 7 || payload["width"].(float64) != 1280 || payload["height"].(float64) != 720 {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestTabWindowBoundsEndpointRejectsInvalidWidth(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/window-bounds", bytes.NewBufferString(`{"window_id":7,"left":10,"top":20,"width":0,"height":720}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestTabWindowBoundsEndpointSurfacesBoundsErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowBoundsErr: errStringDaemon("width must be positive"),
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/window-bounds", bytes.NewBufferString(`{"window_id":7,"left":10,"top":20,"width":1280,"height":720}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}
