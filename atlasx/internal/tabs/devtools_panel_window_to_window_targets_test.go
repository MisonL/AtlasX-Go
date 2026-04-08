package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestOpenDevToolsPanelWindowToWindowsOpensEachSourceTargetInNewWindow(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	createCalls := 0
	createURLs := make([]string, 0, 2)

	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[
			{"id":"src-1","type":"page","title":"Source One","url":"https://openai.com/one","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1"},
			{"id":"src-2","type":"page","title":"Source Two","url":"https://openai.com/two","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-2"},
			{"id":"new-1","type":"page","title":"DevTools One","url":"http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-1"},
			{"id":"new-2","type":"page","title":"DevTools Two","url":"http://127.0.0.1/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2Fsrc-2"}
		]`))
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
			targetID, _ := request.Params["targetId"].(string)
			windowID := 7
			if strings.HasPrefix(targetID, "src-") {
				windowID = 9
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
			createCalls++
			openURL, _ := request.Params["url"].(string)
			createURLs = append(createURLs, openURL)
			targetID := "new-1"
			if createCalls == 2 {
				targetID = "new-2"
			}
			if err := connection.WriteJSON(cdpCommandResponse{
				ID:     request.ID,
				Result: mustMarshalJSON(t, map[string]any{"targetId": targetID}),
			}); err != nil {
				t.Fatalf("write response failed: %v", err)
			}
		default:
			t.Fatalf("unexpected method: %s", request.Method)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	result, err := client.OpenDevToolsPanelWindowToWindows(9, "network")
	if err != nil {
		t.Fatalf("open devtools panel window to windows failed: %v", err)
	}
	if result.SourceWindowID != 9 || result.Panel != "network" || result.Returned != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.OpenedTargets) != 2 || result.OpenedTargets[0].SourceTargetID != "src-1" || result.OpenedTargets[1].SourceTargetID != "src-2" {
		t.Fatalf("unexpected opened targets: %+v", result.OpenedTargets)
	}
	if createCalls != 2 {
		t.Fatalf("unexpected create target calls: %d", createCalls)
	}
	if len(createURLs) != 2 {
		t.Fatalf("unexpected create target urls: %+v", createURLs)
	}
	for _, openedURL := range createURLs {
		if !strings.Contains(openedURL, "panel=network") {
			t.Fatalf("expected panel in url: %s", openedURL)
		}
	}
}

func TestOpenDevToolsPanelWindowToWindowsRejectsBlankPanel(t *testing.T) {
	client := Client{}
	if _, err := client.OpenDevToolsPanelWindowToWindows(9, " "); err == nil {
		t.Fatal("expected open devtools panel window to windows to fail")
	} else if !strings.Contains(err.Error(), "panel is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenDevToolsPanelWindowToWindowsRejectsUnknownSourceWindow(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[]`))
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
		if request.Method != "Browser.getWindowBounds" {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if err := connection.WriteJSON(cdpCommandResponse{
			ID: request.ID,
			Error: &cdpError{
				Code:    -32000,
				Message: "Browser window not found",
			},
		}); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.OpenDevToolsPanelWindowToWindows(9, "network"); err == nil {
		t.Fatal("expected open devtools panel window to windows to fail")
	} else if !strings.Contains(err.Error(), "window 9 not found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenDevToolsPanelWindowToWindowsReturnsEmptyWhenSourceWindowHasNoPageTargets(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	browserWSURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/browser/root"
	mux.HandleFunc("/json/version", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"` + browserWSURL + `"}`))
	})
	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[]`))
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
		if request.Method != "Browser.getWindowBounds" {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		if err := connection.WriteJSON(cdpCommandResponse{
			ID: request.ID,
			Result: mustMarshalJSON(t, map[string]any{
				"bounds": map[string]any{
					"left":        0,
					"top":         0,
					"width":       1200,
					"height":      800,
					"windowState": "normal",
				},
			}),
		}); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	result, err := client.OpenDevToolsPanelWindowToWindows(9, "network")
	if err != nil {
		t.Fatalf("open devtools panel window to windows failed: %v", err)
	}
	if result.SourceWindowID != 9 || result.Panel != "network" || result.Returned != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.OpenedTargets) != 0 {
		t.Fatalf("unexpected opened targets: %+v", result.OpenedTargets)
	}
}
