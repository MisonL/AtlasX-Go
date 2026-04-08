package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabMergeWindowEndpointReturnsMovedTargets(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowMerge: tabs.WindowMergeResult{
			SourceWindowID: 9,
			TargetWindowID: 7,
			Returned:       1,
			MovedTargets: []tabs.WindowMergeTarget{
				{
					SourceTargetID:    "src-1",
					ActivatedTargetID: "dst-1",
					Target: tabs.Target{
						ID:    "new-1",
						Type:  "page",
						Title: "OpenAI",
						URL:   "https://openai.com",
					},
				},
			},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/merge-window", bytes.NewBufferString(`{"source_window_id":9,"target_window_id":7}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{`"source_window_id":9`, `"target_window_id":7`, `"source_target_id":"src-1"`} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabMergeWindowEndpointRejectsInvalidWindowID(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/merge-window", bytes.NewBufferString(`{"source_window_id":0,"target_window_id":7}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestTabMergeWindowEndpointSurfacesFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowMergeErr: errStringDaemon("window 9 not found"),
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/merge-window", bytes.NewBufferString(`{"source_window_id":9,"target_window_id":7}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "window 9 not found") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
