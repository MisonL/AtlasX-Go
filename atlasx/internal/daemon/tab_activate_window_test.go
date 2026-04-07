package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabActivateWindowEndpointReturnsActivatedTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowActivate: tabs.WindowActivateResult{
			WindowID:          7,
			ActivatedTargetID: "tab-1",
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/activate-window", bytes.NewBufferString(`{"window_id":7}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"window_id":7`, `"activated_target_id":"tab-1"`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabActivateWindowEndpointRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/activate-window", bytes.NewBufferString(`{"window_id":0}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestTabActivateWindowEndpointSurfacesMissingWindow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowActivateErr: errStringDaemon("window 7 not found"),
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/activate-window", bytes.NewBufferString(`{"window_id":7}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "window 7 not found") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
