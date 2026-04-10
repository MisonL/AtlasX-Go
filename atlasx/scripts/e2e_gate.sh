#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

ATLASD_PORT="${ATLASX_E2E_PORT:-17537}"
ATLASD_LOG="$(mktemp -t atlasx-atlasd.XXXXXX.log)"
ATLASD_PID=""
FIRST_TAB_ID=""
STATUS_JSON=""
STATUS_COMPACT=""
UNCOVERED=()

cleanup() {
  if [[ -n "$ATLASD_PID" ]]; then
    kill "$ATLASD_PID" >/dev/null 2>&1 || true
    wait "$ATLASD_PID" >/dev/null 2>&1 || true
  fi
  rm -f "$ATLASD_LOG"
}

trap cleanup EXIT

log() {
  printf '%s\n' "$*"
}

mark_uncovered() {
  UNCOVERED+=("$1: $2")
  log "UNCOVERED $1: $2"
}

run_step() {
  local label="$1"
  shift
  log "RUN $label"
  "$@"
}

status_has_true() {
  local key="$1"
  [[ "$STATUS_COMPACT" == *"\"$key\":true"* ]]
}

status_line_value() {
  local input="$1"
  local key="$2"
  printf '%s\n' "$input" | sed -n "s/^${key}=//p" | head -n 1
}

refresh_status() {
  STATUS_JSON="$(curl -sf "http://127.0.0.1:${ATLASD_PORT}/v1/status")"
  STATUS_COMPACT="$(printf '%s' "$STATUS_JSON" | tr -d '[:space:]')"
}

start_atlasd() {
  log "RUN atlasd server"
  go run ./cmd/atlasd --listen "127.0.0.1:${ATLASD_PORT}" >"$ATLASD_LOG" 2>&1 &
  ATLASD_PID="$!"

  for _ in $(seq 1 30); do
    if curl -sf "http://127.0.0.1:${ATLASD_PORT}/healthz" >/dev/null; then
      refresh_status
      return 0
    fi
    sleep 1
  done

  log "atlasd log:"
  cat "$ATLASD_LOG"
  return 1
}

first_tab_id_from_output() {
  printf '%s\n' "$1" | sed -n 's/^id=\([^ ]*\).*/\1/p' | head -n 1
}

run_browser_data_open_smoke() {
  local history_output=""
  local bookmarks_output=""
  local downloads_output=""
  local session_ready="false"

  if status_has_true "managed_session_live"; then
    session_ready="true"
  fi

  history_output="$(go run ./cmd/atlasctl history list 2>&1 || true)"
  if printf '%s\n' "$history_output" | grep -q '^index='; then
    if [[ "$session_ready" != "true" ]]; then
      mark_uncovered "browser-data open smoke" "history 数据已落盘但当前没有受管浏览器会话"
      return 0
    fi
    run_step "history open smoke" go run ./cmd/atlasctl history open 0 >/dev/null
    return 0
  fi

  bookmarks_output="$(go run ./cmd/atlasctl bookmarks list 2>&1 || true)"
  if printf '%s\n' "$bookmarks_output" | grep -q '^index='; then
    if [[ "$session_ready" != "true" ]]; then
      mark_uncovered "browser-data open smoke" "bookmarks 数据已落盘但当前没有受管浏览器会话"
      return 0
    fi
    run_step "bookmarks open smoke" go run ./cmd/atlasctl bookmarks open 0 >/dev/null
    return 0
  fi

  downloads_output="$(go run ./cmd/atlasctl downloads list 2>&1 || true)"
  if printf '%s\n' "$downloads_output" | grep -q '^index='; then
    if [[ "$session_ready" != "true" ]]; then
      mark_uncovered "browser-data open smoke" "downloads 数据已落盘但当前没有受管浏览器会话"
      return 0
    fi
    run_step "downloads open smoke" go run ./cmd/atlasctl downloads open 0 >/dev/null
    return 0
  fi

  mark_uncovered "browser-data open smoke" "history/bookmarks/downloads 当前都没有可打开的已落盘数据"
}

run_sidebar_real_smoke() {
  if ! status_has_true "sidebar_qa_ready"; then
    mark_uncovered "sidebar ask real smoke" "sidebar_qa_ready=false，缺失真实 provider 凭据或配置"
    return 0
  fi
  if [[ -z "$FIRST_TAB_ID" ]]; then
    mark_uncovered "sidebar ask real smoke" "没有可用 page target"
    return 0
  fi

  log "RUN sidebar ask real smoke"
  curl -sf \
    -X POST \
    -H "Content-Type: application/json" \
    --data "{\"tab_id\":\"${FIRST_TAB_ID}\",\"question\":\"请用一句话确认当前页面已连接\"}" \
    "http://127.0.0.1:${ATLASD_PORT}/v1/sidebar/ask" >/dev/null
}

run_runtime_smoke() {
  local runtime_status_output=""
  local plan_status_output=""
  local plan_phase=""
  local plan_version=""
  local plan_channel=""
  local plan_source_url=""
  local plan_sha256=""
  local plan_archive_path=""
  local plan_bundle_path=""

  run_step "runtime status" go run ./cmd/atlasctl runtime status >/dev/null
  runtime_status_output="$(go run ./cmd/atlasctl runtime status)"
  plan_status_output="$(go run ./cmd/atlasctl runtime plan status)"

  if printf '%s\n' "$runtime_status_output" | grep -q 'manifest_present=true'; then
    run_step "runtime verify smoke" go run ./cmd/atlasctl runtime verify >/dev/null
  else
    mark_uncovered "runtime verify smoke" "当前本机没有 staged managed runtime"
  fi

  if [[ "${ATLASX_E2E_ALLOW_INSTALL:-0}" == "1" ]]; then
    if printf '%s\n' "$plan_status_output" | grep -q 'install_plan_present=true'; then
      plan_phase="$(status_line_value "$plan_status_output" "install_plan_phase")"
      if [[ "$plan_phase" != "planned" ]]; then
        plan_version="$(status_line_value "$plan_status_output" "install_plan_version")"
        plan_channel="$(status_line_value "$plan_status_output" "install_plan_channel")"
        plan_source_url="$(status_line_value "$plan_status_output" "install_plan_source_url")"
        plan_sha256="$(status_line_value "$plan_status_output" "install_plan_expected_sha256")"
        plan_archive_path="$(status_line_value "$plan_status_output" "install_plan_archive_path")"
        plan_bundle_path="$(status_line_value "$plan_status_output" "install_plan_staged_bundle_path")"
        run_step "runtime install plan reset" \
          go run ./cmd/atlasctl runtime plan create \
            --version "$plan_version" \
            --channel "$plan_channel" \
            --url "$plan_source_url" \
            --sha256 "$plan_sha256" \
            --archive-path "$plan_archive_path" \
            --bundle-path "$plan_bundle_path" >/dev/null
      fi
      run_step "runtime install smoke" go run ./cmd/atlasctl runtime install >/dev/null
    else
      mark_uncovered "runtime install smoke" "允许安装但当前没有 install plan"
    fi
  else
    mark_uncovered "runtime install smoke" "未设置 ATLASX_E2E_ALLOW_INSTALL=1"
  fi

  run_step "runtime rollback contract" \
    go test ./internal/managedruntime -run 'TestAdvanceInstallPlanRollbackPath|TestInstallRollsBackToPreviousRuntimeAfterStageFailure' >/dev/null
}

run_browser_smoke() {
  local tabs_output=""

  if ! status_has_true "managed_session_live"; then
    mark_uncovered "tabs capture smoke" "当前没有受管浏览器会话"
    run_browser_data_open_smoke
    return 0
  fi

  tabs_output="$(go run ./cmd/atlasctl tabs list)"
  FIRST_TAB_ID="$(first_tab_id_from_output "$tabs_output")"
  if [[ -z "$FIRST_TAB_ID" ]]; then
    mark_uncovered "tabs capture smoke" "tabs list 成功但没有 page target"
  else
    run_step "tabs capture smoke" go run ./cmd/atlasctl tabs capture "$FIRST_TAB_ID" >/dev/null
  fi

  run_browser_data_open_smoke
}

main() {
  run_step "offline go test" go test ./...
  run_step "launch-webapp dry-run" go run ./cmd/atlasctl launch-webapp --dry-run >/dev/null
  run_step "atlasd bootstrap once" go run ./cmd/atlasd --once >/dev/null
  start_atlasd
  run_runtime_smoke
  run_browser_smoke
  refresh_status
  run_sidebar_real_smoke

  log "E2E gate finished"
  if [[ "${#UNCOVERED[@]}" -gt 0 ]]; then
    log "UNCOVERED summary:"
    for item in "${UNCOVERED[@]}"; do
      log "  - ${item}"
    done
  else
    log "UNCOVERED summary: none"
  fi
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
