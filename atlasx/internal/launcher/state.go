package launcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"atlasx/internal/platform/macos"
)

var ErrStateNotFound = errors.New("managed session state not found")

const (
	sessionPollInterval = 100 * time.Millisecond
	forceStopWait       = 2 * time.Second
)

type State struct {
	Mode          string   `json:"mode"`
	Managed       bool     `json:"managed"`
	RuntimeSource string   `json:"runtime_source"`
	BinaryPath    string   `json:"binary_path"`
	Args          []string `json:"args"`
	URL           string   `json:"url"`
	UserDataDir   string   `json:"user_data_dir"`
	StartedAt     string   `json:"started_at"`
}

type StatusReport struct {
	Present    bool
	StateFile  string
	State      State
	Alive      bool
	ProcessIDs []int
	CDP        CDPReport
}

func SaveState(paths macos.Paths, state State) error {
	if err := macos.EnsureDir(paths.StateRoot); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(paths.SessionFile, append(data, '\n'), 0o644)
}

func LoadState(paths macos.Paths) (State, error) {
	data, err := os.ReadFile(paths.SessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, ErrStateNotFound
		}
		return State{}, err
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, err
	}
	return state, nil
}

func ClearState(paths macos.Paths) error {
	if err := os.Remove(paths.SessionFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func Status(paths macos.Paths) (StatusReport, error) {
	state, err := LoadState(paths)
	if err != nil {
		if errors.Is(err, ErrStateNotFound) {
			return StatusReport{
				StateFile: paths.SessionFile,
				CDP:       CDPReport{Status: cdpStatusSessionAbsent},
			}, nil
		}
		return StatusReport{}, err
	}

	report := StatusReport{
		Present:   true,
		StateFile: paths.SessionFile,
		State:     state,
	}
	if !state.Managed || state.UserDataDir == "" {
		report.CDP = CDPReport{Status: cdpStatusNotManaged}
		return report, nil
	}

	processes, err := macos.FindProcessesByUserDataDir(state.UserDataDir)
	if err != nil {
		return StatusReport{}, err
	}

	report.ProcessIDs = make([]int, 0, len(processes))
	for _, process := range processes {
		report.ProcessIDs = append(report.ProcessIDs, process.PID)
	}
	report.Alive = len(report.ProcessIDs) > 0
	if report.Alive {
		report.CDP = ProbeCDP(state.UserDataDir)
	} else {
		report.CDP = CDPReport{
			Status:         cdpStatusSessionDead,
			ActivePortFile: filepath.Join(state.UserDataDir, devToolsActivePortFile),
			Host:           defaultDevToolsHost,
		}
	}
	return report, nil
}

func Stop(paths macos.Paths, wait time.Duration) (StatusReport, error) {
	report, err := Status(paths)
	if err != nil {
		return StatusReport{}, err
	}
	if !report.Present {
		return StatusReport{}, ErrStateNotFound
	}
	if !report.State.Managed {
		return StatusReport{}, errors.New("shared profile session is not managed by AtlasX")
	}
	if !report.Alive {
		if err := ClearState(paths); err != nil {
			return StatusReport{}, err
		}
		return report, nil
	}

	for _, pid := range report.ProcessIDs {
		if err := terminatePID(pid); err != nil {
			return StatusReport{}, err
		}
	}

	current, err := waitForSessionExit(paths, wait)
	if err != nil {
		return StatusReport{}, err
	}
	if !current.Alive {
		if err := ClearState(paths); err != nil {
			return StatusReport{}, err
		}
		return current, nil
	}

	for _, pid := range current.ProcessIDs {
		if err := killPID(pid); err != nil {
			return StatusReport{}, err
		}
	}

	current, err = waitForSessionExit(paths, forceStopWait)
	if err != nil {
		return StatusReport{}, err
	}
	if !current.Alive {
		if err := ClearState(paths); err != nil {
			return StatusReport{}, err
		}
		return current, nil
	}

	return StatusReport{}, fmt.Errorf("managed session still alive after graceful and forced stop windows")
}

func (s StatusReport) Render() string {
	if !s.Present {
		return fmt.Sprintf(
			"session=absent\nstate_file=%s\ncdp_status=%s\n",
			s.StateFile,
			s.CDP.Status,
		)
	}

	return fmt.Sprintf(
		"session=present\nstate_file=%s\nmanaged=%t\nalive=%t\nmode=%s\nruntime_source=%s\nuser_data_dir=%s\nurl=%s\npids=%v\ncdp_status=%s\ncdp_version_endpoint=%s\ncdp_browser_ws=%s\ncdp_active_port_file=%s\n",
		s.StateFile,
		s.State.Managed,
		s.Alive,
		s.State.Mode,
		s.State.RuntimeSource,
		s.State.UserDataDir,
		s.State.URL,
		s.ProcessIDs,
		s.CDP.Status,
		s.CDP.VersionEndpoint,
		s.CDP.BrowserWebSocketURL,
		s.CDP.ActivePortFile,
	)
}

func waitForSessionExit(paths macos.Paths, timeout time.Duration) (StatusReport, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		current, err := Status(paths)
		if err != nil {
			return StatusReport{}, err
		}
		if !current.Alive {
			return current, nil
		}
		time.Sleep(sessionPollInterval)
	}
	return Status(paths)
}
