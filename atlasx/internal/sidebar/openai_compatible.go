package sidebar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"atlasx/internal/settings"
	"atlasx/internal/tabs"
)

const providerRequestTimeout = 2 * time.Second
const providerRetryAttempts = 1
const providerTokenBudget = 1200

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

func askOpenAICompatible(provider settings.SidebarProviderConfig, apiKey string, question string, context tabs.PageContext, memorySnippets []string) (string, string, error) {
	return askChatCompletions(provider, apiKey, question, context, memorySnippets, nil)
}

func askOpenRouter(provider settings.SidebarProviderConfig, apiKey string, question string, context tabs.PageContext, memorySnippets []string) (string, string, error) {
	return askChatCompletions(provider, apiKey, question, context, memorySnippets, map[string]string{
		"HTTP-Referer": "https://atlasx.local",
		"X-Title":      "AtlasX",
	})
}

func askChatCompletions(provider settings.SidebarProviderConfig, apiKey string, question string, context tabs.PageContext, memorySnippets []string, extraHeaders map[string]string) (string, string, error) {
	prompt := buildPrompt(question, context, memorySnippets)
	if estimatePromptTokens(prompt) > providerTokenBudget {
		return "", "", ErrTokenBudgetExceeded
	}

	payload := chatCompletionRequest{
		Model: provider.Model,
		Messages: []chatCompletionMessage{
			{
				Role:    "system",
				Content: "Answer the user's question using only the provided browser page context. If the context is insufficient, say so explicitly.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}

	endpoint := strings.TrimRight(provider.BaseURL, "/") + "/chat/completions"
	client := http.Client{Timeout: providerRequestTimeout}
	var lastErr error
	for attempt := 0; attempt <= providerRetryAttempts; attempt++ {
		request, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return "", "", err
		}
		request.Header.Set("Authorization", "Bearer "+apiKey)
		request.Header.Set("Content-Type", "application/json")
		for key, value := range extraHeaders {
			request.Header.Set(key, value)
		}

		response, err := client.Do(request)
		if err != nil {
			lastErr = err
			if !retryableProviderError(err) || attempt == providerRetryAttempts {
				break
			}
			continue
		}

		answer, model, retryable, err := decodeProviderResponse(response, provider.Model)
		if err == nil {
			return answer, model, nil
		}
		lastErr = err
		if !retryable || attempt == providerRetryAttempts {
			break
		}
	}
	if lastErr == nil {
		lastErr = errors.New("provider returned no answer")
	}
	return "", "", lastErr
}

func buildPrompt(question string, context tabs.PageContext, memorySnippets []string) string {
	basePrompt := fmt.Sprintf(
		"Question:\n%s\n\nContext Summary:\n%s\n\nPage Text:\n%s",
		question,
		buildContextSummary(context),
		context.Text,
	)
	if len(memorySnippets) == 0 || estimatePromptTokens(basePrompt) >= providerTokenBudget {
		return basePrompt
	}

	memoryLines := make([]string, 0, len(memorySnippets))
	for _, snippet := range memorySnippets {
		if strings.TrimSpace(snippet) == "" {
			continue
		}

		candidateLines := append(memoryLines, fmt.Sprintf("%d. %s", len(memoryLines)+1, snippet))
		candidatePrompt := fmt.Sprintf(
			"Question:\n%s\n\nContext Summary:\n%s\n\nRelevant Memory:\n%s\n\nPage Text:\n%s",
			question,
			buildContextSummary(context),
			strings.Join(candidateLines, "\n"),
			context.Text,
		)
		if estimatePromptTokens(candidatePrompt) > providerTokenBudget {
			break
		}
		memoryLines = candidateLines
	}
	if len(memoryLines) == 0 {
		return basePrompt
	}

	return fmt.Sprintf(
		"Question:\n%s\n\nContext Summary:\n%s\n\nRelevant Memory:\n%s\n\nPage Text:\n%s",
		question,
		buildContextSummary(context),
		strings.Join(memoryLines, "\n"),
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

func decodeProviderResponse(response *http.Response, fallbackModel string) (string, string, bool, error) {
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", false, err
	}

	var decoded chatCompletionResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return "", "", false, fmt.Errorf("decode provider response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		if decoded.Error != nil && decoded.Error.Message != "" {
			return "", "", response.StatusCode >= http.StatusInternalServerError, fmt.Errorf("provider status %d: %s", response.StatusCode, decoded.Error.Message)
		}
		return "", "", response.StatusCode >= http.StatusInternalServerError, fmt.Errorf("provider status %d", response.StatusCode)
	}
	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", "", false, fmt.Errorf("provider returned no answer")
	}

	model := decoded.Model
	if model == "" {
		model = fallbackModel
	}
	return decoded.Choices[0].Message.Content, model, false, nil
}

func estimatePromptTokens(prompt string) int {
	if prompt == "" {
		return 0
	}
	return (len([]rune(prompt)) + 3) / 4
}

func retryableProviderError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		return true
	}
	return false
}

func defaultProviderTimeoutMS() int {
	return int(providerRequestTimeout / time.Millisecond)
}

func defaultProviderRetryAttempts() int {
	return providerRetryAttempts
}

func defaultProviderTokenBudget() int {
	return providerTokenBudget
}
