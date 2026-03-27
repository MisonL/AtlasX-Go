package daemon

import (
	"encoding/json"
	"net/http"

	"atlasx/internal/diagnostics"
	"atlasx/internal/settings"
)

const DefaultListenAddr = settings.DefaultListenAddr

type Status struct {
	Ready        bool   `json:"ready"`
	ChromeStatus string `json:"chrome_status"`
	SupportRoot  string `json:"support_root"`
	ConfigFile   string `json:"config_file"`
}

func Bootstrap() (Status, error) {
	report, err := diagnostics.Generate()
	if err != nil {
		return Status{}, err
	}
	return Status{
		Ready:        report.ChromeStatus == "ok",
		ChromeStatus: report.ChromeStatus,
		SupportRoot:  report.Paths.SupportRoot,
		ConfigFile:   report.Paths.ConfigFile,
	}, nil
}

func NewMux(status Status) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, status)
	})
	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, status)
	})
	return mux
}

func (s Status) Render() string {
	payload, _ := json.MarshalIndent(s, "", "  ")
	return string(payload) + "\n"
}

func writeJSON(w http.ResponseWriter, code int, payload Status) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
