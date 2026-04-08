package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoctorEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/doctor", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"Paths",
		"Config",
		"Chrome",
		"ChromeStatus",
		"RuntimeManifest",
		"IsolatedPath",
		"SharedModeName",
		"Session",
	)
}
