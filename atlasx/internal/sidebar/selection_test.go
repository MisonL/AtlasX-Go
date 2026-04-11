package sidebar

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlasx/internal/settings"
	"atlasx/internal/tabs"
)

func TestBuildSelectionQuestionRejectsEmptySelection(t *testing.T) {
	_, err := BuildSelectionQuestion("   ", "what matters here")
	if err == nil || err.Error() != "selection_text is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAskSelectionWithOpenAICompatibleProvider(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-key")

	var capturedBody map[string]any
	handlerErrCh := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			reportHandlerError(handlerErrCh, "read body failed: %v", err)
			http.Error(w, "read body failed", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(body, &capturedBody); err != nil {
			reportHandlerError(handlerErrCh, "decode body failed: %v", err)
			http.Error(w, "decode body failed", http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"model":"gpt-5.4","choices":[{"message":{"content":"Selected answer"}}]}`))
	}))
	defer server.Close()

	config := FromSettings(settings.Config{
		SidebarDefaultProvider: "primary",
		SidebarProviders: []settings.SidebarProviderConfig{
			{
				ID:        "primary",
				Provider:  "openai",
				Model:     "gpt-5.4",
				BaseURL:   server.URL,
				APIKeyEnv: "OPENAI_API_KEY",
			},
		},
	})

	response, err := config.AskSelectionWithMemory(SelectionAskRequest{
		TabID:         "tab-1",
		SelectionText: "Atlas isolates profile state.",
		Question:      "这句话是什么意思？",
	}, tabs.PageContext{
		ID:            "tab-1",
		Title:         "Atlas",
		URL:           "https://chatgpt.com/atlas",
		Text:          "Atlas page text",
		CapturedAt:    "2026-04-07T12:00:00Z",
		TextLength:    15,
		TextLimit:     4096,
		TextTruncated: false,
	}, []string{
		`qa_turn occurred_at=2026-04-07T11:59:00Z title="Atlas" url=https://chatgpt.com/atlas question="what is atlas" answer="Atlas is memory-aware." cited_urls=https://chatgpt.com/atlas`,
	})
	if err != nil {
		t.Fatalf("ask selection failed: %v", err)
	}
	requireNoHandlerError(t, handlerErrCh)
	if response.Answer != "Selected answer" {
		t.Fatalf("unexpected response: %+v", response)
	}
	if response.Provider != "openai" || response.Model != "gpt-5.4" {
		t.Fatalf("unexpected response: %+v", response)
	}

	messages, ok := capturedBody["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("unexpected request body: %+v", capturedBody)
	}
	userMessage, ok := messages[1].(map[string]any)
	if !ok {
		t.Fatalf("unexpected request body: %+v", capturedBody)
	}
	content, _ := userMessage["content"].(string)
	if !strings.Contains(content, "Selected Text:\nAtlas isolates profile state.") {
		t.Fatalf("missing selection text: %s", content)
	}
	if !strings.Contains(content, "User Question:\n这句话是什么意思？") {
		t.Fatalf("missing selection question: %s", content)
	}
	if !strings.Contains(content, "Relevant Memory:") {
		t.Fatalf("missing memory section: %s", content)
	}
}
