package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestMoveToNewWindowMovesSingleTargetIntoNewWindow(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	closeCalls := make([]string, 0, 1)
	listCalls := 0

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		listCalls++
		if listCalls == 1 {
			_, _ = w.Write([]byte(`[
				{"id":"src-1","type":"page","title":"One","url":"https://openai.com/one"},
				{"id":"dst-1","type":"page","title":"Dst","url":"https://chatgpt.com"}
			]`))
			return
		}
		_, _ = w.Write([]byte(`[
			{"id":"src-1","type":"page","title":"One","url":"https://openai.com/one"},
			{"id":"dst-1","type":"page","title":"Dst","url":"https://chatgpt.com"},
			{"id":"new-1","type":"page","title":"One","url":"https://openai.com/one"}
		]`))
	})
	mux.HandleFunc("/json/close/src-1", func(w http.ResponseWriter, r *http.Request) {
		closeCalls = append(closeCalls, "src-1")
		w.WriteHeader(http.StatusOK)
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
		case "Browser.getWindowForTarget":
			windowID := 9
			if request.Params["targetId"] == "dst-1" {
				windowID = 7
			}
			if err := connection.WriteJSON(cdpCommandResponse{
				ID: request.ID,
				Result: mustMarshalJSON(t, map[string]any{
					"windowId": windowID,
					"bounds": map[string]any{
						"left":        20,
						"top":         30,
						"width":       1440,
						"height":      900,
						"windowState": "normal",
					},
				}),
			}); err != nil {
				t.Fatalf("write response failed: %v", err)
			}
		case "Target.createTarget":
			if request.Params["newWindow"] != true {
				t.Fatalf("expected newWindow=true, got %+v", request.Params)
			}
			if err := connection.WriteJSON(cdpCommandResponse{
				ID: request.ID,
				Result: mustMarshalJSON(t, map[string]any{
					"targetId": "new-1",
				}),
			}); err != nil {
				t.Fatalf("write response failed: %v", err)
			}
		default:
			t.Fatalf("unexpected method: %s", request.Method)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	result, err := client.MoveToNewWindow("src-1")
	if err != nil {
		t.Fatalf("move to new window failed: %v", err)
	}
	if result.SourceWindowID != 9 || result.SourceTargetID != "src-1" || result.Target.ID != "new-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(closeCalls) != 1 || closeCalls[0] != "src-1" {
		t.Fatalf("unexpected close calls: %+v", closeCalls)
	}
}

func TestMoveToNewWindowRejectsUnknownPageTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json/list":
			_, _ = w.Write([]byte(`[]`))
		case "/json/version":
			_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"ws://127.0.0.1/devtools/browser/root"}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.MoveToNewWindow("src-1"); err == nil {
		t.Fatal("expected move to new window to fail")
	} else if !strings.Contains(err.Error(), "page target src-1 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMoveToNewWindowSurfacesOpenFailure(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"src-1","type":"page","title":"One","url":"https://openai.com/one"}]`))
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
		case "Browser.getWindowForTarget":
			if err := connection.WriteJSON(cdpCommandResponse{
				ID: request.ID,
				Result: mustMarshalJSON(t, map[string]any{
					"windowId": 9,
					"bounds": map[string]any{
						"left":        20,
						"top":         30,
						"width":       1440,
						"height":      900,
						"windowState": "normal",
					},
				}),
			}); err != nil {
				t.Fatalf("write response failed: %v", err)
			}
		case "Target.createTarget":
			if err := connection.WriteJSON(cdpCommandResponse{
				ID: request.ID,
				Error: &cdpError{
					Code:    -32000,
					Message: "createTarget failed",
				},
			}); err != nil {
				t.Fatalf("write response failed: %v", err)
			}
		default:
			t.Fatalf("unexpected method: %s", request.Method)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.MoveToNewWindow("src-1"); err == nil {
		t.Fatal("expected move to new window to fail")
	} else if !strings.Contains(err.Error(), "createTarget failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
