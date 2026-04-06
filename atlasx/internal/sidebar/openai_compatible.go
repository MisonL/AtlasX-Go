package sidebar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"atlasx/internal/settings"
	"atlasx/internal/tabs"
)

const providerRequestTimeout = 10 * time.Second

type chatCompletionRequest struct {
	Model    string                  `json:"model"`
	Messages []chatCompletionMessage `json:"messages"`
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Model   string               `json:"model"`
	Error   *chatCompletionError `json:"error,omitempty"`
	Choices []struct {
		Message chatCompletionMessage `json:"message"`
	} `json:"choices"`
}

type chatCompletionError struct {
	Message string `json:"message"`
}

func providerSupported(provider string) bool {
	return provider == "openai" || provider == "openai-compatible" || provider == "openrouter"
}

func askOpenAICompatible(provider settings.SidebarProviderConfig, apiKey string, question string, context tabs.PageContext) (string, string, error) {
	return askChatCompletions(provider, apiKey, question, context, nil)
}

func askOpenRouter(provider settings.SidebarProviderConfig, apiKey string, question string, context tabs.PageContext) (string, string, error) {
	return askChatCompletions(provider, apiKey, question, context, map[string]string{
		"HTTP-Referer": "https://atlasx.local",
		"X-Title":      "AtlasX",
	})
}

func askChatCompletions(provider settings.SidebarProviderConfig, apiKey string, question string, context tabs.PageContext, extraHeaders map[string]string) (string, string, error) {
	payload := chatCompletionRequest{
		Model: provider.Model,
		Messages: []chatCompletionMessage{
			{
				Role:    "system",
				Content: "Answer the user's question using only the provided browser page context. If the context is insufficient, say so explicitly.",
			},
			{
				Role:    "user",
				Content: buildPrompt(question, context),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}

	request, err := http.NewRequest(http.MethodPost, strings.TrimRight(provider.BaseURL, "/")+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	request.Header.Set("Authorization", "Bearer "+apiKey)
	request.Header.Set("Content-Type", "application/json")
	for key, value := range extraHeaders {
		request.Header.Set(key, value)
	}

	client := http.Client{Timeout: providerRequestTimeout}
	response, err := client.Do(request)
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", err
	}

	var decoded chatCompletionResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return "", "", fmt.Errorf("decode provider response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		if decoded.Error != nil && decoded.Error.Message != "" {
			return "", "", fmt.Errorf("provider status %d: %s", response.StatusCode, decoded.Error.Message)
		}
		return "", "", fmt.Errorf("provider status %d", response.StatusCode)
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", "", fmt.Errorf("provider returned no answer")
	}

	model := decoded.Model
	if model == "" {
		model = provider.Model
	}
	return decoded.Choices[0].Message.Content, model, nil
}

func buildPrompt(question string, context tabs.PageContext) string {
	return fmt.Sprintf(
		"Question:\n%s\n\nContext Summary:\n%s\n\nPage Text:\n%s",
		question,
		buildContextSummary(context),
		context.Text,
	)
}

func buildContextSummary(context tabs.PageContext) string {
	return fmt.Sprintf(
		"title=%q url=%s captured_at=%s text_length=%d text_limit=%d text_truncated=%t",
		context.Title,
		context.URL,
		context.CapturedAt,
		context.TextLength,
		context.TextLimit,
		context.TextTruncated,
	)
}
