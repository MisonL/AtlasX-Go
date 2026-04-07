package sidebar

import (
	"errors"
	"fmt"
	"strings"

	"atlasx/internal/tabs"
)

const selectionQuestionTemplate = "Answer the user's question about the selected text below. Use the selected text first and only use other current page context when needed.\n\nSelected Text:\n%s\n\nUser Question:\n%s"

type SelectionAskRequest struct {
	TabID         string `json:"tab_id"`
	SelectionText string `json:"selection_text"`
	Question      string `json:"question"`
	ProviderID    string `json:"provider_id,omitempty"`
}

func BuildSelectionQuestion(selectionText string, question string) (string, error) {
	trimmedSelection := strings.TrimSpace(selectionText)
	if trimmedSelection == "" {
		return "", errors.New("selection_text is required")
	}

	trimmedQuestion := strings.TrimSpace(question)
	if trimmedQuestion == "" {
		return "", errors.New("question is required")
	}

	return fmt.Sprintf(selectionQuestionTemplate, trimmedSelection, trimmedQuestion), nil
}

func (c Config) AskSelection(request SelectionAskRequest, context tabs.PageContext) (AskResponse, error) {
	return c.AskSelectionWithMemory(request, context, nil)
}

func (c Config) AskSelectionWithMemory(request SelectionAskRequest, context tabs.PageContext, memorySnippets []string) (AskResponse, error) {
	if request.TabID == "" {
		return AskResponse{}, errors.New("tab_id is required")
	}

	question, err := BuildSelectionQuestion(request.SelectionText, request.Question)
	if err != nil {
		return AskResponse{}, err
	}

	return c.AskWithMemory(AskRequest{
		TabID:      request.TabID,
		Question:   question,
		ProviderID: request.ProviderID,
	}, context, memorySnippets)
}
