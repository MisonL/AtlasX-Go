package tabs

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
)

const maxCapturedTextBytes = 4096

type PageContext struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Text  string `json:"text"`
}

type runtimeEvaluateResult struct {
	Result runtimeRemoteObject `json:"result"`
}

type runtimeRemoteObject struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (c Client) Capture(targetID string) (PageContext, error) {
	targets, err := c.List()
	if err != nil {
		return PageContext{}, err
	}

	target, err := findPageTarget(targets, targetID)
	if err != nil {
		return PageContext{}, err
	}
	if target.WebSocketDebuggerURL == "" {
		return PageContext{}, errors.New("target does not expose a websocket debugger url")
	}

	text, err := capturePageText(target.WebSocketDebuggerURL)
	if err != nil {
		return PageContext{}, err
	}

	return PageContext{
		ID:    target.ID,
		Title: target.Title,
		URL:   target.URL,
		Text:  text,
	}, nil
}

func capturePageText(websocketURL string) (string, error) {
	response, err := runPageCommand(websocketURL, cdpCommandRequest{
		ID:     1,
		Method: "Runtime.evaluate",
		Params: map[string]any{
			"expression":    captureTextExpression(),
			"returnByValue": true,
		},
	})
	if err != nil {
		return "", err
	}

	var payload runtimeEvaluateResult
	if err := json.Unmarshal(response.Result, &payload); err != nil {
		return "", err
	}
	if payload.Result.Type != "string" {
		return "", fmt.Errorf("unexpected runtime result type %q", payload.Result.Type)
	}
	return payload.Result.Value, nil
}

func runPageCommand(websocketURL string, request cdpCommandRequest) (cdpCommandResponse, error) {
	connection, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		return cdpCommandResponse{}, err
	}
	defer connection.Close()

	if err := connection.WriteJSON(request); err != nil {
		return cdpCommandResponse{}, err
	}

	for {
		var response cdpCommandResponse
		if err := connection.ReadJSON(&response); err != nil {
			return cdpCommandResponse{}, err
		}
		if response.ID != request.ID {
			continue
		}
		if response.Error != nil {
			return cdpCommandResponse{}, fmt.Errorf("cdp error %d: %s", response.Error.Code, response.Error.Message)
		}
		return response, nil
	}
}

func captureTextExpression() string {
	return fmt.Sprintf(`(() => {
		const text = document.body ? (document.body.innerText || "") : "";
		return text.slice(0, %d);
	})()`, maxCapturedTextBytes)
}
