package tabs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"atlasx/internal/launcher"
	"atlasx/internal/platform/macos"
)

const requestTimeout = 2 * time.Second

type Target struct {
	ID                   string `json:"id"`
	Type                 string `json:"type"`
	Title                string `json:"title"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

type Client struct {
	baseURL    string
	httpClient http.Client
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
		baseURL: baseURL,
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
		return Target{}, fmt.Errorf("unexpected status %d", response.StatusCode)
	}

	var target Target
	if err := json.NewDecoder(response.Body).Decode(&target); err != nil {
		return Target{}, err
	}
	return target, nil
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
