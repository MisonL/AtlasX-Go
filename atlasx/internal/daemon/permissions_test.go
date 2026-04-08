package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPermissionsEndpointReturnsBoundaryFacts(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/v1/permissions", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"granted_state_observable":false`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
	if !bytes.Contains(recorder.Body.Bytes(), []byte(`"permission_write_supported":false`)) {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}

func TestPermissionsEndpointRejectsWrongMethod(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/v1/permissions", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
}
