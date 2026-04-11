package tabs

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDevToolsTargetResolvesRelativeFrontendURL(t *testing.T) {
	handlerErrs := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/list" {
			select {
			case handlerErrs <- "unexpected path: " + r.URL.Path:
			default:
			}
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1:9222/devtools/page/1"}]`))
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	target, err := client.DevTools("1")
	if err != nil {
		t.Fatalf("devtools lookup failed: %v", err)
	}
	if target.DevToolsFrontendURL != server.URL+"/devtools/inspector.html?ws=127.0.0.1:9222/devtools/page/1" {
		t.Fatalf("unexpected devtools url: %+v", target)
	}
	select {
	case handlerErr := <-handlerErrs:
		t.Fatal(handlerErr)
	default:
	}
}

func TestDevToolsTargetRejectsMissingFrontendURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas"}]`))
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	_, err := client.DevTools("1")
	if err == nil || err.Error() != "target does not expose a devtools frontend url" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveDevToolsPanelURLAppendsPanelQuery(t *testing.T) {
	panelURL, err := resolveDevToolsPanelURL("http://127.0.0.1:9222/devtools/inspector.html?ws=127.0.0.1:9222/devtools/page/1", "network")
	if err != nil {
		t.Fatalf("resolve devtools panel url failed: %v", err)
	}
	expected := "http://127.0.0.1:9222/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2F1"
	if panelURL != expected {
		t.Fatalf("unexpected panel url: %s", panelURL)
	}
}

func TestResolveDevToolsPanelURLRejectsBlankPanel(t *testing.T) {
	if _, err := resolveDevToolsPanelURL("http://127.0.0.1:9222/devtools/inspector.html?ws=127.0.0.1:9222/devtools/page/1", ""); err == nil {
		t.Fatal("expected blank panel failure")
	} else if err.Error() != "panel is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDevToolsPanelTargetResolvesPanelFrontendURL(t *testing.T) {
	handlerErrs := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/list" {
			select {
			case handlerErrs <- "unexpected path: " + r.URL.Path:
			default:
			}
			http.Error(w, "unexpected path", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`[{"id":"1","type":"page","title":"Atlas","url":"https://chatgpt.com/atlas","devtoolsFrontendUrl":"/devtools/inspector.html?ws=127.0.0.1:9222/devtools/page/1"}]`))
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	target, err := client.DevToolsPanel("1", "network")
	if err != nil {
		t.Fatalf("devtools panel lookup failed: %v", err)
	}
	expected := server.URL + "/devtools/inspector.html?panel=network&ws=127.0.0.1%3A9222%2Fdevtools%2Fpage%2F1"
	if target.DevToolsFrontendURL != expected {
		t.Fatalf("unexpected devtools panel url: %+v", target)
	}
	select {
	case handlerErr := <-handlerErrs:
		t.Fatal(handlerErr)
	default:
	}
}
