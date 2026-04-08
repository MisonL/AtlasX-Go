package daemon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/defaultbrowser"
)

func TestDefaultBrowserEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	previous := readDaemonDefaultBrowserStatus
	readDaemonDefaultBrowserStatus = func() (defaultbrowser.Status, error) {
		return defaultbrowser.Status{
			Source:        "launchservices",
			HTTPBundleID:  "org.mozilla.firefox",
			HTTPRole:      "all",
			HTTPKnown:     true,
			HTTPSBundleID: "org.mozilla.firefox",
			HTTPSRole:     "all",
			HTTPSKnown:    true,
			Consistent:    true,
		}, nil
	}
	t.Cleanup(func() {
		readDaemonDefaultBrowserStatus = previous
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/default-browser", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"source",
		"http_bundle_id",
		"http_role",
		"http_known",
		"https_bundle_id",
		"https_role",
		"https_known",
		"consistent",
	)
}

func TestDefaultBrowserSetEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	previous := setDaemonDefaultBrowserBundleID
	setDaemonDefaultBrowserBundleID = func(bundleID string) (defaultbrowser.Status, error) {
		return defaultbrowser.Status{
			Source:        "launchservices",
			HTTPBundleID:  bundleID,
			HTTPRole:      "all",
			HTTPKnown:     true,
			HTTPSBundleID: bundleID,
			HTTPSRole:     "all",
			HTTPSKnown:    true,
			Consistent:    true,
		}, nil
	}
	t.Cleanup(func() {
		setDaemonDefaultBrowserBundleID = previous
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/default-browser/set", strings.NewReader(`{"bundle_id":"com.openai.atlasx"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"source",
		"http_bundle_id",
		"http_role",
		"http_known",
		"https_bundle_id",
		"https_role",
		"https_known",
		"consistent",
	)
}
