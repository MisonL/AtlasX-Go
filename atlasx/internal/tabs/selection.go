package tabs

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const maxCapturedSelectionLength = 1024

type SelectionContext struct {
	ID                     string `json:"id"`
	Title                  string `json:"title"`
	URL                    string `json:"url"`
	SelectionText          string `json:"selection_text"`
	CapturedAt             string `json:"captured_at"`
	SelectionPresent       bool   `json:"selection_present"`
	SelectionTextTruncated bool   `json:"selection_text_truncated"`
	SelectionTextLength    int    `json:"selection_text_length"`
	SelectionTextLimit     int    `json:"selection_text_limit"`
	CaptureError           string `json:"capture_error"`
}

type SelectionCaptureError struct {
	Context SelectionContext
	Cause   error
}

func (e *SelectionCaptureError) Error() string {
	return e.Cause.Error()
}

func (e *SelectionCaptureError) Unwrap() error {
	return e.Cause
}

type selectionEvaluateResult struct {
	Result selectionRemoteObject `json:"result"`
}

type selectionRemoteObject struct {
	Type  string              `json:"type"`
	Value selectionTextResult `json:"value"`
}

type selectionTextResult struct {
	SelectionText          string `json:"selection_text"`
	SelectionPresent       bool   `json:"selection_present"`
	SelectionTextTruncated bool   `json:"selection_text_truncated"`
	SelectionTextLength    int    `json:"selection_text_length"`
	SelectionTextLimit     int    `json:"selection_text_limit"`
}

func (c Client) CaptureSelection(targetID string) (SelectionContext, error) {
	targets, err := c.List()
	if err != nil {
		return SelectionContext{}, err
	}

	target, err := findPageTarget(targets, targetID)
	if err != nil {
		return SelectionContext{}, err
	}

	context := SelectionContext{
		ID:                 target.ID,
		Title:              target.Title,
		URL:                target.URL,
		CapturedAt:         time.Now().UTC().Format(time.RFC3339Nano),
		SelectionTextLimit: maxCapturedSelectionLength,
	}
	if target.WebSocketDebuggerURL == "" {
		return captureSelectionFailure(context, errors.New("target does not expose a websocket debugger url"))
	}

	result, err := capturePageSelection(target.WebSocketDebuggerURL)
	if err != nil {
		return captureSelectionFailure(context, err)
	}

	context.SelectionText = result.SelectionText
	context.SelectionPresent = result.SelectionPresent
	context.SelectionTextTruncated = result.SelectionTextTruncated
	context.SelectionTextLength = result.SelectionTextLength
	if result.SelectionTextLimit > 0 {
		context.SelectionTextLimit = result.SelectionTextLimit
	}
	return context, nil
}

func captureSelectionFailure(context SelectionContext, err error) (SelectionContext, error) {
	context.CaptureError = err.Error()
	return context, &SelectionCaptureError{
		Context: context,
		Cause:   err,
	}
}

func capturePageSelection(websocketURL string) (selectionTextResult, error) {
	response, err := runPageCommand(websocketURL, cdpCommandRequest{
		ID:     2,
		Method: "Runtime.evaluate",
		Params: map[string]any{
			"expression":    captureSelectionExpression(),
			"returnByValue": true,
		},
	})
	if err != nil {
		return selectionTextResult{}, err
	}

	var payload selectionEvaluateResult
	if err := json.Unmarshal(response.Result, &payload); err != nil {
		return selectionTextResult{}, err
	}
	if payload.Result.Type != "object" {
		return selectionTextResult{}, fmt.Errorf("unexpected runtime result type %q", payload.Result.Type)
	}
	if payload.Result.Value.SelectionTextLimit == 0 {
		payload.Result.Value.SelectionTextLimit = maxCapturedSelectionLength
	}
	return payload.Result.Value, nil
}

func captureSelectionExpression() string {
	return fmt.Sprintf(`(() => {
		const normalize = (value) => (value || "").replace(/\s+/g, " ").trim();
		const limit = %d;
		const selection = window.getSelection ? normalize(window.getSelection().toString()) : "";
		let text = selection;
		if (!text) {
			const active = document.activeElement;
			if (active && typeof active.value === "string" && typeof active.selectionStart === "number" && typeof active.selectionEnd === "number") {
				const start = Math.min(active.selectionStart, active.selectionEnd);
				const end = Math.max(active.selectionStart, active.selectionEnd);
				text = normalize(active.value.slice(start, end));
			}
		}
		const truncated = text.length > limit;
		return {
			selection_text: truncated ? text.slice(0, limit) : text,
			selection_present: text.length > 0,
			selection_text_truncated: truncated,
			selection_text_length: text.length,
			selection_text_limit: limit
		};
	})()`, maxCapturedSelectionLength)
}
