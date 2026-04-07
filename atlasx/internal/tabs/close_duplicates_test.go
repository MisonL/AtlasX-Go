package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCloseDuplicatesClosesLaterMatchingPages(t *testing.T) {
	var closedTargets []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json/list":
			_, _ = w.Write([]byte(`[
				{"id":"tab-1","type":"page","title":"Docs","url":"https://OpenAI.com/docs#top"},
				{"id":"tab-2","type":"page","title":"Docs Copy","url":"https://openai.com/docs"},
				{"id":"tab-3","type":"page","title":"Blog","url":"https://openai.com/blog"},
				{"id":"worker-1","type":"service_worker","title":"Worker","url":"https://openai.com/docs"}
			]`))
		case "/json/close/tab-2":
			closedTargets = append(closedTargets, "tab-2")
			w.WriteHeader(http.StatusOK)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	result, err := client.CloseDuplicates()
	if err != nil {
		t.Fatalf("close duplicates failed: %v", err)
	}
	if result.Returned != 1 {
		t.Fatalf("unexpected returned: %+v", result)
	}
	if len(result.Groups) != 1 {
		t.Fatalf("unexpected groups: %+v", result.Groups)
	}
	group := result.Groups[0]
	if group.URL != "https://openai.com/docs" || group.KeptTargetID != "tab-1" {
		t.Fatalf("unexpected group: %+v", group)
	}
	if len(group.ClosedTargetIDs) != 1 || group.ClosedTargetIDs[0] != "tab-2" {
		t.Fatalf("unexpected closed targets: %+v", group)
	}
	if len(closedTargets) != 1 || closedTargets[0] != "tab-2" {
		t.Fatalf("unexpected close calls: %+v", closedTargets)
	}
}

func TestCloseDuplicatesReturnsEmptyWhenNoDuplicates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json/list":
			_, _ = w.Write([]byte(`[
				{"id":"tab-1","type":"page","title":"Docs","url":"https://openai.com/docs"},
				{"id":"tab-2","type":"page","title":"Blog","url":"https://openai.com/blog"}
			]`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	result, err := client.CloseDuplicates()
	if err != nil {
		t.Fatalf("close duplicates failed: %v", err)
	}
	if result.Returned != 0 || len(result.Groups) != 0 || len(result.ClosedTargets) != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestCloseDuplicatesSurfacesTargetCloseFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json/list":
			_, _ = w.Write([]byte(`[
				{"id":"tab-1","type":"page","title":"Docs","url":"https://openai.com/docs"},
				{"id":"tab-2","type":"page","title":"Docs Copy","url":"https://openai.com/docs"}
			]`))
		case "/json/close/tab-2":
			http.Error(w, "close failed", http.StatusInternalServerError)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	if _, err := client.CloseDuplicates(); err == nil {
		t.Fatal("expected close duplicates to fail")
	} else if !strings.Contains(err.Error(), "unexpected status 500") {
		t.Fatalf("unexpected error: %v", err)
	}
}
