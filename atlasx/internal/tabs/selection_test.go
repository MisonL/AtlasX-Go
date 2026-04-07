package tabs

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestCaptureSelection(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/page/1"

	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","webSocketDebuggerUrl":"` + wsURL + `"}]`))
	})
	mux.HandleFunc("/devtools/page/1", func(w http.ResponseWriter, r *http.Request) {
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
		if !strings.Contains(expression, "window.getSelection") {
			t.Fatalf("unexpected expression: %s", expression)
		}

		response := cdpCommandResponse{
			ID: request.ID,
			Result: mustMarshalJSON(t, map[string]any{
				"result": map[string]any{
					"type": "object",
					"value": map[string]any{
						"selection_text":           "Atlas selected text",
						"selection_present":        true,
						"selection_text_truncated": false,
						"selection_text_length":    20,
						"selection_text_limit":     maxCapturedSelectionLength,
					},
				},
			}),
		}
		if err := connection.WriteJSON(response); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	context, err := client.CaptureSelection("1")
	if err != nil {
		t.Fatalf("capture selection failed: %v", err)
	}
	if !context.SelectionPresent {
		t.Fatalf("expected selection present: %+v", context)
	}
	if context.SelectionText != "Atlas selected text" {
		t.Fatalf("unexpected selection text: %+v", context)
	}
	if context.SelectionTextLength != 20 {
		t.Fatalf("unexpected selection text length: %+v", context)
	}
	if context.SelectionTextLimit != maxCapturedSelectionLength {
		t.Fatalf("unexpected selection text limit: %+v", context)
	}
}

func TestCaptureSelectionReturnsStructuredFailure(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/page/1"

	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","webSocketDebuggerUrl":"` + wsURL + `"}]`))
	})
	mux.HandleFunc("/devtools/page/1", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		defer connection.Close()

		var request cdpCommandRequest
		if err := connection.ReadJSON(&request); err != nil {
			t.Fatalf("read request failed: %v", err)
		}

		response := cdpCommandResponse{
			ID: request.ID,
			Error: &cdpError{
				Code:    -32000,
				Message: "selection failed",
			},
		}
		if err := connection.WriteJSON(response); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	_, err := client.CaptureSelection("1")
	if err == nil {
		t.Fatal("expected capture selection to fail")
	}

	var captureErr *SelectionCaptureError
	if !errors.As(err, &captureErr) {
		t.Fatalf("expected selection capture error, got %T", err)
	}
	if captureErr.Context.CaptureError == "" {
		t.Fatalf("expected capture error in context: %+v", captureErr.Context)
	}
}
