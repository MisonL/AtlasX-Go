package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdatesEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	request := httptest.NewRequest(http.MethodGet, "/v1/updates", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"runtime_root",
		"manifest_present",
		"staged_version",
		"staged_channel",
		"staged_bundle_path",
		"staged_binary_path",
		"staged_ready",
		"plan_present",
		"plan_version",
		"plan_channel",
		"plan_bundle_name",
		"plan_source_url",
		"plan_archive_path",
		"plan_phase",
		"plan_last_error",
		"plan_pending",
		"plan_in_flight",
	)
}
