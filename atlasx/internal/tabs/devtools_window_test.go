package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gorilla/websocket"
)

func TestOpenDevToolsWindowUsesResolvedFrontendURL(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	var listCallCount int32

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		switch atomic.AddInt32(&listCallCount, 1) {
		case 1:
			_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1/devtools/page/tab-1"}]`))
		default:
			_, _ = w.Write([]byte(`[{"id":"devtools-window-1","type":"page","title":"DevTools","url":"http://127.0.0.1/devtools/inspector.html?ws=127.0.0.1/devtools/page/tab-1"}]`))
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
		if request.Method != "Target.createTarget" {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.Params["url"] != server.URL+"/devtools/inspector.html?ws=127.0.0.1/devtools/page/tab-1" {
			t.Fatalf("unexpected url: %+v", request.Params)
		}
		if err := connection.WriteJSON(cdpCommandResponse{
			ID:     request.ID,
			Result: mustMarshalJSON(t, map[string]any{"targetId": "devtools-window-1"}),
		}); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	target, err := client.OpenDevToolsWindow("tab-1")
	if err != nil {
		t.Fatalf("open devtools window failed: %v", err)
	}
	if target.ID != "devtools-window-1" || !strings.Contains(target.URL, "/devtools/inspector.html") {
		t.Fatalf("unexpected target: %+v", target)
	}
}

func TestOpenDevToolsWindowSurfacesMissingFrontendURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas"}]`))
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.OpenDevToolsWindow("tab-1"); err == nil {
		t.Fatal("expected missing frontend url failure")
	} else if !strings.Contains(err.Error(), "target does not expose a devtools frontend url") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenDevToolsInWindowActivatesWindowAndOpensDevToolsURL(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	var listCallCount int32
	steps := make([]string, 0, 2)

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		switch atomic.AddInt32(&listCallCount, 1) {
		case 1:
			_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1/devtools/page/tab-1"}]`))
		default:
			_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas"}]`))
		}
	})
	mux.HandleFunc("/json/activate/tab-1", func(w http.ResponseWriter, r *http.Request) {
		steps = append(steps, "activate")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/json/new", func(w http.ResponseWriter, r *http.Request) {
		steps = append(steps, "open")
		if !strings.Contains(r.URL.RawQuery, "devtools%2Finspector.html") {
			t.Fatalf("expected encoded devtools path in raw query, got %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"id":"devtools-window-tab","type":"page","title":"DevTools","url":"http://127.0.0.1/devtools/inspector.html?ws=127.0.0.1%2Fdevtools%2Fpage%2Ftab-1"}`))
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
	result, err := client.OpenDevToolsInWindow("tab-1", 7)
	if err != nil {
		t.Fatalf("open devtools in window failed: %v", err)
	}
	if result.WindowID != 7 || result.ActivatedTargetID != "tab-1" || result.Target.ID != "devtools-window-tab" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(steps) != 2 || steps[0] != "activate" || steps[1] != "open" {
		t.Fatalf("unexpected steps: %+v", steps)
	}
}

func TestOpenDevToolsInWindowRejectsUnknownWindow(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	var listCallCount int32

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		switch atomic.AddInt32(&listCallCount, 1) {
		case 1:
			_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1/devtools/page/tab-1"}]`))
		default:
			_, _ = w.Write([]byte(`[]`))
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
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.OpenDevToolsInWindow("tab-1", 7); err == nil {
		t.Fatal("expected unknown window failure")
	} else if !strings.Contains(err.Error(), "window 7 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenDevToolsPanelWindowUsesPanelFrontendURL(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	var listCallCount int32

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		switch atomic.AddInt32(&listCallCount, 1) {
		case 1:
			_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1/devtools/page/tab-1"}]`))
		default:
			_, _ = w.Write([]byte(`[{"id":"devtools-window-1","type":"page","title":"DevTools","url":"http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%2Fdevtools%2Fpage%2Ftab-1"}]`))
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
		if request.Method != "Target.createTarget" {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if request.Params["url"] != server.URL+"/devtools/inspector.html?panel=network&ws=127.0.0.1%2Fdevtools%2Fpage%2Ftab-1" {
			t.Fatalf("unexpected url: %+v", request.Params)
		}
		if err := connection.WriteJSON(cdpCommandResponse{
			ID:     request.ID,
			Result: mustMarshalJSON(t, map[string]any{"targetId": "devtools-window-1"}),
		}); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	target, err := client.OpenDevToolsPanelWindow("tab-1", "network")
	if err != nil {
		t.Fatalf("open devtools panel window failed: %v", err)
	}
	if target.ID != "devtools-window-1" || !strings.Contains(target.URL, "panel=network") {
		t.Fatalf("unexpected target: %+v", target)
	}
}

func TestOpenDevToolsPanelWindowRejectsBlankPanel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1/devtools/page/tab-1"}]`))
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.OpenDevToolsPanelWindow("tab-1", ""); err == nil {
		t.Fatal("expected blank panel failure")
	} else if !strings.Contains(err.Error(), "panel is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenDevToolsPanelInWindowActivatesWindowAndOpensPanelURL(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	var listCallCount int32
	steps := make([]string, 0, 2)

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		switch atomic.AddInt32(&listCallCount, 1) {
		case 1:
			_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1/devtools/page/tab-1"}]`))
		default:
			_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas"}]`))
		}
	})
	mux.HandleFunc("/json/activate/tab-1", func(w http.ResponseWriter, r *http.Request) {
		steps = append(steps, "activate")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/json/new", func(w http.ResponseWriter, r *http.Request) {
		steps = append(steps, "open")
		if !strings.Contains(r.URL.RawQuery, "panel%3Dnetwork") {
			t.Fatalf("expected encoded panel in raw query, got %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"id":"devtools-window-tab","type":"page","title":"DevTools","url":"http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%2Fdevtools%2Fpage%2Ftab-1"}`))
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
	result, err := client.OpenDevToolsPanelInWindow("tab-1", "network", 7)
	if err != nil {
		t.Fatalf("open devtools panel in window failed: %v", err)
	}
	if result.WindowID != 7 || result.ActivatedTargetID != "tab-1" || result.Target.ID != "devtools-window-tab" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(steps) != 2 || steps[0] != "activate" || steps[1] != "open" {
		t.Fatalf("unexpected steps: %+v", steps)
	}
}

func TestOpenDevToolsPanelInWindowRejectsUnknownWindow(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	var listCallCount int32

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		switch atomic.AddInt32(&listCallCount, 1) {
		case 1:
			_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1/devtools/page/tab-1"}]`))
		default:
			_, _ = w.Write([]byte(`[]`))
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
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.OpenDevToolsPanelInWindow("tab-1", "network", 7); err == nil {
		t.Fatal("expected unknown window failure")
	} else if !strings.Contains(err.Error(), "window 7 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}
