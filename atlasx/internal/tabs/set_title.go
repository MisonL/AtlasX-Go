package tabs

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type TitleUpdateResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type titleEvaluateResult struct {
	Result titleRemoteObject `json:"result"`
}

type titleRemoteObject struct {
	Type  string           `json:"type"`
	Value titleResultValue `json:"value"`
}

type titleResultValue struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func (c Client) SetTitle(targetID string, title string) (TitleUpdateResult, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return TitleUpdateResult{}, errors.New("title is required")
	}

	targets, err := c.List()
	if err != nil {
		return TitleUpdateResult{}, err
	}

	target, err := findPageTarget(targets, targetID)
	if err != nil {
		return TitleUpdateResult{}, err
	}
	if target.WebSocketDebuggerURL == "" {
		return TitleUpdateResult{}, errors.New("target does not expose a websocket debugger url")
	}

	value, err := updatePageTitle(target.WebSocketDebuggerURL, title)
	if err != nil {
		return TitleUpdateResult{}, err
	}
	return TitleUpdateResult{
		ID:    target.ID,
		Title: value.Title,
		URL:   value.URL,
	}, nil
}

func updatePageTitle(websocketURL string, title string) (titleResultValue, error) {
	expression, err := setTitleExpression(title)
	if err != nil {
		return titleResultValue{}, err
	}

	response, err := runPageCommand(websocketURL, cdpCommandRequest{
		ID:     4,
		Method: "Runtime.evaluate",
		Params: map[string]any{
			"expression":    expression,
			"returnByValue": true,
		},
	})
	if err != nil {
		return titleResultValue{}, err
	}

	var payload titleEvaluateResult
	if err := json.Unmarshal(response.Result, &payload); err != nil {
		return titleResultValue{}, err
	}
	if payload.Result.Type != "object" {
		return titleResultValue{}, fmt.Errorf("unexpected runtime result type %q", payload.Result.Type)
	}
	return payload.Result.Value, nil
}

func setTitleExpression(title string) (string, error) {
	encodedTitle, err := json.Marshal(title)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`(() => {
		const nextTitle = %s;
		document.title = nextTitle;
		return {
			title: document.title || "",
			url: location.href || ""
		};
	})()`, string(encodedTitle)), nil
}
