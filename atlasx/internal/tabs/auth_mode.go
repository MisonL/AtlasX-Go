package tabs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const maxAuthStorageKeys = 16

type AuthModeView struct {
	ID                     string   `json:"id"`
	Title                  string   `json:"title"`
	URL                    string   `json:"url"`
	CapturedAt             string   `json:"captured_at"`
	Host                   string   `json:"host"`
	Path                   string   `json:"path"`
	Mode                   string   `json:"mode"`
	Inferred               bool     `json:"inferred"`
	Reason                 string   `json:"reason"`
	LoginPromptPresent     bool     `json:"login_prompt_present"`
	WorkspaceSignalPresent bool     `json:"workspace_signal_present"`
	CookieCount            int      `json:"cookie_count"`
	CookieNames            []string `json:"cookie_names"`
	LocalStorageCount      int      `json:"local_storage_count"`
	LocalStorageKeys       []string `json:"local_storage_keys"`
	SessionStorageCount    int      `json:"session_storage_count"`
	SessionStorageKeys     []string `json:"session_storage_keys"`
}

type authStorageSignals struct {
	CookieCount         int      `json:"cookie_count"`
	CookieNames         []string `json:"cookie_names"`
	LocalStorageCount   int      `json:"local_storage_count"`
	LocalStorageKeys    []string `json:"local_storage_keys"`
	SessionStorageCount int      `json:"session_storage_count"`
	SessionStorageKeys  []string `json:"session_storage_keys"`
}

type authModeEvaluateResult struct {
	Result authModeRemoteObject `json:"result"`
}

type authModeRemoteObject struct {
	Type  string             `json:"type"`
	Value authStorageSignals `json:"value"`
}

func (c Client) AuthMode(targetID string) (AuthModeView, error) {
	context, target, err := c.captureAuthContext(targetID)
	if err != nil {
		return AuthModeView{}, err
	}

	signals, err := captureAuthSignals(target.WebSocketDebuggerURL)
	if err != nil {
		return AuthModeView{}, err
	}

	view := AuthModeView{
		ID:                  context.ID,
		Title:               context.Title,
		URL:                 context.URL,
		CapturedAt:          context.CapturedAt,
		Inferred:            true,
		CookieCount:         signals.CookieCount,
		CookieNames:         append([]string(nil), signals.CookieNames...),
		LocalStorageCount:   signals.LocalStorageCount,
		LocalStorageKeys:    append([]string(nil), signals.LocalStorageKeys...),
		SessionStorageCount: signals.SessionStorageCount,
		SessionStorageKeys:  append([]string(nil), signals.SessionStorageKeys...),
	}

	parsedURL, err := url.Parse(context.URL)
	if err == nil {
		view.Host = strings.ToLower(parsedURL.Hostname())
		view.Path = parsedURL.EscapedPath()
		if view.Path == "" {
			view.Path = "/"
		}
	}

	view.LoginPromptPresent = loginPromptPresent(view.Path, context.Text)
	view.WorkspaceSignalPresent = workspaceSignalPresent(view.Path, context.Text)
	view.Mode, view.Reason = inferAuthMode(view)
	return view, nil
}

func (c Client) captureAuthContext(targetID string) (PageContext, Target, error) {
	targets, err := c.List()
	if err != nil {
		return PageContext{}, Target{}, err
	}

	target, err := findPageTarget(targets, targetID)
	if err != nil {
		return PageContext{}, Target{}, err
	}

	context, err := c.Capture(targetID)
	if err != nil {
		return PageContext{}, Target{}, err
	}
	return context, target, nil
}

func captureAuthSignals(websocketURL string) (authStorageSignals, error) {
	response, err := runPageCommand(websocketURL, cdpCommandRequest{
		ID:     7,
		Method: "Runtime.evaluate",
		Params: map[string]any{
			"expression":    authSignalsExpression(),
			"returnByValue": true,
		},
	})
	if err != nil {
		return authStorageSignals{}, err
	}

	var payload authModeEvaluateResult
	if err := json.Unmarshal(response.Result, &payload); err != nil {
		return authStorageSignals{}, err
	}
	if payload.Result.Type != "object" {
		return authStorageSignals{}, fmt.Errorf("unexpected runtime result type %q", payload.Result.Type)
	}
	return payload.Result.Value, nil
}

func authSignalsExpression() string {
	return fmt.Sprintf(`(() => {
		const limit = %d;
		const readKeys = (storage) => {
			try {
				const keys = [];
				for (let index = 0; index < storage.length; index += 1) {
					const key = storage.key(index);
					if (typeof key === "string" && key.trim() !== "") {
						keys.push(key.trim());
					}
				}
				return keys.slice(0, limit);
			} catch (error) {
				return [];
			}
		};
		const readCount = (storage) => {
			try {
				return storage.length;
			} catch (error) {
				return 0;
			}
		};
		const cookieNames = document.cookie
			.split(";")
			.map((entry) => entry.split("=")[0].trim())
			.filter((entry) => entry.length > 0);
		const localStorageKeys = readKeys(window.localStorage);
		const sessionStorageKeys = readKeys(window.sessionStorage);
		return {
			cookie_count: cookieNames.length,
			cookie_names: cookieNames.slice(0, limit),
			local_storage_count: readCount(window.localStorage),
			local_storage_keys: localStorageKeys,
			session_storage_count: readCount(window.sessionStorage),
			session_storage_keys: sessionStorageKeys
		};
	})()`, maxAuthStorageKeys)
}

func inferAuthMode(view AuthModeView) (string, string) {
	if !isAtlasAuthHost(view.Host) {
		return "unknown", "host_outside_auth_scope"
	}
	if view.WorkspaceSignalPresent && !view.LoginPromptPresent {
		return "logged_in", "workspace_signals_observed"
	}
	if view.LoginPromptPresent && !view.WorkspaceSignalPresent {
		return "logged_out", "login_prompts_observed"
	}
	if view.CookieCount == 0 && view.LocalStorageCount == 0 && view.SessionStorageCount == 0 {
		return "logged_out", "no_auth_artifacts_observed"
	}
	return "unknown", "mixed_or_insufficient_signals"
}

func isAtlasAuthHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	switch {
	case host == "chatgpt.com":
		return true
	case strings.HasSuffix(host, ".chatgpt.com"):
		return true
	case host == "chat.openai.com":
		return true
	case host == "openai.com":
		return true
	case strings.HasSuffix(host, ".openai.com"):
		return true
	default:
		return false
	}
}

func loginPromptPresent(path string, text string) bool {
	if strings.HasPrefix(path, "/auth/") || strings.HasPrefix(path, "/login") || strings.HasPrefix(path, "/logout") || strings.HasPrefix(path, "/logged-out") {
		return true
	}
	return containsAnyFold(text,
		"log in",
		"sign up",
		"continue with google",
		"continue with apple",
		"continue with microsoft",
		"continue with email",
	)
}

func workspaceSignalPresent(path string, text string) bool {
	if strings.HasPrefix(path, "/c/") || strings.HasPrefix(path, "/g/") || strings.HasPrefix(path, "/projects") || strings.HasPrefix(path, "/library") || strings.HasPrefix(path, "/gpts") {
		return true
	}
	return containsAnyFold(text,
		"new chat",
		"projects",
		"gpts",
		"library",
		"temporary chat",
	)
}

func containsAnyFold(text string, patterns ...string) bool {
	lowered := strings.ToLower(text)
	for _, pattern := range patterns {
		if strings.Contains(lowered, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}
