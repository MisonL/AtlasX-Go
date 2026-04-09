#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
DEFAULT_OUTPUT_DIR="$ROOT_DIR/../docs/reviews/release-evidence/$TIMESTAMP"
OUTPUT_DIR="${1:-${ATLASX_RELEASE_EVIDENCE_DIR:-$DEFAULT_OUTPUT_DIR}}"

mkdir -p "$OUTPUT_DIR"
OUTPUT_DIR="$(cd "$OUTPUT_DIR" && pwd)"
SUMMARY_FILE="$OUTPUT_DIR/SUMMARY.md"
ATLASD_ONCE_LOG="$OUTPUT_DIR/atlasd-once.log"
E2E_GATE_LOG="$OUTPUT_DIR/e2e-gate.log"
TASKS_FILE="${ATLASX_RELEASE_EVIDENCE_TASKS_FILE:-$ROOT_DIR/../tasks.csv}"

RESULT_LINES=()
FAILURES=()
RUNTIME_MANIFEST_VERSION="none"
RUNTIME_MANIFEST_CHANNEL="none"
SIDEBAR_DEFAULT_PROVIDER="none"
UNCOVERED_ITEMS=()
TASKS_TOTAL="0"
TASKS_DONE="0"
TASKS_DOING="0"
TASKS_TODO="0"
ATLASD_READY="false_or_unknown"
MANAGED_SESSION_LIVE="false_or_unknown"
SIDEBAR_QA_READY="false_or_unknown"
RELEASE_READY="false"
RELEASE_BLOCKERS=()

log() {
  printf '%s\n' "$*"
}

run_and_capture() {
  local label="$1"
  local logfile="$2"
  shift 2

  local logfile_path="$OUTPUT_DIR/$logfile"
  local exit_code=0

  log "RUN $label"
  if "$@" >"$logfile_path" 2>&1; then
    exit_code=0
  else
    exit_code=$?
    FAILURES+=("$label")
  fi

  RESULT_LINES+=("$label|$exit_code|$logfile")
}

extract_json_string_field() {
  local logfile_path="$1"
  local field_name="$2"

  if [[ ! -f "$logfile_path" ]]; then
    printf 'none'
    return 0
  fi

  node -e '
const fs = require("fs");
const filePath = process.argv[1];
const fieldName = process.argv[2];
try {
  const payload = JSON.parse(fs.readFileSync(filePath, "utf8"));
  const value = payload[fieldName];
  if (typeof value === "string" && value.trim() !== "") {
    process.stdout.write(value.trim());
  } else {
    process.stdout.write("none");
  }
} catch (error) {
  process.stdout.write("none");
}
' "$logfile_path" "$field_name"
}

extract_json_bool_field() {
  local logfile_path="$1"
  local field_name="$2"

  if [[ ! -f "$logfile_path" ]]; then
    printf 'false_or_unknown'
    return 0
  fi

  node -e '
const fs = require("fs");
const filePath = process.argv[1];
const fieldName = process.argv[2];
try {
  const payload = JSON.parse(fs.readFileSync(filePath, "utf8"));
  const value = payload[fieldName];
  if (value === true) {
    process.stdout.write("true");
  } else if (value === false) {
    process.stdout.write("false");
  } else {
    process.stdout.write("false_or_unknown");
  }
} catch (error) {
  process.stdout.write("false_or_unknown");
}
' "$logfile_path" "$field_name"
}

refresh_metadata_from_atlasd_once() {
  RUNTIME_MANIFEST_VERSION="$(extract_json_string_field "$ATLASD_ONCE_LOG" "runtime_manifest_version")"
  RUNTIME_MANIFEST_CHANNEL="$(extract_json_string_field "$ATLASD_ONCE_LOG" "runtime_manifest_channel")"
  SIDEBAR_DEFAULT_PROVIDER="$(extract_json_string_field "$ATLASD_ONCE_LOG" "sidebar_qa_default_provider")"
  ATLASD_READY="$(extract_json_bool_field "$ATLASD_ONCE_LOG" "ready")"
  MANAGED_SESSION_LIVE="$(extract_json_bool_field "$ATLASD_ONCE_LOG" "managed_session_live")"
  SIDEBAR_QA_READY="$(extract_json_bool_field "$ATLASD_ONCE_LOG" "sidebar_qa_ready")"
}

refresh_uncovered_from_e2e_gate() {
  UNCOVERED_ITEMS=()

  if [[ ! -f "$E2E_GATE_LOG" ]]; then
    return 0
  fi

  while IFS= read -r line; do
    UNCOVERED_ITEMS+=("${line#  - }")
  done < <(grep '^  - ' "$E2E_GATE_LOG" || true)
}

refresh_task_counts() {
  local counts

  if [[ ! -f "$TASKS_FILE" ]]; then
    TASKS_TOTAL="0"
    TASKS_DONE="0"
    TASKS_DOING="0"
    TASKS_TODO="0"
    return 0
  fi

  counts="$(node -e '
const fs = require("fs");
const filePath = process.argv[1];
const lines = fs.readFileSync(filePath, "utf8").trim().split(/\r?\n/).slice(1);
const stats = { total: 0, done: 0, doing: 0, todo: 0 };
for (const line of lines) {
  const cols = [];
  let cur = "";
  let quoted = false;
  for (let i = 0; i < line.length; i += 1) {
    const ch = line[i];
    if (ch === "\"") {
      if (quoted && line[i + 1] === "\"") {
        cur += "\"";
        i += 1;
      } else {
        quoted = !quoted;
      }
      continue;
    }
    if (ch === "," && !quoted) {
      cols.push(cur);
      cur = "";
      continue;
    }
    cur += ch;
  }
  cols.push(cur);
  stats.total += 1;
  const status = cols[5] || "";
  if (status === "已完成") stats.done += 1;
  if (status === "进行中") stats.doing += 1;
  if (status === "未开始") stats.todo += 1;
}
process.stdout.write(`${stats.total}|${stats.done}|${stats.doing}|${stats.todo}`);
' "$TASKS_FILE")"

  IFS='|' read -r TASKS_TOTAL TASKS_DONE TASKS_DOING TASKS_TODO <<<"$counts"
}

refresh_release_readiness() {
  RELEASE_BLOCKERS=()

  if [[ "${#FAILURES[@]}" -gt 0 ]]; then
    RELEASE_BLOCKERS+=("command_failures_present")
  fi
  if [[ "$ATLASD_READY" != "true" ]]; then
    RELEASE_BLOCKERS+=("atlasd_ready_false")
  fi
  if [[ "$TASKS_DOING" != "0" ]]; then
    RELEASE_BLOCKERS+=("tasks_in_progress_present")
  fi
  if [[ "$TASKS_TODO" != "0" ]]; then
    RELEASE_BLOCKERS+=("tasks_not_started_present")
  fi
  if [[ "${#UNCOVERED_ITEMS[@]}" -gt 0 ]]; then
    RELEASE_BLOCKERS+=("uncovered_items_present")
  fi

  if [[ "${#RELEASE_BLOCKERS[@]}" -eq 0 ]]; then
    RELEASE_READY="true"
  else
    RELEASE_READY="false"
  fi
}

write_summary() {
  {
    printf '# Release Evidence\n\n'
    printf -- '- collected_at=%s\n' "$TIMESTAMP"
    printf -- '- output_dir=%s\n' "$OUTPUT_DIR"
    printf -- '- runtime_manifest_version=%s\n' "$RUNTIME_MANIFEST_VERSION"
    printf -- '- runtime_manifest_channel=%s\n' "$RUNTIME_MANIFEST_CHANNEL"
    printf -- '- sidebar_default_provider=%s\n' "$SIDEBAR_DEFAULT_PROVIDER"
    printf -- '- atlasd_ready=%s\n' "$ATLASD_READY"
    printf -- '- managed_session_live=%s\n' "$MANAGED_SESSION_LIVE"
    printf -- '- sidebar_qa_ready=%s\n' "$SIDEBAR_QA_READY"
    printf -- '- uncovered_count=%s\n' "${#UNCOVERED_ITEMS[@]}"
    printf -- '- tasks_total=%s\n' "$TASKS_TOTAL"
    printf -- '- tasks_done=%s\n' "$TASKS_DONE"
    printf -- '- tasks_doing=%s\n' "$TASKS_DOING"
    printf -- '- tasks_todo=%s\n' "$TASKS_TODO"
    printf -- '- release_ready=%s\n' "$RELEASE_READY"
    printf '\n## Results\n\n'
    for line in "${RESULT_LINES[@]}"; do
      IFS='|' read -r label exit_code logfile <<<"$line"
      printf -- '- %s\n' "$label"
      printf '  exit_code=%s\n' "$exit_code"
      printf '  log_file=%s\n' "$logfile"
    done

    if [[ "${#FAILURES[@]}" -eq 0 ]]; then
      printf '\n## Final Status\n\n'
      printf -- '- success=true\n'
      printf -- '- failed_steps=none\n'
    else
      printf '\n## Final Status\n\n'
      printf -- '- success=false\n'
      printf -- '- failed_steps=%s\n' "$(IFS=,; printf '%s' "${FAILURES[*]}")"
    fi

    printf '\n## Uncovered\n\n'
    if [[ "${#UNCOVERED_ITEMS[@]}" -eq 0 ]]; then
      printf -- '- uncovered_items=none\n'
    else
      for item in "${UNCOVERED_ITEMS[@]}"; do
        printf -- '- %s\n' "$item"
      done
    fi

    printf '\n## Release Blockers\n\n'
    if [[ "${#RELEASE_BLOCKERS[@]}" -eq 0 ]]; then
      printf -- '- release_blockers=none\n'
    else
      for item in "${RELEASE_BLOCKERS[@]}"; do
        printf -- '- %s\n' "$item"
      done
    fi
  } >"$SUMMARY_FILE"
}

run_and_capture "go test ./..." "go-test.log" env ATLASX_RELEASE_EVIDENCE_ACTIVE=1 go test ./...
run_and_capture "bash scripts/e2e_gate.sh" "e2e-gate.log" env ATLASX_RELEASE_EVIDENCE_ACTIVE=1 bash scripts/e2e_gate.sh
run_and_capture "go run ./cmd/atlasd --once" "atlasd-once.log" go run ./cmd/atlasd --once
refresh_metadata_from_atlasd_once
refresh_uncovered_from_e2e_gate
refresh_task_counts
refresh_release_readiness
write_summary

if [[ "${#FAILURES[@]}" -gt 0 ]]; then
  log "release evidence capture failed"
  log "summary: $SUMMARY_FILE"
  exit 1
fi

log "release evidence capture finished"
log "summary: $SUMMARY_FILE"
