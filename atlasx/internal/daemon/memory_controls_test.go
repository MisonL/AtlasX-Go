package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMemoryControlsEndpointReturnsControls(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/memory/controls", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"persist_enabled":true`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"page_visibility_enabled":true`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"hidden_hosts":[]`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestMemoryControlsEndpointUpdatesControls(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodPost, "/v1/memory/controls", bytes.NewBufferString(`{"persist_enabled":false}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"persist_enabled":false`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"page_visibility_enabled":true`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestMemoryControlsEndpointUpdatesPageVisibility(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodPost, "/v1/memory/controls", bytes.NewBufferString(`{"page_visibility_enabled":false}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"page_visibility_enabled":false`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestMemoryControlsEndpointUpdatesSiteVisibility(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodPost, "/v1/memory/controls", bytes.NewBufferString(`{"site_host":"https://ChatGPT.com/atlas","site_visibility_enabled":false}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"hidden_hosts":["chatgpt.com"]`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestMemoryControlsEndpointRejectsSiteVisibilityWithoutHost(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodPost, "/v1/memory/controls", bytes.NewBufferString(`{"site_visibility_enabled":false}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"error":"site_host is required when site_visibility_enabled is provided"`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestMemoryControlsEndpointRejectsMissingPersistFlag(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodPost, "/v1/memory/controls", bytes.NewBufferString(`{}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}
