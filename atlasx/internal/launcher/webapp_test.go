package launcher

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"atlasx/internal/platform/macos"
	"atlasx/internal/profile"
)

func TestBuildArgsIncludesAppModeAndManagedProfile(t *testing.T) {
	args := BuildArgs(
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		profile.Selection{Mode: profile.ModeIsolated, UserDataDir: "/tmp/atlasx"},
		"https://chatgpt.com/atlas?get-started",
	)

	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--app=https://chatgpt.com/atlas?get-started") {
		t.Fatalf("missing app arg: %s", joined)
	}
	if !strings.Contains(joined, "--user-data-dir=/tmp/atlasx") {
		t.Fatalf("missing user data dir: %s", joined)
	}
	if !strings.Contains(joined, "--remote-debugging-port=0") {
		t.Fatalf("missing remote debugging port: %s", joined)
	}
}

func TestBuildArgsSkipsManagedProfileForSharedMode(t *testing.T) {
	args := BuildArgs(
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		profile.Selection{Mode: profile.ModeShared},
		"https://chatgpt.com/atlas?get-started",
	)

	joined := strings.Join(args, " ")
	if strings.Contains(joined, "--user-data-dir=") {
		t.Fatalf("unexpected user data dir for shared mode: %s", joined)
	}
	if strings.Contains(joined, "--remote-debugging-port=") {
		t.Fatalf("unexpected remote debugging arg for shared mode: %s", joined)
	}
}

func TestStatusAbsentWithoutStateFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	report, err := Status(paths)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if report.Present {
		t.Fatal("expected absent session")
	}
}

func TestStatusClearsStaleStateWhenProcessIsGone(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if err := SaveState(paths, State{
		Mode:        profile.ModeIsolated,
		Managed:     true,
		UserDataDir: filepath.Join(t.TempDir(), "profile"),
		StartedAt:   time.Now().Add(-10 * time.Second).UTC().Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("save state failed: %v", err)
	}

	previousFind := findProcessesByUserDataDir
	findProcessesByUserDataDir = func(string) ([]macos.ProcessInfo, error) {
		return nil, nil
	}
	t.Cleanup(func() {
		findProcessesByUserDataDir = previousFind
	})

	report, err := Status(paths)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if report.Present {
		t.Fatalf("expected stale state to be cleaned: %+v", report)
	}
	if !report.Stale || !report.StateCleaned {
		t.Fatalf("expected stale cleaned report: %+v", report)
	}
	if _, err := os.Stat(paths.SessionFile); !os.IsNotExist(err) {
		t.Fatalf("expected session file removed, got: %v", err)
	}
}

func TestStatusMarksOldManagedSessionStaleWhenCDPIsDown(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if err := SaveState(paths, State{
		Mode:        profile.ModeIsolated,
		Managed:     true,
		UserDataDir: filepath.Join(t.TempDir(), "profile"),
		StartedAt:   time.Now().Add(-10 * time.Second).UTC().Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("save state failed: %v", err)
	}

	previousFind := findProcessesByUserDataDir
	previousProbe := probeManagedCDP
	findProcessesByUserDataDir = func(string) ([]macos.ProcessInfo, error) {
		return []macos.ProcessInfo{{PID: 123}}, nil
	}
	probeManagedCDP = func(string) CDPReport {
		return CDPReport{Status: cdpStatusVersionDown, LastError: "timeout"}
	}
	t.Cleanup(func() {
		findProcessesByUserDataDir = previousFind
		probeManagedCDP = previousProbe
	})

	report, err := Status(paths)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !report.Alive || report.Ready || !report.Stale {
		t.Fatalf("unexpected stale session report: %+v", report)
	}
}

func TestStatusHealsCDPWhenRetrySucceeds(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	if err := SaveState(paths, State{
		Mode:        profile.ModeIsolated,
		Managed:     true,
		UserDataDir: filepath.Join(t.TempDir(), "profile"),
		StartedAt:   time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("save state failed: %v", err)
	}

	previousFind := findProcessesByUserDataDir
	previousProbe := probeManagedCDP
	attempts := 0
	findProcessesByUserDataDir = func(string) ([]macos.ProcessInfo, error) {
		return []macos.ProcessInfo{{PID: 123}}, nil
	}
	probeManagedCDP = func(string) CDPReport {
		attempts++
		if attempts == 1 {
			return CDPReport{Status: cdpStatusVersionDown, LastError: "warming"}
		}
		return CDPReport{Status: cdpStatusOK, VersionEndpoint: "http://127.0.0.1/json/version"}
	}
	t.Cleanup(func() {
		findProcessesByUserDataDir = previousFind
		probeManagedCDP = previousProbe
	})

	report, err := Status(paths)
	if err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !report.Ready || report.Stale || report.CDP.Status != cdpStatusOK {
		t.Fatalf("expected healed cdp report: %+v", report)
	}
	if attempts < 2 {
		t.Fatalf("expected cdp heal retry, attempts=%d", attempts)
	}
}

func TestSaveAndLoadState(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	state := State{
		Mode:          profile.ModeIsolated,
		Managed:       true,
		RuntimeSource: "managed_auto",
		BinaryPath:    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		URL:           "https://chatgpt.com/atlas?get-started",
		UserDataDir:   filepath.Join(t.TempDir(), "profile"),
	}
	if err := SaveState(paths, state); err != nil {
		t.Fatalf("save state failed: %v", err)
	}

	loaded, err := LoadState(paths)
	if err != nil {
		t.Fatalf("load state failed: %v", err)
	}
	if loaded.UserDataDir != state.UserDataDir {
		t.Fatalf("unexpected user data dir: %s", loaded.UserDataDir)
	}
	if loaded.RuntimeSource != state.RuntimeSource {
		t.Fatalf("unexpected runtime source: %s", loaded.RuntimeSource)
	}

	if _, err := os.Stat(paths.SessionFile); err != nil {
		t.Fatalf("state file not created: %v", err)
	}
}

func TestProbeCDPFromActivePortAndVersionEndpoint(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"webSocketDebuggerUrl":"ws://127.0.0.1/devtools/browser/test"}`))
	}))
	server.Listener = listener
	server.Start()
	defer server.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	dir := t.TempDir()
	activePortFile := filepath.Join(dir, devToolsActivePortFile)
	content := []byte(strings.Join([]string{strconv.Itoa(port), "/devtools/browser/test"}, "\n"))
	if err := os.WriteFile(activePortFile, content, 0o644); err != nil {
		t.Fatalf("write active port failed: %v", err)
	}

	report := ProbeCDP(dir)
	if report.Status != cdpStatusOK {
		t.Fatalf("unexpected cdp status: %s (%s)", report.Status, report.LastError)
	}
	if report.BrowserWebSocketURL == "" {
		t.Fatal("expected browser websocket url")
	}
}

func TestProbeCDPHandlesMissingPortFile(t *testing.T) {
	report := ProbeCDP(t.TempDir())
	if report.Status != cdpStatusPortFileAbsent {
		t.Fatalf("unexpected cdp status: %s", report.Status)
	}
}

func TestResultRenderIncludesRuntimeSource(t *testing.T) {
	rendered := Result{
		DryRun:        true,
		BinaryPath:    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		RuntimeSource: "managed_auto",
		Mode:          profile.ModeIsolated,
		Managed:       true,
		StateFile:     "/tmp/state.json",
		Args:          []string{"--app=https://example.com"},
	}.Render()

	if !strings.Contains(rendered, "runtime_source=managed_auto") {
		t.Fatalf("missing runtime source: %s", rendered)
	}
}
