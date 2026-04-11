package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWindowsGroupsPageTargetsByWindow(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[
			{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas"},
			{"id":"tab-2","type":"page","title":"OpenAI","url":"https://openai.com"},
			{"id":"worker-1","type":"worker","title":"Worker","url":"https://example.com/worker"}
		]`))
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
		if request.Method != "Browser.getWindowForTarget" {
			t.Fatalf("unexpected method: %s", request.Method)
		}

		targetID, _ := request.Params["targetId"].(string)
		var payload map[string]any
		switch targetID {
		case "tab-1":
			payload = map[string]any{
				"windowId": 7,
				"bounds": map[string]any{
					"left":        20,
					"top":         30,
					"width":       1440,
					"height":      900,
					"windowState": "normal",
				},
			}
		case "tab-2":
			payload = map[string]any{
				"windowId": 9,
				"bounds": map[string]any{
					"left":        100,
					"top":         120,
					"width":       1280,
					"height":      820,
					"windowState": "maximized",
				},
			}
		default:
			t.Fatalf("unexpected targetId: %s", targetID)
		}

		if err := connection.WriteJSON(cdpCommandResponse{
			ID:     request.ID,
			Result: mustMarshalJSON(t, payload),
		}); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	windows, err := client.Windows()
	if err != nil {
		t.Fatalf("windows failed: %v", err)
	}
	if len(windows) != 2 {
		t.Fatalf("unexpected window count: %d", len(windows))
	}
	if windows[0].WindowID != 7 || windows[0].Returned != 1 || windows[0].Targets[0].ID != "tab-1" {
		t.Fatalf("unexpected first window: %+v", windows[0])
	}
	if windows[1].WindowID != 9 || windows[1].State != "maximized" || windows[1].Targets[0].ID != "tab-2" {
		t.Fatalf("unexpected second window: %+v", windows[1])
	}
}

func TestWindowsReturnsEmptyWithoutPageTargets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"worker-1","type":"worker","title":"Worker","url":"https://example.com/worker"}]`))
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	windows, err := client.Windows()
	if err != nil {
		t.Fatalf("windows failed: %v", err)
	}
	if len(windows) != 0 {
		t.Fatalf("expected no windows, got %+v", windows)
	}
}

func TestWindowsFailsWithoutBrowserWebSocket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json/list":
			_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas"}]`))
		case "/json/version":
			_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":""}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.Windows(); err == nil {
		t.Fatal("expected windows to fail without browser websocket")
	} else if !strings.Contains(err.Error(), "browser websocket debugger url is not available") {
		t.Fatalf("unexpected error: %v", err)
	}
}
