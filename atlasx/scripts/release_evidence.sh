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

RESULT_LINES=()
FAILURES=()

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

write_summary() {
  {
    printf '# Release Evidence\n\n'
    printf -- '- collected_at=%s\n' "$TIMESTAMP"
    printf -- '- output_dir=%s\n' "$OUTPUT_DIR"
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
  } >"$SUMMARY_FILE"
}

run_and_capture "go test ./..." "go-test.log" env ATLASX_RELEASE_EVIDENCE_ACTIVE=1 go test ./...
run_and_capture "bash scripts/e2e_gate.sh" "e2e-gate.log" env ATLASX_RELEASE_EVIDENCE_ACTIVE=1 bash scripts/e2e_gate.sh
run_and_capture "go run ./cmd/atlasd --once" "atlasd-once.log" go run ./cmd/atlasd --once
write_summary

if [[ "${#FAILURES[@]}" -gt 0 ]]; then
  log "release evidence capture failed"
  log "summary: $SUMMARY_FILE"
  exit 1
fi

log "release evidence capture finished"
log "summary: $SUMMARY_FILE"
