package daemon

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabEmulateDeviceEndpointReturnsStructuredResult(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		deviceResult: tabs.DeviceEmulationResult{
			ID:       "tab-1",
			Title:    "Atlas",
			URL:      "https://chatgpt.com/atlas",
			Preset:   "iphone-13",
			Viewport: "390x844@3",
			Mobile:   true,
			Touch:    true,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/emulate-device", bytes.NewBufferString(`{"id":"tab-1","preset":"iphone-13"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{
		`"preset":"iphone-13"`,
		`"viewport":"390x844@3"`,
		`"mobile":true`,
		`"touch":true`,
	} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabEmulateDeviceEndpointSupportsOffPreset(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		deviceResult: tabs.DeviceEmulationResult{
			ID:       "tab-1",
			Title:    "Atlas",
			URL:      "https://chatgpt.com/atlas",
			Preset:   "off",
			Viewport: "off",
			Mobile:   false,
			Touch:    false,
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/emulate-device", bytes.NewBufferString(`{"id":"tab-1","preset":"off"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"preset":"off"`) || !strings.Contains(recorder.Body.String(), `"mobile":false`) {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

func TestTabEmulateDeviceEndpointSurfacesPresetError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		deviceErr: errStringDaemon(`unknown device preset "unknown"`),
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/emulate-device", bytes.NewBufferString(`{"id":"tab-1","preset":"unknown"}`))
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["error"] != `unknown device preset "unknown"` {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
