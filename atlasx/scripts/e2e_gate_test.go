package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
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

func runGateShellTest(t *testing.T, script string, env map[string]string) string {
	t.Helper()

	repoDir := filepath.Clean(filepath.Join(".."))
	stubDir := t.TempDir()
	stubGoPath := filepath.Join(stubDir, "go")
	stubGo := `#!/usr/bin/env bash
set -euo pipefail
cmd="${1-}"
target="${2-}"
action="${3-}"
arg="${4-}"
if [[ "$cmd" == "run" && "$target" == "./cmd/atlasctl" ]]; then
  case "$action" in
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
