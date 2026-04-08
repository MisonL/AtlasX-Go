package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestMergeWindowMovesTargetsIntoTargetWindow(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	closeCalls := make([]string, 0, 2)
	openCalls := 0

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[
			{"id":"src-1","type":"page","title":"One","url":"https://openai.com/one"},
			{"id":"src-2","type":"page","title":"Two","url":"https://openai.com/two"},
			{"id":"dst-1","type":"page","title":"Dst","url":"https://chatgpt.com"}
		]`))
	})
	mux.HandleFunc("/json/activate/dst-1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/json/new", func(w http.ResponseWriter, r *http.Request) {
		openCalls++
		switch openCalls {
		case 1:
			_, _ = w.Write([]byte(`{"id":"new-1","type":"page","title":"One","url":"https://openai.com/one"}`))
		case 2:
			_, _ = w.Write([]byte(`{"id":"new-2","type":"page","title":"Two","url":"https://openai.com/two"}`))
		default:
			t.Fatalf("unexpected open call count: %d", openCalls)
		}
	})
	mux.HandleFunc("/json/close/src-1", func(w http.ResponseWriter, r *http.Request) {
		closeCalls = append(closeCalls, "src-1")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/json/close/src-2", func(w http.ResponseWriter, r *http.Request) {
		closeCalls = append(closeCalls, "src-2")
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
	result, err := client.MergeWindow(9, 7)
	if err != nil {
		t.Fatalf("merge window failed: %v", err)
	}
	if result.SourceWindowID != 9 || result.TargetWindowID != 7 || result.Returned != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.MovedTargets) != 2 || result.MovedTargets[0].SourceTargetID != "src-1" || result.MovedTargets[1].SourceTargetID != "src-2" {
		t.Fatalf("unexpected moved targets: %+v", result.MovedTargets)
	}
	if len(closeCalls) != 2 || closeCalls[0] != "src-1" || closeCalls[1] != "src-2" {
		t.Fatalf("unexpected close calls: %+v", closeCalls)
	}
}

func TestMergeWindowRejectsSameWindowID(t *testing.T) {
	client := Client{}
	if _, err := client.MergeWindow(7, 7); err == nil {
		t.Fatal("expected merge window to fail")
	} else if !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMergeWindowRejectsUnknownSourceWindow(t *testing.T) {
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
	if _, err := client.MergeWindow(9, 7); err == nil {
		t.Fatal("expected merge window to fail")
	} else if !strings.Contains(err.Error(), "window 9 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
