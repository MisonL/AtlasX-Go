package daemon

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabSelectionEndpointReturnsStructuredSelection(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		selection: tabs.SelectionContext{
			ID:                     "tab-1",
			Title:                  "Atlas",
			URL:                    "https://chatgpt.com/atlas",
			SelectionText:          "Atlas selected text",
			CapturedAt:             "2026-04-07T12:00:00Z",
			SelectionPresent:       true,
			SelectionTextTruncated: false,
			SelectionTextLength:    20,
			SelectionTextLimit:     1024,
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/selection?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"selection_text":"Atlas selected text"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"selection_present":true`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestTabSelectionEndpointReturnsStructuredFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		selectionErr: &tabs.SelectionCaptureError{
			Context: tabs.SelectionContext{
				ID:                 "tab-1",
				Title:              "Atlas",
				URL:                "https://chatgpt.com/atlas",
				CapturedAt:         "2026-04-07T12:00:00Z",
				SelectionTextLimit: 1024,
				CaptureError:       "cdp error -32000: selection failed",
			},
			Cause: errors.New("cdp error -32000: selection failed"),
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/selection?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"capture_error":"cdp error -32000: selection failed"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}
