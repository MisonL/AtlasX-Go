package launcher

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultDevToolsHost     = "127.0.0.1"
	devToolsActivePortFile  = "DevToolsActivePort"
	cdpProbeRequestTimeout  = 1500 * time.Millisecond
	cdpReadyPollInterval    = 100 * time.Millisecond
	cdpStatusOK             = "ok"
	cdpStatusSessionAbsent  = "session_absent"
	cdpStatusSessionDead    = "session_not_alive"
	cdpStatusNotManaged     = "session_not_managed"
	cdpStatusPortFileAbsent = "devtools_active_port_missing"
	cdpStatusPortFileBad    = "devtools_active_port_invalid"
	cdpStatusVersionDown    = "devtools_version_unreachable"
)

type CDPReport struct {
	Status               string
	ActivePortFile       string
	Host                 string
	Port                 int
	VersionEndpoint      string
	BrowserWebSocketURL  string
	BrowserWebSocketPath string
	LastError            string
}

type versionResponse struct {
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
}

func ProbeCDP(userDataDir string) CDPReport {
	report := CDPReport{
		Status:         cdpStatusPortFileAbsent,
		ActivePortFile: filepath.Join(userDataDir, devToolsActivePortFile),
		Host:           defaultDevToolsHost,
	}

	lines, err := readDevToolsActivePort(report.ActivePortFile)
	if err != nil {
		report.LastError = err.Error()
		return report
	}
	if len(lines) < 1 {
		report.Status = cdpStatusPortFileBad
		report.LastError = "missing port line"
		return report
	}

	port, err := strconv.Atoi(lines[0])
	if err != nil {
		report.Status = cdpStatusPortFileBad
		report.LastError = err.Error()
		return report
	}

	report.Port = port
	report.VersionEndpoint = fmt.Sprintf("http://%s:%d/json/version", report.Host, report.Port)
	if len(lines) > 1 {
		report.BrowserWebSocketPath = lines[1]
	}

	if err := populateVersion(report.VersionEndpoint, &report); err != nil {
		report.Status = cdpStatusVersionDown
		report.LastError = err.Error()
		if report.BrowserWebSocketPath != "" {
			report.BrowserWebSocketURL = fmt.Sprintf("ws://%s:%d%s", report.Host, report.Port, report.BrowserWebSocketPath)
		}
		return report
	}

	report.Status = cdpStatusOK
	return report
}

func readDevToolsActivePort(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rawLines := strings.Split(strings.TrimSpace(string(data)), "\n")
	lines := make([]string, 0, len(rawLines))
	for _, line := range rawLines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			lines = append(lines, trimmed)
		}
	}
	return lines, nil
}

func populateVersion(endpoint string, report *CDPReport) error {
	client := http.Client{Timeout: cdpProbeRequestTimeout}
	response, err := client.Get(endpoint)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", response.StatusCode)
	}

	var payload versionResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return err
	}

	report.BrowserWebSocketURL = payload.WebSocketDebuggerURL
	if report.BrowserWebSocketURL == "" && report.BrowserWebSocketPath != "" {
		report.BrowserWebSocketURL = fmt.Sprintf("ws://%s:%d%s", report.Host, report.Port, report.BrowserWebSocketPath)
	}
	return nil
}
