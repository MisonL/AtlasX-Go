package tabs

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestBaseURLFromVersionEndpoint(t *testing.T) {
	baseURL, err := baseURLFromVersionEndpoint("http://127.0.0.1:9222/json/version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if baseURL != "http://127.0.0.1:9222" {
		t.Fatalf("unexpected base url: %s", baseURL)
	}
}

func TestListTargets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas"}]`))
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	targets, err := client.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("unexpected target count: %d", len(targets))
	}
}

func TestOpenTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/json/new" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if !strings.Contains(r.URL.RawQuery, "https%3A%2F%2Fopenai.com") {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"id":"2","type":"page","title":"OpenAI","url":"https://openai.com"}`))
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	target, err := client.Open("https://openai.com")
	if err != nil {
		t.Fatalf("open failed: %v", err)
	}
	if target.URL != "https://openai.com" {
		t.Fatalf("unexpected target url: %s", target.URL)
	}
}

func TestActivateTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/activate/123" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if err := client.Activate("123"); err != nil {
		t.Fatalf("activate failed: %v", err)
	}
}

func TestCloseTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/close/321" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if err := client.Close("321"); err != nil {
		t.Fatalf("close failed: %v", err)
	}
}

func TestPageTargetsFiltersNonPages(t *testing.T) {
	targets := []Target{
		{ID: "1", Type: "page"},
		{ID: "2", Type: "worker"},
		{ID: "3", Type: "iframe"},
	}

	pages := PageTargets(targets)
	if len(pages) != 1 {
		t.Fatalf("unexpected page count: %d", len(pages))
	}
	if pages[0].ID != "1" {
		t.Fatalf("unexpected page id: %s", pages[0].ID)
	}
}

func TestFindPageTarget(t *testing.T) {
	target, err := findPageTarget([]Target{
		{ID: "1", Type: "worker"},
		{ID: "2", Type: "page"},
	}, "2")
	if err != nil {
		t.Fatalf("find page target failed: %v", err)
	}
	if target.ID != "2" {
		t.Fatalf("unexpected target id: %s", target.ID)
	}
}

func TestFindPageTargetRejectsNonPage(t *testing.T) {
	_, err := findPageTarget([]Target{{ID: "1", Type: "worker"}}, "1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "is not a page") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCapturePageContext(t *testing.T) {
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

		response := cdpCommandResponse{
			ID: request.ID,
			Result: mustMarshalJSON(t, map[string]any{
				"result": map[string]any{
					"type": "object",
					"value": map[string]any{
						"text":           "Atlas page text",
						"text_truncated": false,
						"text_length":    15,
						"text_limit":     maxCapturedTextLength,
					},
				},
			}),
		}
		if err := connection.WriteJSON(response); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	context, err := client.Capture("1")
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if context.Title != "Atlas" {
		t.Fatalf("unexpected title: %s", context.Title)
	}
	if context.Text != "Atlas page text" {
		t.Fatalf("unexpected text: %s", context.Text)
	}
	if context.TextLength != 15 {
		t.Fatalf("unexpected text length: %d", context.TextLength)
	}
	if context.TextLimit != maxCapturedTextLength {
		t.Fatalf("unexpected text limit: %d", context.TextLimit)
	}
	if context.TextTruncated {
		t.Fatalf("expected text_truncated=false")
	}
	if context.CaptureError != "" {
		t.Fatalf("unexpected capture error: %s", context.CaptureError)
	}
	if _, err := time.Parse(time.RFC3339Nano, context.CapturedAt); err != nil {
		t.Fatalf("captured_at is not RFC3339: %v", err)
	}
}

func TestCapturePageContextIncludesTruncationMetadata(t *testing.T) {
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
			Result: mustMarshalJSON(t, map[string]any{
				"result": map[string]any{
					"type": "object",
					"value": map[string]any{
						"text":           strings.Repeat("A", maxCapturedTextLength),
						"text_truncated": true,
						"text_length":    maxCapturedTextLength + 128,
						"text_limit":     maxCapturedTextLength,
					},
				},
			}),
		}
		if err := connection.WriteJSON(response); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	context, err := client.Capture("1")
	if err != nil {
		t.Fatalf("capture failed: %v", err)
	}
	if !context.TextTruncated {
		t.Fatalf("expected text_truncated=true")
	}
	if len(context.Text) != maxCapturedTextLength {
		t.Fatalf("unexpected text size: %d", len(context.Text))
	}
	if context.TextLength != maxCapturedTextLength+128 {
		t.Fatalf("unexpected text length: %d", context.TextLength)
	}
}

func TestCapturePageContextReturnsStructuredFailure(t *testing.T) {
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
		if err := connection.WriteJSON(cdpCommandResponse{
			ID: request.ID,
			Error: &cdpError{
				Code:    -32000,
				Message: "capture failed",
			},
		}); err != nil {
			t.Fatalf("write response failed: %v", err)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	context, err := client.Capture("1")
	if err == nil {
		t.Fatal("expected error")
	}
	var captureErr *CaptureError
	if !errors.As(err, &captureErr) {
		t.Fatalf("expected CaptureError, got %T", err)
	}
	if context.ID != "1" || context.Title != "Atlas" || context.URL != "https://chatgpt.com/atlas" {
		t.Fatalf("unexpected context: %+v", context)
	}
	if context.CaptureError == "" {
		t.Fatal("expected capture_error")
	}
	if captureErr.Context.CaptureError != context.CaptureError {
		t.Fatalf("capture error mismatch: %+v", captureErr.Context)
	}
	if context.TextLimit != maxCapturedTextLength {
		t.Fatalf("unexpected text limit: %d", context.TextLimit)
	}
}

func TestCaptureTextExpressionUsesLimit(t *testing.T) {
	expression := captureTextExpression()
	if !strings.Contains(expression, "4096") {
		t.Fatalf("missing limit in expression: %s", expression)
	}
	if !strings.Contains(expression, "text_truncated") || !strings.Contains(expression, "text_length") {
		t.Fatalf("missing structured fields in expression: %s", expression)
	}
}

func mustMarshalJSON(t *testing.T, payload any) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	return data
}
