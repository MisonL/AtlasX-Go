package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
