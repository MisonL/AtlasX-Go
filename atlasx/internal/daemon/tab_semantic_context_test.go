package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabSemanticContextEndpointReturnsStructuredContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		semanticContext: tabs.SemanticContext{
			ID:               "tab-1",
			Title:            "Atlas",
			URL:              "https://chatgpt.com/atlas",
			CapturedAt:       "2026-04-07T13:00:00Z",
			Returned:         3,
			HeadingsReturned: 1,
			LinksReturned:    1,
			FormsReturned:    1,
			Headings:         []tabs.SemanticHeading{{Level: 1, Text: "Atlas"}},
			Links:            []tabs.SemanticLink{{Text: "OpenAI Docs", URL: "https://platform.openai.com/docs"}},
			Forms:            []tabs.SemanticForm{{Action: "https://chatgpt.com/search", Method: "GET", InputCount: 2}},
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/semantic-context?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range [][]byte{
		[]byte(`"returned":3`),
		[]byte(`"headings_returned":1`),
		[]byte(`"links_returned":1`),
		[]byte(`"forms_returned":1`),
		[]byte(`"text":"Atlas"`),
	} {
		if !bytes.Contains(recorder.Body.Bytes(), fragment) {
			t.Fatalf("unexpected response body: %s", recorder.Body.String())
		}
	}
}

func TestTabSemanticContextEndpointReturnsStructuredFailure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	context := tabs.SemanticContext{
		ID:           "tab-1",
		Title:        "Atlas",
		URL:          "https://chatgpt.com/atlas",
		CapturedAt:   "2026-04-07T13:00:00Z",
		CaptureError: "cdp error -32000: semantic capture failed",
	}
	restoreDaemonHooks(t, &stubTabsClient{
		semanticContext: context,
		semanticErr: &tabs.SemanticCaptureError{
			Context: context,
			Cause:   errStringDaemon("cdp error -32000: semantic capture failed"),
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/semantic-context?id=tab-1", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"capture_error":"cdp error -32000: semantic capture failed"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}
