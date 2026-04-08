package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabMoveToNewWindowEndpointReturnsTarget(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowMoveNew: tabs.WindowMoveToNewResult{
			SourceWindowID: 9,
			SourceTargetID: "src-1",
			Target: tabs.Target{
				ID:    "new-1",
				Type:  "page",
				Title: "OpenAI",
				URL:   "https://openai.com",
			},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/move-to-new-window", bytes.NewBufferString(`{"id":"src-1"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"source_window_id":9`, `"source_target_id":"src-1"`, `"id":"new-1"`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabMoveToNewWindowEndpointRejectsMissingTargetID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/move-to-new-window", bytes.NewBufferString(`{}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestTabMoveToNewWindowEndpointSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowMoveNewErr: errStringDaemon("page target src-1 not found"),
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/move-to-new-window", bytes.NewBufferString(`{"id":"src-1"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "page target src-1 not found") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
