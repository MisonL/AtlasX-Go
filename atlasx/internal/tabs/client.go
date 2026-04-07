package tabs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"atlasx/internal/launcher"
	"atlasx/internal/openurl"
	"atlasx/internal/platform/macos"
)

const requestTimeout = 2 * time.Second

type Target struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	DevToolsFrontendURL  string `json:"devtoolsFrontendUrl"`
}

type Client struct {
	baseURL             string
	browserWebSocketURL string
	httpClient          http.Client
}

type cdpCommandRequest struct {
	ID     int            `json:"id"`
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

type cdpCommandResponse struct {
	ID     int             `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *cdpError       `json:"error,omitempty"`
}

type cdpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type versionResponse struct {
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func New(paths macos.Paths) (Client, error) {
	status, err := launcher.Status(paths)
	if err != nil {
		return Client{}, err
	}
	if !status.Present {
		return Client{}, errors.New("no managed browser session")
	}
	if !status.Alive {
		return Client{}, errors.New("managed browser session is not alive")
	}
	if status.CDP.Status != "ok" {
		return Client{}, fmt.Errorf("managed browser cdp is not ready: %s", status.CDP.Status)
	}

	baseURL, err := baseURLFromVersionEndpoint(status.CDP.VersionEndpoint)
	if err != nil {
		return Client{}, err
	}

	return Client{
		baseURL:             baseURL,
		browserWebSocketURL: status.CDP.BrowserWebSocketURL,
		httpClient: http.Client{
			Timeout: requestTimeout,
		},
	}, nil
}

func (c Client) List() ([]Target, error) {
	response, err := c.httpClient.Get(c.baseURL + "/json/list")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", response.StatusCode)
	}

	var targets []Target
	if err := json.NewDecoder(response.Body).Decode(&targets); err != nil {
		return nil, err
	}
	return targets, nil
}

func PageTargets(targets []Target) []Target {
	pages := make([]Target, 0, len(targets))
	for _, target := range targets {
		if target.Type == "page" {
			pages = append(pages, target)
		}
	}
	return pages
}

func (c Client) Open(targetURL string) (Target, error) {
	targetURL, err := openurl.Validate(targetURL)
	if err != nil {
		return Target{}, err
	}

	request, err := http.NewRequest(http.MethodPut, c.baseURL+"/json/new?"+url.QueryEscape(targetURL), nil)
	if err != nil {
		return Target{}, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Target{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return Target{}, unexpectedStatus(response)
	}

	var target Target
	if err := json.NewDecoder(response.Body).Decode(&target); err != nil {
		return Target{}, err
	}
	return target, nil
}

type createTargetResult struct {
	TargetID string `json:"targetId"`
}

func (c Client) OpenWindow(targetURL string) (Target, error) {
	targetURL, err := openurl.Validate(targetURL)
	if err != nil {
		return Target{}, err
	}

	browserWS, err := c.browserWS()
	if err != nil {
		return Target{}, err
	}

	response, err := runPageCommand(browserWS, cdpCommandRequest{
		ID:     20,
		Method: "Target.createTarget",
		Params: map[string]any{
			"url":       targetURL,
			"newWindow": true,
		},
	})
	if err != nil {
		return Target{}, err
	}

	var payload createTargetResult
	if err := json.Unmarshal(response.Result, &payload); err != nil {
		return Target{}, err
	}
	if payload.TargetID == "" {
		return Target{}, errors.New("missing targetId in createTarget response")
	}

	targets, err := c.List()
	if err != nil {
		return Target{}, err
	}
	for _, target := range targets {
		if target.ID == payload.TargetID {
			return target, nil
		}
	}
	return Target{}, fmt.Errorf("target %s not found after open-window", payload.TargetID)
}

func (c Client) Activate(targetID string) error {
	request, err := http.NewRequest(http.MethodGet, c.baseURL+"/json/activate/"+targetID, nil)
	if err != nil {
		return err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return unexpectedStatus(response)
	}
	return nil
}

func (c Client) Close(targetID string) error {
	request, err := http.NewRequest(http.MethodGet, c.baseURL+"/json/close/"+targetID, nil)
	if err != nil {
		return err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return unexpectedStatus(response)
	}
	return nil
}

func (c Client) Navigate(targetID string, targetURL string) error {
	targetURL, err := openurl.Validate(targetURL)
	if err != nil {
		return err
	}

	targets, err := c.List()
	if err != nil {
		return err
	}

	target, err := findPageTarget(targets, targetID)
	if err != nil {
		return err
	}
	if target.WebSocketDebuggerURL == "" {
		return errors.New("target does not expose a websocket debugger url")
	}

	request := cdpCommandRequest{
		ID:     1,
		Method: "Page.navigate",
		Params: map[string]any{
			"url": targetURL,
		},
	}

	_, err = runPageCommand(target.WebSocketDebuggerURL, request)
	return err
}

func baseURLFromVersionEndpoint(endpoint string) (string, error) {
	if endpoint == "" {
		return "", errors.New("empty cdp version endpoint")
	}
	if !strings.HasSuffix(endpoint, "/json/version") {
		return "", fmt.Errorf("unsupported cdp version endpoint %q", endpoint)
	}
	return strings.TrimSuffix(endpoint, "/json/version"), nil
}

func (c Client) browserWS() (string, error) {
	if strings.TrimSpace(c.browserWebSocketURL) != "" {
		return c.browserWebSocketURL, nil
	}

	response, err := c.httpClient.Get(c.baseURL + "/json/version")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", response.StatusCode)
	}

	var payload versionResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", err
	}
	if strings.TrimSpace(payload.WebSocketDebuggerURL) == "" {
		return "", errors.New("browser websocket debugger url is not available")
	}
	return payload.WebSocketDebuggerURL, nil
}

func unexpectedStatus(response *http.Response) error {
	body, _ := io.ReadAll(response.Body)
	if len(body) == 0 {
		return fmt.Errorf("unexpected status %d", response.StatusCode)
	}
	return fmt.Errorf("unexpected status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
}

func findPageTarget(targets []Target, targetID string) (Target, error) {
	for _, target := range targets {
		if target.ID != targetID {
			continue
		}
		if target.Type != "page" {
			return Target{}, fmt.Errorf("target %s is not a page", targetID)
		}
		return target, nil
	}
	return Target{}, fmt.Errorf("target %s not found", targetID)
}
