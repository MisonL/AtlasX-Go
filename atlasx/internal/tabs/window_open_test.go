package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestOpenInWindowActivatesWindowBeforeOpeningTarget(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	steps := make([]string, 0, 2)

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas"}]`))
	})
	mux.HandleFunc("/json/activate/tab-1", func(w http.ResponseWriter, r *http.Request) {
		steps = append(steps, "activate")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/json/new", func(w http.ResponseWriter, r *http.Request) {
		steps = append(steps, "open")
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		_, _ = w.Write([]byte(`{"id":"tab-2","type":"page","title":"OpenAI","url":"https://openai.com"}`))
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
	result, err := client.OpenInWindow(7, "https://openai.com")
	if err != nil {
		t.Fatalf("open in window failed: %v", err)
	}
	if result.WindowID != 7 || result.ActivatedTargetID != "tab-1" || result.Target.ID != "tab-2" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(steps) != 2 || steps[0] != "activate" || steps[1] != "open" {
		t.Fatalf("unexpected steps: %+v", steps)
	}
}

func TestOpenInWindowRejectsUnknownWindow(t *testing.T) {
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
	if _, err := client.OpenInWindow(7, "https://openai.com"); err == nil {
		t.Fatal("expected missing window failure")
	} else if !strings.Contains(err.Error(), "window 7 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenInWindowSurfacesOpenFailure(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas"}]`))
	})
	mux.HandleFunc("/json/activate/tab-1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/json/new", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "open failed", http.StatusInternalServerError)
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
	if _, err := client.OpenInWindow(7, "https://openai.com"); err == nil {
		t.Fatal("expected open in window to fail")
	} else if !strings.Contains(err.Error(), "unexpected status 500") {
		t.Fatalf("unexpected error: %v", err)
	}
}
