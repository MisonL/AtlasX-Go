package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gorilla/websocket"
)

func TestEmulateDeviceAppliesPreset(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/page/1"
	var requestCount int32

	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","webSocketDebuggerUrl":"` + wsURL + `"}]`))
	})
	mux.HandleFunc("/devtools/page/1", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		defer connection.Close()

		var first cdpCommandRequest
		if err := connection.ReadJSON(&first); err != nil {
			t.Fatalf("read first request failed: %v", err)
		}
		switch atomic.AddInt32(&requestCount, 1) {
		case 1:
			if first.Method != "Emulation.setDeviceMetricsOverride" {
				t.Fatalf("unexpected first method: %s", first.Method)
			}
			if first.Params["width"].(float64) != 390 || first.Params["height"].(float64) != 844 {
				t.Fatalf("unexpected first params: %+v", first.Params)
			}
		case 2:
			if first.Method != "Emulation.setTouchEmulationEnabled" {
				t.Fatalf("unexpected second method: %s", first.Method)
			}
			if first.Params["enabled"] != true {
				t.Fatalf("unexpected second params: %+v", first.Params)
			}
		default:
			t.Fatalf("unexpected request count")
		}
		if err := connection.WriteJSON(cdpCommandResponse{ID: first.ID, Result: mustMarshalJSON(t, map[string]any{})}); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	result, err := client.EmulateDevice("1", "iphone-13")
	if err != nil {
		t.Fatalf("emulate device failed: %v", err)
	}
	if result.Preset != "iphone-13" || result.Viewport != "390x844@3" || !result.Mobile || !result.Touch {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestEmulateDeviceClearsPreset(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/page/1"
	var requestCount int32

	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","webSocketDebuggerUrl":"` + wsURL + `"}]`))
	})
	mux.HandleFunc("/devtools/page/1", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		defer connection.Close()

		var first cdpCommandRequest
		if err := connection.ReadJSON(&first); err != nil {
			t.Fatalf("read first request failed: %v", err)
		}
		switch atomic.AddInt32(&requestCount, 1) {
		case 1:
			if first.Method != "Emulation.clearDeviceMetricsOverride" {
				t.Fatalf("unexpected first method: %s", first.Method)
			}
		case 2:
			if first.Method != "Emulation.setTouchEmulationEnabled" {
				t.Fatalf("unexpected second method: %s", first.Method)
			}
			if first.Params["enabled"] != false {
				t.Fatalf("unexpected second params: %+v", first.Params)
			}
		default:
			t.Fatalf("unexpected request count")
		}
		if err := connection.WriteJSON(cdpCommandResponse{ID: first.ID, Result: mustMarshalJSON(t, map[string]any{})}); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	result, err := client.EmulateDevice("1", "off")
	if err != nil {
		t.Fatalf("clear device emulation failed: %v", err)
	}
	if result.Preset != "off" || result.Viewport != "off" || result.Mobile || result.Touch {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestEmulateDeviceRejectsUnknownPreset(t *testing.T) {
	client := Client{}
	if _, err := client.EmulateDevice("1", "unknown"); err == nil {
		t.Fatal("expected unknown preset failure")
	} else if !strings.Contains(err.Error(), "unknown device preset") {
		t.Fatalf("unexpected error: %v", err)
	}
}
