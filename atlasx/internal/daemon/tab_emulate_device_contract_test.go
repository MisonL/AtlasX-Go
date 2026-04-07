package daemon

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabEmulateDeviceEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		deviceResult: tabs.DeviceEmulationResult{
			ID:       "tab-1",
			Title:    "Atlas",
			URL:      "https://chatgpt.com/atlas",
			Preset:   "desktop-wide",
			Viewport: "1440x900@1",
			Mobile:   false,
			Touch:    false,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/emulate-device", bytes.NewBufferString(`{"id":"tab-1","preset":"desktop-wide"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload,
		"id",
		"title",
		"url",
		"preset",
		"viewport",
		"mobile",
		"touch",
	)
}
