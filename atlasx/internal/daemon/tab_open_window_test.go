package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTabsOpenWindowEndpoint(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	client := &stubTabsClient{}
	restoreDaemonHooks(t, client)

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/open-window", bytes.NewBufferString(`{"url":"https://openai.com"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if client.openedURL != "https://openai.com" {
		t.Fatalf("unexpected opened url: %s", client.openedURL)
	}
	if !strings.Contains(recorder.Body.String(), `"id":"tab-window"`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestTabsOpenWindowEndpointRejectsUnsupportedScheme(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	client := &stubTabsClient{}
	restoreDaemonHooks(t, client)

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/open-window", bytes.NewBufferString(`{"url":"file:///tmp/secret.txt"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if client.openedURL != "" {
		t.Fatalf("unexpected opened url: %s", client.openedURL)
	}
}
