package tabs

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	maxSemanticHeadings = 8
	maxSemanticLinks    = 8
	maxSemanticForms    = 4
)

type SemanticHeading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
}

type SemanticLink struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type SemanticForm struct {
	Action     string `json:"action"`
	Method     string `json:"method"`
	InputCount int    `json:"input_count"`
}

type SemanticContext struct {
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	URL              string            `json:"url"`
	CapturedAt       string            `json:"captured_at"`
	Returned         int               `json:"returned"`
	HeadingsReturned int               `json:"headings_returned"`
	LinksReturned    int               `json:"links_returned"`
	FormsReturned    int               `json:"forms_returned"`
	Headings         []SemanticHeading `json:"headings"`
	Links            []SemanticLink    `json:"links"`
	Forms            []SemanticForm    `json:"forms"`
	CaptureError     string            `json:"capture_error"`
}

type SemanticCaptureError struct {
	Context SemanticContext
	Cause   error
}

func (e *SemanticCaptureError) Error() string {
	return e.Cause.Error()
}

func (e *SemanticCaptureError) Unwrap() error {
	return e.Cause
}

type semanticEvaluateResult struct {
	Result semanticRemoteObject `json:"result"`
}

type semanticRemoteObject struct {
	Type  string          `json:"type"`
	Value semanticPayload `json:"value"`
}

type semanticPayload struct {
	Headings []SemanticHeading `json:"headings"`
	Links    []SemanticLink    `json:"links"`
	Forms    []SemanticForm    `json:"forms"`
}

func (c Client) CaptureSemanticContext(targetID string) (SemanticContext, error) {
	targets, err := c.List()
	if err != nil {
		return SemanticContext{}, err
	}

	target, err := findPageTarget(targets, targetID)
	if err != nil {
		return SemanticContext{}, err
	}

	context := SemanticContext{
		ID:         target.ID,
		Title:      target.Title,
		URL:        target.URL,
		CapturedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Headings:   []SemanticHeading{},
		Links:      []SemanticLink{},
		Forms:      []SemanticForm{},
	}
	if target.WebSocketDebuggerURL == "" {
		return captureSemanticFailure(context, errors.New("target does not expose a websocket debugger url"))
	}

	payload, err := captureSemanticPayload(target.WebSocketDebuggerURL)
	if err != nil {
		return captureSemanticFailure(context, err)
	}

	context.Headings = payload.Headings
	context.Links = payload.Links
	context.Forms = payload.Forms
	context.HeadingsReturned = len(payload.Headings)
	context.LinksReturned = len(payload.Links)
	context.FormsReturned = len(payload.Forms)
	context.Returned = context.HeadingsReturned + context.LinksReturned + context.FormsReturned
	return context, nil
}

func captureSemanticFailure(context SemanticContext, err error) (SemanticContext, error) {
	context.CaptureError = err.Error()
	if context.Headings == nil {
		context.Headings = []SemanticHeading{}
	}
	if context.Links == nil {
		context.Links = []SemanticLink{}
	}
	if context.Forms == nil {
		context.Forms = []SemanticForm{}
	}
	return context, &SemanticCaptureError{
		Context: context,
		Cause:   err,
	}
}

func captureSemanticPayload(websocketURL string) (semanticPayload, error) {
	response, err := runPageCommand(websocketURL, cdpCommandRequest{
		ID:     3,
		Method: "Runtime.evaluate",
		Params: map[string]any{
			"expression":    captureSemanticExpression(),
			"returnByValue": true,
		},
	})
	if err != nil {
		return semanticPayload{}, err
	}

	var payload semanticEvaluateResult
	if err := json.Unmarshal(response.Result, &payload); err != nil {
		return semanticPayload{}, err
	}
	if payload.Result.Type != "object" {
		return semanticPayload{}, fmt.Errorf("unexpected runtime result type %q", payload.Result.Type)
	}
	if payload.Result.Value.Headings == nil {
		payload.Result.Value.Headings = []SemanticHeading{}
	}
	if payload.Result.Value.Links == nil {
		payload.Result.Value.Links = []SemanticLink{}
	}
	if payload.Result.Value.Forms == nil {
		payload.Result.Value.Forms = []SemanticForm{}
	}
	return payload.Result.Value, nil
}

func captureSemanticExpression() string {
	return fmt.Sprintf(`(() => {
		const normalize = (value) => (value || "").replace(/\s+/g, " ").trim();
		const headings = Array.from(document.querySelectorAll("h1,h2,h3"))
			.map((node) => ({
				level: Number((node.tagName || "H1").replace("H", "")) || 1,
				text: normalize(node.innerText || node.textContent || "")
			}))
			.filter((item) => item.text.length > 0)
			.slice(0, %d);
		const links = Array.from(document.querySelectorAll("a[href]"))
			.map((node) => ({
				text: normalize(node.innerText || node.textContent || node.getAttribute("aria-label") || node.getAttribute("title") || node.href || ""),
				url: normalize(node.href || "")
			}))
			.filter((item) => item.url.length > 0)
			.slice(0, %d);
		const forms = Array.from(document.forms || [])
			.map((form) => ({
				action: normalize(form.action || ""),
				method: normalize((form.method || "GET").toUpperCase()),
				input_count: form.querySelectorAll("input, textarea, select, button").length
			}))
			.slice(0, %d);
		return {
			headings,
			links,
			forms
		};
	})()`, maxSemanticHeadings, maxSemanticLinks, maxSemanticForms)
}
