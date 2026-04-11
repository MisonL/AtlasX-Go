package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestSetWindowStateReturnsUpdatedBounds(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/devtools/browser/root", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		defer func() {
			_ = connection.Close()
		}()

		var request cdpCommandRequest
		if err := connection.ReadJSON(&request); err != nil {
			t.Fatalf("read request failed: %v", err)
		}

		switch request.Method {
		case "Browser.setWindowBounds":
			if request.Params["windowId"] != float64(7) && request.Params["windowId"] != 7 {
				t.Fatalf("unexpected windowId: %+v", request.Params)
			}
			bounds, _ := request.Params["bounds"].(map[string]any)
			if bounds["windowState"] != "maximized" {
				t.Fatalf("unexpected bounds: %+v", bounds)
			}
			if err := connection.WriteJSON(cdpCommandResponse{
				ID:     request.ID,
				Result: mustMarshalJSON(t, map[string]any{}),
			}); err != nil {
				t.Fatalf("write response failed: %v", err)
			}
		case "Browser.getWindowBounds":
			if request.Params["windowId"] != float64(7) && request.Params["windowId"] != 7 {
				t.Fatalf("unexpected windowId: %+v", request.Params)
			}
			if err := connection.WriteJSON(cdpCommandResponse{
				ID: request.ID,
				Result: mustMarshalJSON(t, map[string]any{
					"bounds": map[string]any{
						"left":        20,
						"top":         30,
						"width":       1440,
						"height":      900,
						"windowState": "maximized",
					},
				}),
			}); err != nil {
				t.Fatalf("write response failed: %v", err)
			}
		default:
			t.Fatalf("unexpected method: %s", request.Method)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	result, err := client.SetWindowState(7, "maximized")
	if err != nil {
		t.Fatalf("set window state failed: %v", err)
	}
	if result.WindowID != 7 || result.State != "maximized" || result.Width != 1440 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestSetWindowStateRejectsInvalidState(t *testing.T) {
	client := Client{}
	if _, err := client.SetWindowState(7, "unknown"); err == nil {
		t.Fatal("expected invalid state failure")
	} else if !strings.Contains(err.Error(), `unknown window state "unknown"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetWindowStateRejectsNonPositiveWindowID(t *testing.T) {
	client := Client{}
	if _, err := client.SetWindowState(0, "normal"); err == nil {
		t.Fatal("expected invalid window id failure")
	} else if !strings.Contains(err.Error(), "window id must be positive") {
		t.Fatalf("unexpected error: %v", err)
	}
}
