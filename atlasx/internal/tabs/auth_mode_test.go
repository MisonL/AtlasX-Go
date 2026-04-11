package tabs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestInferAuthModeReturnsLoggedInForWorkspaceSignals(t *testing.T) {
	view := AuthModeView{
		Host:                   "chatgpt.com",
		Path:                   "/c/abc123",
		WorkspaceSignalPresent: true,
		CookieCount:            1,
	}

	mode, reason := inferAuthMode(view)
	if mode != "logged_in" || reason != "workspace_signals_observed" {
		t.Fatalf("unexpected inference: mode=%s reason=%s", mode, reason)
	}
}

func TestInferAuthModeReturnsLoggedOutForLoginSignals(t *testing.T) {
	view := AuthModeView{
		Host:               "chatgpt.com",
		Path:               "/auth/login",
		LoginPromptPresent: true,
	}

	mode, reason := inferAuthMode(view)
	if mode != "logged_out" || reason != "login_prompts_observed" {
		t.Fatalf("unexpected inference: mode=%s reason=%s", mode, reason)
	}
}

func TestInferAuthModeReturnsUnknownOutsideScope(t *testing.T) {
	view := AuthModeView{
		Host: "example.com",
		Path: "/",
	}

	mode, reason := inferAuthMode(view)
	if mode != "unknown" || reason != "host_outside_auth_scope" {
		t.Fatalf("unexpected inference: mode=%s reason=%s", mode, reason)
	}
}

func TestAuthModeReturnsLoggedOutFromPromptAndStorageSignals(t *testing.T) {
	upgrader := websocket.Upgrader{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/devtools/page/tab-1"

	mux.HandleFunc("/json/list", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"tab-1","type":"page","title":"ChatGPT","url":"https://chatgpt.com/auth/login","webSocketDebuggerUrl":"` + wsURL + `"}]`))
	})
	mux.HandleFunc("/devtools/page/tab-1", func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade failed: %v", err)
		}
		defer func() {
			_ = connection.Close()
		}()

		var request cdpCommandRequest
		if err := connection.ReadJSON(&request); err != nil {
			t.Fatalf("read request failed: %v", err)
		}
		if request.Method != "Runtime.evaluate" {
			t.Fatalf("unexpected method: %s", request.Method)
		}
		expression, _ := request.Params["expression"].(string)
		switch {
		case strings.Contains(expression, "document.body ? (document.body.innerText || \"\") : \"\""):
			if err := connection.WriteJSON(cdpCommandResponse{
				ID: request.ID,
				Result: mustMarshalJSON(t, runtimeEvaluateResult{
					Result: runtimeRemoteObject{
						Type: "object",
						Value: captureTextResult{
							Text:          "Log in Sign up Continue with Google",
							TextTruncated: false,
							TextLength:    36,
							TextLimit:     maxCapturedTextLength,
						},
					},
				}),
			}); err != nil {
				t.Fatalf("write response failed: %v", err)
			}
		case strings.Contains(expression, "cookie_count"):
			if err := connection.WriteJSON(cdpCommandResponse{
				ID: request.ID,
				Result: mustMarshalJSON(t, authModeEvaluateResult{
					Result: authModeRemoteObject{
						Type: "object",
						Value: authStorageSignals{
							CookieCount:         0,
							CookieNames:         []string{},
							LocalStorageCount:   0,
							LocalStorageKeys:    []string{},
							SessionStorageCount: 0,
							SessionStorageKeys:  []string{},
						},
					},
				}),
			}); err != nil {
				t.Fatalf("write response failed: %v", err)
			}
		default:
			t.Fatalf("unexpected expression: %s", expression)
		}
	})

	client := Client{baseURL: server.URL, httpClient: *server.Client()}
	view, err := client.AuthMode("tab-1")
	if err != nil {
		t.Fatalf("auth mode failed: %v", err)
	}
	if view.Mode != "logged_out" || !view.Inferred {
		t.Fatalf("unexpected auth mode view: %+v", view)
	}
	if view.Reason != "login_prompts_observed" {
		t.Fatalf("unexpected reason: %+v", view)
	}
	if view.Host != "chatgpt.com" || view.Path != "/auth/login" {
		t.Fatalf("unexpected host/path: %+v", view)
	}
	if view.URL != "https://chatgpt.com/auth/login" {
		t.Fatalf("unexpected url: %+v", view)
	}
	if view.LoginPromptPresent != true || view.WorkspaceSignalPresent {
		t.Fatalf("unexpected prompt flags: %+v", view)
	}
	if view.CookieCount != 0 || view.LocalStorageCount != 0 || view.SessionStorageCount != 0 {
		t.Fatalf("unexpected storage counts: %+v", view)
	}
}
