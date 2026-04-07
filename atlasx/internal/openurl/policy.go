package openurl

import (
	"fmt"
	"net/url"
	"strings"
)

func Validate(targetURL string) (string, error) {
	trimmed := strings.TrimSpace(targetURL)
	if trimmed == "" {
		return "", fmt.Errorf("url must not be empty")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid url %q: %w", trimmed, err)
	}
	if !parsed.IsAbs() {
		return "", fmt.Errorf("url must be absolute: %q", trimmed)
	}

	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
		if parsed.Host == "" {
			return "", fmt.Errorf("url host must not be empty: %q", trimmed)
		}
		return trimmed, nil
	default:
		return "", fmt.Errorf("unsupported url scheme %q", parsed.Scheme)
	}
}
