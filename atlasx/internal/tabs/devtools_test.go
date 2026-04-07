package tabs

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDevToolsTargetResolvesRelativeFrontendURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/json/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
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
