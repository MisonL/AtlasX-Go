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

func TestSaveAndLoadState(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	paths, err := macos.DiscoverPaths()
	if err != nil {
		t.Fatalf("discover paths failed: %v", err)
	}

	state := State{
		Mode:        profile.ModeIsolated,
		Managed:     true,
		BinaryPath:  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		URL:         "https://chatgpt.com/atlas?get-started",
		UserDataDir: filepath.Join(t.TempDir(), "profile"),
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
