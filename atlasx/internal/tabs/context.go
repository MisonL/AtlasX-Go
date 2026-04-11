package tabs

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

const maxCapturedTextLength = 4096
const cdpCommandTimeout = 10 * time.Second

type PageContext struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	URL           string `json:"url"`
	Text          string `json:"text"`
	CapturedAt    string `json:"captured_at"`
	TextTruncated bool   `json:"text_truncated"`
	TextLength    int    `json:"text_length"`
	TextLimit     int    `json:"text_limit"`
	CaptureError  string `json:"capture_error"`
}

type CaptureError struct {
	Context PageContext
	Cause   error
}

func (e *CaptureError) Error() string {
	if e == nil || e.Cause == nil {
		return "capture error: unknown cause"
	}
	return e.Cause.Error()
}

func (e *CaptureError) Unwrap() error {
	return e.Cause
}

type runtimeEvaluateResult struct {
	Result runtimeRemoteObject `json:"result"`
}

type runtimeRemoteObject struct {
	Type  string            `json:"type"`
	Value captureTextResult `json:"value"`
}

type captureTextResult struct {
	Text          string `json:"text"`
	TextTruncated bool   `json:"text_truncated"`
	TextLength    int    `json:"text_length"`
	TextLimit     int    `json:"text_limit"`
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

	context := PageContext{
		ID:         target.ID,
		Title:      target.Title,
		URL:        target.URL,
		CapturedAt: time.Now().UTC().Format(time.RFC3339Nano),
		TextLimit:  maxCapturedTextLength,
	}
	if target.WebSocketDebuggerURL == "" {
		return captureFailure(context, errors.New("target does not expose a websocket debugger url"))
	}

	textResult, err := capturePageText(target.WebSocketDebuggerURL)
	if err != nil {
		return captureFailure(context, err)
	}

	context.Text = textResult.Text
	context.TextTruncated = textResult.TextTruncated
	context.TextLength = textResult.TextLength
	if textResult.TextLimit > 0 {
		context.TextLimit = textResult.TextLimit
	}
	return context, nil
}

func captureFailure(context PageContext, err error) (PageContext, error) {
	context.CaptureError = err.Error()
	return context, &CaptureError{
		Context: context,
		Cause:   err,
	}
}

func capturePageText(websocketURL string) (captureTextResult, error) {
	response, err := runPageCommand(websocketURL, cdpCommandRequest{
		ID:     1,
		Method: "Runtime.evaluate",
		Params: map[string]any{
			"expression":    captureTextExpression(),
			"returnByValue": true,
		},
	})
	if err != nil {
		return captureTextResult{}, err
	}

	var payload runtimeEvaluateResult
	if err := json.Unmarshal(response.Result, &payload); err != nil {
		return captureTextResult{}, err
	}
	if payload.Result.Type != "object" {
		return captureTextResult{}, fmt.Errorf("unexpected runtime result type %q", payload.Result.Type)
	}
	if payload.Result.Value.TextLimit == 0 {
		payload.Result.Value.TextLimit = maxCapturedTextLength
	}
	return payload.Result.Value, nil
}

func runPageCommand(websocketURL string, request cdpCommandRequest) (cdpCommandResponse, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: cdpCommandTimeout,
	}
	connection, _, err := dialer.Dial(websocketURL, nil)
	if err != nil {
		return cdpCommandResponse{}, err
	}
	defer func() {
		_ = connection.Close()
	}()

	if err := connection.SetWriteDeadline(time.Now().Add(cdpCommandTimeout)); err != nil {
		return cdpCommandResponse{}, err
	}
	if err := connection.WriteJSON(request); err != nil {
		return cdpCommandResponse{}, err
	}

	for {
		if err := connection.SetReadDeadline(time.Now().Add(cdpCommandTimeout)); err != nil {
			return cdpCommandResponse{}, err
		}
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
		const limit = %d;
		const truncated = text.length > limit;
		return {
			text: truncated ? text.slice(0, limit) : text,
			text_truncated: truncated,
			text_length: text.length,
			text_limit: limit
		};
	})()`, maxCapturedTextLength)
}
