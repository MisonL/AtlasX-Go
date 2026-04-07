package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestSetWindowBoundsReturnsUpdatedBounds(t *testing.T) {
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
		defer connection.Close()

		var request cdpCommandRequest
		if err := connection.ReadJSON(&request); err != nil {
			t.Fatalf("read request failed: %v", err)
		}

		switch request.Method {
		case "Browser.setWindowBounds":
			bounds, _ := request.Params["bounds"].(map[string]any)
			if bounds["left"] != float64(10) || bounds["top"] != float64(20) || bounds["width"] != float64(1280) || bounds["height"] != float64(720) {
				t.Fatalf("unexpected bounds: %+v", bounds)
			}
			if err := connection.WriteJSON(cdpCommandResponse{
				ID:     request.ID,
				Result: mustMarshalJSON(t, map[string]any{}),
			}); err != nil {
				t.Fatalf("write response failed: %v", err)
			}
		case "Browser.getWindowBounds":
			if err := connection.WriteJSON(cdpCommandResponse{
				ID: request.ID,
				Result: mustMarshalJSON(t, map[string]any{
					"bounds": map[string]any{
						"left":        10,
						"top":         20,
						"width":       1280,
						"height":      720,
						"windowState": "normal",
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
	result, err := client.SetWindowBounds(7, 10, 20, 1280, 720)
	if err != nil {
		t.Fatalf("set window bounds failed: %v", err)
	}
	if result.WindowID != 7 || result.Left != 10 || result.Width != 1280 || result.Height != 720 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestSetWindowBoundsRejectsNonPositiveWidth(t *testing.T) {
	client := Client{}
	if _, err := client.SetWindowBounds(7, 10, 20, 0, 720); err == nil {
		t.Fatal("expected invalid width failure")
	} else if !strings.Contains(err.Error(), "width must be positive") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetWindowBoundsRejectsNonPositiveHeight(t *testing.T) {
	client := Client{}
	if _, err := client.SetWindowBounds(7, 10, 20, 1280, 0); err == nil {
		t.Fatal("expected invalid height failure")
	} else if !strings.Contains(err.Error(), "height must be positive") {
		t.Fatalf("unexpected error: %v", err)
	}
}
