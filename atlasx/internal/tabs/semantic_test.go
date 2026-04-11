package tabs

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestCaptureSemanticContext(t *testing.T) {
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
		defer func() {
			_ = connection.Close()
		}()

		var request cdpCommandRequest
		if err := connection.ReadJSON(&request); err != nil {
			t.Fatalf("read request failed: %v", err)
		}
		if request.Method != "Runtime.evaluate" {
			t.Fatalf("unexpected method: %s", request.Method)
		}

		response := cdpCommandResponse{
			ID: request.ID,
			Result: mustMarshalJSON(t, map[string]any{
				"result": map[string]any{
					"type": "object",
					"value": map[string]any{
						"headings": []map[string]any{{"level": 1, "text": "Atlas"}},
						"links":    []map[string]any{{"text": "OpenAI Docs", "url": "https://platform.openai.com/docs"}},
						"forms":    []map[string]any{{"action": "https://chatgpt.com/search", "method": "GET", "input_count": 2}},
					},
				},
			}),
		}
		if err := connection.WriteJSON(response); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	context, err := client.CaptureSemanticContext("1")
	if err != nil {
		t.Fatalf("capture semantic context failed: %v", err)
	}
	if context.Returned != 3 || context.HeadingsReturned != 1 || context.LinksReturned != 1 || context.FormsReturned != 1 {
		t.Fatalf("unexpected counts: %+v", context)
	}
	if len(context.Headings) != 1 || context.Headings[0].Text != "Atlas" {
		t.Fatalf("unexpected headings: %+v", context.Headings)
	}
	if len(context.Links) != 1 || context.Links[0].URL != "https://platform.openai.com/docs" {
		t.Fatalf("unexpected links: %+v", context.Links)
	}
	if len(context.Forms) != 1 || context.Forms[0].InputCount != 2 {
		t.Fatalf("unexpected forms: %+v", context.Forms)
	}
	if context.CaptureError != "" {
		t.Fatalf("unexpected capture error: %s", context.CaptureError)
	}
	if _, err := time.Parse(time.RFC3339Nano, context.CapturedAt); err != nil {
		t.Fatalf("captured_at is not RFC3339: %v", err)
	}
}

func TestCaptureSemanticContextReturnsStructuredFailure(t *testing.T) {
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
		defer func() {
			_ = connection.Close()
		}()

		var request cdpCommandRequest
		if err := connection.ReadJSON(&request); err != nil {
			t.Fatalf("read request failed: %v", err)
		}

		response := cdpCommandResponse{
			ID: request.ID,
			Error: &cdpError{
				Code:    -32000,
				Message: "semantic capture failed",
			},
		}
		if err := connection.WriteJSON(response); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	_, err := client.CaptureSemanticContext("1")
	if err == nil {
		t.Fatal("expected capture semantic context to fail")
	}
	var captureErr *SemanticCaptureError
	if !errors.As(err, &captureErr) {
		t.Fatalf("expected semantic capture error, got %T", err)
	}
	if captureErr.Context.CaptureError != "cdp error -32000: semantic capture failed" {
		t.Fatalf("unexpected capture error: %+v", captureErr.Context)
	}
	if len(captureErr.Context.Headings) != 0 || len(captureErr.Context.Links) != 0 || len(captureErr.Context.Forms) != 0 {
		t.Fatalf("expected empty collections on failure: %+v", captureErr.Context)
	}
}
