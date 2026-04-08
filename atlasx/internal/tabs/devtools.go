package tabs

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var devToolsPanelPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

type DevToolsTarget struct {
	ID                  string `json:"id"`
	Title               string `json:"title"`
	URL                 string `json:"url"`
	DevToolsFrontendURL string `json:"devtools_frontend_url"`
}

func (c Client) DevTools(targetID string) (DevToolsTarget, error) {
	targets, err := c.List()
	if err != nil {
		return DevToolsTarget{}, err
	}

	target, err := findPageTarget(targets, targetID)
	if err != nil {
		return DevToolsTarget{}, err
	}

	devToolsURL, err := resolveDevToolsFrontendURL(c.baseURL, target.DevToolsFrontendURL)
	if err != nil {
		return DevToolsTarget{}, err
	}

	return DevToolsTarget{
		ID:                  target.ID,
		Title:               target.Title,
		URL:                 target.URL,
		DevToolsFrontendURL: devToolsURL,
	}, nil
}

func resolveDevToolsFrontendURL(baseURL string, devToolsURL string) (string, error) {
	if devToolsURL == "" {
		return "", fmt.Errorf("target does not expose a devtools frontend url")
	}

	parsedURL, err := url.Parse(devToolsURL)
	if err != nil {
		return "", err
	}
	if parsedURL.IsAbs() {
		return parsedURL.String(), nil
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(parsedURL).String(), nil
}

func resolveDevToolsPanelURL(devToolsURL string, panel string) (string, error) {
	panel = strings.TrimSpace(panel)
	if panel == "" {
		return "", fmt.Errorf("panel is required")
	}
	if !devToolsPanelPattern.MatchString(panel) {
		return "", fmt.Errorf("invalid devtools panel %q", panel)
	}

	parsedURL, err := url.Parse(devToolsURL)
	if err != nil {
		return "", err
	}

	query := parsedURL.Query()
	query.Set("panel", panel)
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}
