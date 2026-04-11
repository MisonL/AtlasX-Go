package tabs

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestOpenDevToolsWindowIntoWindowOpensSourceWindowTargetsInTargetWindow(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	activateCalls := make([]string, 0, 2)
	openURLs := make([]string, 0, 2)

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[
			{"id":"src-1","type":"page","title":"Source One","url":"https://openai.com/one","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1"},
			{"id":"src-2","type":"page","title":"Source Two","url":"https://openai.com/two","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-2"},
			{"id":"dst-1","type":"page","title":"Target","url":"https://chatgpt.com/workspace","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fdst-1"}
		]`))
	})
	mux.HandleFunc("/json/activate/dst-1", func(w http.ResponseWriter, r *http.Request) {
		activateCalls = append(activateCalls, "dst-1")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/json/new", func(w http.ResponseWriter, r *http.Request) {
		decodedURL, err := url.QueryUnescape(r.URL.RawQuery)
		if err != nil {
			t.Fatalf("decode open url failed: %v", err)
		}
		openURLs = append(openURLs, decodedURL)
		switch len(openURLs) {
		case 1:
			_, _ = w.Write([]byte(`{"id":"devtools-open-1","type":"page","title":"DevTools 1","url":"` + decodedURL + `"}`))
		case 2:
			_, _ = w.Write([]byte(`{"id":"devtools-open-2","type":"page","title":"DevTools 2","url":"` + decodedURL + `"}`))
		default:
			t.Fatalf("unexpected open call count: %d", len(openURLs))
		}
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

		targetID, _ := request.Params["targetId"].(string)
		windowID := 9
		if targetID == "dst-1" {
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
	result, err := client.OpenDevToolsWindowIntoWindow(9, 7)
	if err != nil {
		t.Fatalf("open devtools window into window failed: %v", err)
	}
	if result.SourceWindowID != 9 || result.TargetWindowID != 7 || result.Returned != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.OpenedTargets) != 2 || result.OpenedTargets[0].SourceTargetID != "src-1" || result.OpenedTargets[1].SourceTargetID != "src-2" {
		t.Fatalf("unexpected opened targets: %+v", result.OpenedTargets)
	}
	if len(activateCalls) != 2 {
		t.Fatalf("unexpected activate call count: %d", len(activateCalls))
	}
	if len(openURLs) != 2 {
		t.Fatalf("unexpected open url count: %d", len(openURLs))
	}
	if !strings.Contains(openURLs[0], "src-1") || !strings.Contains(openURLs[1], "src-2") {
		t.Fatalf("unexpected open urls: %+v", openURLs)
	}
}

func TestOpenDevToolsWindowIntoWindowRejectsSameWindowID(t *testing.T) {
	client := Client{}
	if _, err := client.OpenDevToolsWindowIntoWindow(7, 7); err == nil {
		t.Fatal("expected open devtools window into window to fail")
	} else if !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenDevToolsWindowIntoWindowRejectsUnknownSourceWindow(t *testing.T) {
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
	if _, err := client.OpenDevToolsWindowIntoWindow(9, 7); err == nil {
		t.Fatal("expected open devtools window into window to fail")
	} else if !strings.Contains(err.Error(), "window 9 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenDevToolsWindowIntoWindowRejectsUnknownTargetWindow(t *testing.T) {
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
			{"id":"src-1","type":"page","title":"Source One","url":"https://openai.com/one","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1"}
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
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.OpenDevToolsWindowIntoWindow(9, 7); err == nil {
		t.Fatal("expected open devtools window into window to fail")
	} else if !strings.Contains(err.Error(), "window 7 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenDevToolsWindowIntoWindowReturnsPartialResultOnOpenFailure(t *testing.T) {
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
			{"id":"src-1","type":"page","title":"Source One","url":"https://openai.com/one","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1"},
			{"id":"src-2","type":"page","title":"Source Two","url":"https://openai.com/two","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-2"},
			{"id":"dst-1","type":"page","title":"Target","url":"https://chatgpt.com/workspace","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fdst-1"}
		]`))
	})
	mux.HandleFunc("/json/activate/dst-1", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/json/new", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "src-1") {
			_, _ = w.Write([]byte(`{"id":"devtools-open-1","type":"page","title":"DevTools 1","url":"https://example.com/devtools/src-1"}`))
			return
		}
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

		targetID, _ := request.Params["targetId"].(string)
		windowID := 9
		if targetID == "dst-1" {
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
	result, err := client.OpenDevToolsWindowIntoWindow(9, 7)
	if err == nil {
		t.Fatal("expected open devtools window into window to fail")
	}
	if !strings.Contains(err.Error(), "open devtools for target src-2 in window 7") {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SourceWindowID != 9 || result.TargetWindowID != 7 || result.Returned != 1 {
		t.Fatalf("unexpected partial result: %+v", result)
	}
	if len(result.OpenedTargets) != 1 || result.OpenedTargets[0].SourceTargetID != "src-1" {
		t.Fatalf("unexpected opened targets: %+v", result.OpenedTargets)
	}
}
