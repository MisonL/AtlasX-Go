package sidebar

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"atlasx/internal/platform/macos"
)

type RuntimeState struct {
	TimeoutMS     int    `json:"timeout_ms"`
	RetryAttempts int    `json:"retry_attempts"`
	TokenBudget   int    `json:"token_budget"`
	LastTraceID   string `json:"last_trace_id"`
	LastError     string `json:"last_error"`
	LastErrorAt   string `json:"last_error_at"`
}

func LoadRuntimeState(paths macos.Paths) (RuntimeState, error) {
	data, err := os.ReadFile(runtimeStatePath(paths))
	if err != nil {
		if os.IsNotExist(err) {
			return defaultRuntimeState(), nil
		}
		return RuntimeState{}, err
	}

	var state RuntimeState
	if err := json.Unmarshal(data, &state); err != nil {
		return RuntimeState{}, err
	}
	if state.TimeoutMS == 0 {
		state.TimeoutMS = defaultProviderTimeoutMS()
	}
	if state.RetryAttempts == 0 {
		state.RetryAttempts = defaultProviderRetryAttempts()
	}
	if state.TokenBudget == 0 {
		state.TokenBudget = defaultProviderTokenBudget()
	}
	return state, nil
}

func SaveRuntimeResult(paths macos.Paths, traceID string, err error) error {
	state := defaultRuntimeState()
	state.LastTraceID = traceID
	if err != nil {
		state.LastError = err.Error()
		state.LastErrorAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if saveErr := macos.EnsureDir(paths.StateRoot); saveErr != nil {
		return saveErr
	}
	data, saveErr := json.MarshalIndent(state, "", "  ")
	if saveErr != nil {
		return saveErr
	}
	return os.WriteFile(runtimeStatePath(paths), append(data, '\n'), 0o644)
}

func NewTraceID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf)
}

func defaultRuntimeState() RuntimeState {
	return RuntimeState{
		TimeoutMS:     defaultProviderTimeoutMS(),
		RetryAttempts: defaultProviderRetryAttempts(),
		TokenBudget:   defaultProviderTokenBudget(),
	}
}

func runtimeStatePath(paths macos.Paths) string {
	return filepath.Join(paths.StateRoot, "sidebar-qa-status.json")
}
