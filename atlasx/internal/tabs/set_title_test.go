package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestSetTitleUpdatesPageTitle(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[
			{"id":"tab-1","type":"page","title":"Before","url":"https://openai.com/work","webSocketDebuggerUrl":"ws` + strings.TrimPrefix(server.URL, "http") + `/devtools/page/tab-1"}
		]`))
	})
	mux.HandleFunc("/devtools/page/tab-1", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		defer connection.Close()

		var request cdpCommandRequest
		if err := connection.ReadJSON(&request); err != nil {
			t.Fatalf("read request failed: %v", err)
		}
		if request.Method != "Runtime.evaluate" {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		expression, _ := request.Params["expression"].(string)
		if !strings.Contains(expression, "Atlas Workbench") {
			t.Fatalf("unexpected expression: %s", expression)
		}
		if err := connection.WriteJSON(cdpCommandResponse{
			ID: request.ID,
			Result: mustMarshalJSON(t, map[string]any{
				"result": map[string]any{
					"type": "object",
					"value": map[string]any{
						"title": "Atlas Workbench",
						"url":   "https://openai.com/work",
					},
				},
			}),
		}); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	result, err := client.SetTitle("tab-1", "Atlas Workbench")
	if err != nil {
		t.Fatalf("set title failed: %v", err)
	}
	if result.ID != "tab-1" || result.Title != "Atlas Workbench" || result.URL != "https://openai.com/work" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestSetTitleRejectsBlankTitle(t *testing.T) {
	client := Client{}
	if _, err := client.SetTitle("tab-1", " "); err == nil {
		t.Fatal("expected set title to fail")
	} else if !strings.Contains(err.Error(), "title is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetTitleRejectsUnknownTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.SetTitle("tab-1", "Atlas"); err == nil {
		t.Fatal("expected set title to fail")
	} else if !strings.Contains(err.Error(), "target tab-1 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetTitleRejectsTargetWithoutWebSocket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[
			{"id":"tab-1","type":"page","title":"Before","url":"https://openai.com/work"}
		]`))
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.SetTitle("tab-1", "Atlas"); err == nil {
		t.Fatal("expected set title to fail")
	} else if !strings.Contains(err.Error(), "target does not expose a websocket debugger url") {
		t.Fatalf("unexpected error: %v", err)
	}
}
