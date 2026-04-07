package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabCloseWindowEndpointReturnsClosedTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowClose: tabs.WindowCloseResult{
			WindowID:      7,
			Returned:      2,
			ClosedTargets: []string{"tab-1", "tab-2"},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/close-window", bytes.NewBufferString(`{"window_id":7}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"window_id":7`, `"returned":2`, `"closed_targets":["tab-1","tab-2"]`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabCloseWindowEndpointRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/close-window", bytes.NewBufferString(`{"window_id":0}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestTabCloseWindowEndpointSurfacesMissingWindow(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowCloseErr: errStringDaemon("window 7 not found"),
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/close-window", bytes.NewBufferString(`{"window_id":7}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "window 7 not found") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
