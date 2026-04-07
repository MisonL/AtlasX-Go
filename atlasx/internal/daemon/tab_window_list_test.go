package daemon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabWindowsEndpointReturnsStructuredWindows(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windows: []tabs.WindowSummary{
			{
				WindowID: 7,
				State:    "normal",
				Left:     20,
				Top:      30,
				Width:    1440,
				Height:   900,
				Returned: 1,
				Targets: []tabs.Target{
					{ID: "tab-1", Type: "page", Title: "Atlas", URL: "https://chatgpt.com/atlas"},
				},
			},
		},
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/windows", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	for _, fragment := range []string{
		`"returned":1`,
		`"window_id":7`,
		`"state":"normal"`,
		`"title":"Atlas"`,
	} {
		if !strings.Contains(recorder.Body.String(), fragment) {
			t.Fatalf("unexpected body: %s", recorder.Body.String())
		}
	}
}

func TestTabWindowsEndpointSurfacesBrowserErrors(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		windowsErr: errStringDaemon("browser websocket debugger url is not available"),
	})

	request := httptest.NewRequest(http.MethodGet, "/v1/tabs/windows", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "browser websocket debugger url is not available") {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}
