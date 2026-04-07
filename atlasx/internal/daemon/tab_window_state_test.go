package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabWindowStateEndpointReturnsStructuredBounds(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowState: tabs.WindowBounds{
			WindowID: 7,
			State:    "maximized",
			Left:     20,
			Top:      30,
			Width:    1440,
			Height:   900,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/window-state", bytes.NewBufferString(`{"window_id":7,"state":"maximized"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"window_id":7`, `"state":"maximized"`, `"width":1440`, `"height":900`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabWindowStateEndpointRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/window-state", bytes.NewBufferString(`{"window_id":0,"state":"normal"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "window_id must be positive") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestTabWindowStateEndpointSurfacesStateErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowStateErr: errStringDaemon(`unknown window state "unknown"`),
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/window-state", bytes.NewBufferString(`{"window_id":7,"state":"unknown"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["error"] != `unknown window state "unknown"` {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
