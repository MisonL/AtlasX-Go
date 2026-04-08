package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestMoveToWindowMovesSingleTargetIntoTargetWindow(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	closeCalls := make([]string, 0, 1)

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[
			{"id":"src-1","type":"page","title":"One","url":"https://openai.com/one"},
			{"id":"dst-1","type":"page","title":"Dst","url":"https://chatgpt.com"}
		]`))
	})
	mux.HandleFunc("/json/activate/dst-1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/json/new", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":"new-1","type":"page","title":"One","url":"https://openai.com/one"}`))
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
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	result, err := client.MoveToWindow("src-1", 7)
	if err != nil {
		t.Fatalf("move to window failed: %v", err)
	}
	if result.SourceWindowID != 9 || result.TargetWindowID != 7 || result.SourceTargetID != "src-1" || result.Target.ID != "new-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(closeCalls) != 1 || closeCalls[0] != "src-1" {
		t.Fatalf("unexpected close calls: %+v", closeCalls)
	}
}

func TestMoveToWindowRejectsUnknownPageTarget(t *testing.T) {
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
	if _, err := client.MoveToWindow("src-1", 7); err == nil {
		t.Fatal("expected move to window to fail")
	} else if !strings.Contains(err.Error(), "page target src-1 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMoveToWindowRejectsSameWindowID(t *testing.T) {
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
		if err := connection.WriteJSON(cdpCommandResponse{
			ID: request.ID,
			Result: mustMarshalJSON(t, map[string]any{
				"windowId": 7,
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
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.MoveToWindow("src-1", 7); err == nil {
		t.Fatal("expected move to window to fail")
	} else if !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("unexpected error: %v", err)
	}
}
