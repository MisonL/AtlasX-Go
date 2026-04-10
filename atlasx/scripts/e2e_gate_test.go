package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRunBrowserDataOpenSmokeMarksNoDataWithoutSession(t *testing.T) {
	output := runGateShellTest(t, `source ./scripts/e2e_gate.sh
STATUS_COMPACT='{"managed_session_live":false}'
UNCOVERED=()
run_browser_data_open_smoke
printf 'count=%s\n' "${#UNCOVERED[@]}"
for item in "${UNCOVERED[@]}"; do
  printf 'item=%s\n' "$item"
done
`, map[string]string{
		"STUB_HISTORY_LIST":   "",
		"STUB_BOOKMARKS_LIST": "",
		"STUB_DOWNLOADS_LIST": "",
	})

	if !strings.Contains(output, "count=1") {
		t.Fatalf("expected one uncovered item, got output=%s", output)
	}
	if !strings.Contains(output, "item=browser-data open smoke: history/bookmarks/downloads 当前都没有可打开的已落盘数据") {
		t.Fatalf("expected no-data uncovered reason, got output=%s", output)
	}
}

func TestRunBrowserDataOpenSmokeMarksSessionMissingWhenDataPresent(t *testing.T) {
	output := runGateShellTest(t, `source ./scripts/e2e_gate.sh
STATUS_COMPACT='{"managed_session_live":false}'
UNCOVERED=()
run_browser_data_open_smoke
printf 'count=%s\n' "${#UNCOVERED[@]}"
for item in "${UNCOVERED[@]}"; do
  printf 'item=%s\n' "$item"
done
`, map[string]string{
		"STUB_HISTORY_LIST":   "index=0 last_visit_time=2026-04-09T00:00:00Z visit_count=1 title=\"AtlasX\" url=https://chatgpt.com\n",
		"STUB_BOOKMARKS_LIST": "",
		"STUB_DOWNLOADS_LIST": "",
	})

	if !strings.Contains(output, "count=1") {
		t.Fatalf("expected one uncovered item, got output=%s", output)
	}
	if !strings.Contains(output, "item=browser-data open smoke: history 数据已落盘但当前没有受管浏览器会话") {
		t.Fatalf("expected session-missing uncovered reason, got output=%s", output)
	}
}

func TestRunBrowserSmokeKeepsTabsCaptureAndBrowserDataReasonsIndependent(t *testing.T) {
	output := runGateShellTest(t, `source ./scripts/e2e_gate.sh
STATUS_COMPACT='{"managed_session_live":false}'
UNCOVERED=()
run_browser_smoke
printf 'count=%s\n' "${#UNCOVERED[@]}"
for item in "${UNCOVERED[@]}"; do
  printf 'item=%s\n' "$item"
done
`, map[string]string{
		"STUB_HISTORY_LIST":   "",
		"STUB_BOOKMARKS_LIST": "index=0 root=bookmark_bar name=\"AtlasX\" url=https://chatgpt.com\n",
		"STUB_DOWNLOADS_LIST": "",
	})

	if !strings.Contains(output, "count=2") {
		t.Fatalf("expected two uncovered items, got output=%s", output)
	}
	if !strings.Contains(output, "item=tabs capture smoke: 当前没有受管浏览器会话") {
		t.Fatalf("expected tabs capture uncovered reason, got output=%s", output)
	}
	if !strings.Contains(output, "item=browser-data open smoke: bookmarks 数据已落盘但当前没有受管浏览器会话") {
		t.Fatalf("expected browser-data uncovered reason, got output=%s", output)
	}
}

func TestRunRuntimeSmokeResetsStagedPlanBeforeInstall(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "go.log")
	output := runGateShellTest(t, `source ./scripts/e2e_gate.sh
ATLASX_E2E_ALLOW_INSTALL=1
UNCOVERED=()
run_runtime_smoke
printf 'count=%s\n' "${#UNCOVERED[@]}"
for item in "${UNCOVERED[@]}"; do
  printf 'item=%s\n' "$item"
done
`, map[string]string{
		"STUB_GO_LOG": logFile,
		"STUB_RUNTIME_STATUS": strings.Join([]string{
			"manifest_present=true",
		}, "\n") + "\n",
		"STUB_PLAN_STATUS": strings.Join([]string{
			"install_plan_present=true",
			"install_plan_version=146.0.7680.178",
			"install_plan_channel=stable",
			"install_plan_source_url=https://storage.googleapis.com/chrome-for-testing-public/146.0.7680.178/mac-x64/chrome-mac-x64.zip",
			"install_plan_expected_sha256=f09a7c7a7c3f2fabdecac516fdb466fb753e7118180adb1b93777baf49cd3151",
			"install_plan_archive_path=/tmp/chrome-146.0.7680.178-mac-x64.zip",
			"install_plan_staged_bundle_path=/Users/mison/Library/Application Support/AtlasX/runtime/Chromium.app",
			"install_plan_phase=staged",
		}, "\n") + "\n",
	})

	if !strings.Contains(output, "count=0") {
		t.Fatalf("expected no uncovered items, got output=%s", output)
	}

	logData, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read stub log failed: %v", err)
	}
	logOutput := string(logData)
	if !strings.Contains(logOutput, "run ./cmd/atlasctl runtime plan create --version 146.0.7680.178 --channel stable --url https://storage.googleapis.com/chrome-for-testing-public/146.0.7680.178/mac-x64/chrome-mac-x64.zip --sha256 f09a7c7a7c3f2fabdecac516fdb466fb753e7118180adb1b93777baf49cd3151 --archive-path /tmp/chrome-146.0.7680.178-mac-x64.zip --bundle-path /Users/mison/Library/Application Support/AtlasX/runtime/Chromium.app") {
		t.Fatalf("expected runtime plan reset invocation, got log=%s", logOutput)
	}
	if !strings.Contains(logOutput, "run ./cmd/atlasctl runtime install") {
		t.Fatalf("expected runtime install invocation, got log=%s", logOutput)
	}
}

func TestResolveAtlasdPortUsesEphemeralPortWhenUnset(t *testing.T) {
	output := runGateShellTest(t, `source ./scripts/e2e_gate.sh
ATLASD_PORT=""
port="$(resolve_atlasd_port)"
printf 'port=%s\n' "$port"
`, map[string]string{})

	line := strings.TrimSpace(output)
	if !strings.HasPrefix(line, "port=") {
		t.Fatalf("expected port output, got %s", output)
	}
	value, err := strconv.Atoi(strings.TrimPrefix(line, "port="))
	if err != nil {
		t.Fatalf("expected numeric port, got %s err=%v", output, err)
	}
	if value <= 0 {
		t.Fatalf("expected positive port, got %d", value)
	}
}

func runGateShellTest(t *testing.T, script string, env map[string]string) string {
	t.Helper()

	repoDir := filepath.Clean(filepath.Join(".."))
	stubDir := t.TempDir()
	stubGoPath := filepath.Join(stubDir, "go")
	stubGo := `#!/usr/bin/env bash
set -euo pipefail
if [[ -n "${STUB_GO_LOG:-}" ]]; then
  printf '%s\n' "$*" >>"${STUB_GO_LOG}"
fi
cmd="${1-}"
target="${2-}"
action="${3-}"
arg="${4-}"
arg2="${5-}"
if [[ "$cmd" == "run" && "$target" == "./cmd/atlasctl" ]]; then
  case "$action" in
    runtime)
      case "$arg" in
        status)
          printf '%s' "${STUB_RUNTIME_STATUS:-}"
          exit 0
          ;;
        verify)
          exit "${STUB_RUNTIME_VERIFY_EXIT:-0}"
          ;;
        install)
          printf '%s' "${STUB_RUNTIME_INSTALL_OUTPUT:-}"
          exit "${STUB_RUNTIME_INSTALL_EXIT:-0}"
          ;;
        plan)
          case "$arg2" in
            status)
              printf '%s' "${STUB_PLAN_STATUS:-}"
              exit 0
              ;;
            create)
              printf '%s' "${STUB_PLAN_CREATE_OUTPUT:-}"
              exit "${STUB_PLAN_CREATE_EXIT:-0}"
              ;;
          esac
          ;;
      esac
      ;;
    history)
      if [[ "$arg" == "list" ]]; then
        printf '%s' "${STUB_HISTORY_LIST:-}"
        exit 0
      fi
      if [[ "$arg" == "open" ]]; then
        exit 0
      fi
      ;;
    bookmarks)
      if [[ "$arg" == "list" ]]; then
        printf '%s' "${STUB_BOOKMARKS_LIST:-}"
        exit 0
      fi
      if [[ "$arg" == "open" ]]; then
        exit 0
      fi
      ;;
    downloads)
      if [[ "$arg" == "list" ]]; then
        printf '%s' "${STUB_DOWNLOADS_LIST:-}"
        exit 0
      fi
      if [[ "$arg" == "open" ]]; then
        exit 0
      fi
      ;;
    tabs)
      if [[ "$arg" == "list" ]]; then
        printf '%s' "${STUB_TABS_LIST:-}"
        exit 0
      fi
      if [[ "$arg" == "capture" ]]; then
        exit 0
      fi
      ;;
  esac
fi
if [[ "$cmd" == "test" ]]; then
  exit 0
fi
printf 'unexpected go invocation: %s\n' "$*" >&2
exit 1
`
	if err := os.WriteFile(stubGoPath, []byte(stubGo), 0o755); err != nil {
		t.Fatalf("write stub go: %v", err)
	}

	command := exec.Command("bash", "-c", script)
	command.Dir = repoDir
	command.Env = append(os.Environ(), "PATH="+stubDir+":"+os.Getenv("PATH"))
	for key, value := range env {
		command.Env = append(command.Env, key+"="+value)
	}
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("shell script failed: %v output=%s", err, string(output))
	}
	return string(output)
}
