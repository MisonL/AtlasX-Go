package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabOpenInWindowEndpointReturnsTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowOpen: tabs.WindowOpenResult{
			WindowID:          7,
			ActivatedTargetID: "tab-1",
			Target: tabs.Target{
				ID:    "tab-2",
				Type:  "page",
				Title: "OpenAI",
				URL:   "https://openai.com",
			},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/open-in-window", bytes.NewBufferString(`{"window_id":7,"url":"https://openai.com"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"window_id":7`, `"activated_target_id":"tab-1"`, `"id":"tab-2"`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabOpenInWindowEndpointRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/open-in-window", bytes.NewBufferString(`{"window_id":0,"url":"https://openai.com"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestTabOpenInWindowEndpointSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowOpenErr: errStringDaemon("window 7 not found"),
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/open-in-window", bytes.NewBufferString(`{"window_id":7,"url":"https://openai.com"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "window 7 not found") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
