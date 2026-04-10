# Repository Guidelines

## Project Structure & Module Organization

The repository root contains planning and review material in `docs/`, the task source in `tasks.csv`, and the Go application in `atlasx/`.

- `atlasx/cmd/atlasctl`: CLI entrypoint for local control actions.
- `atlasx/cmd/atlasd`: local HTTP daemon entrypoint.
- `atlasx/internal/*`: feature packages such as `sidebar`, `managedruntime`, `tabs`, `memory`, and `daemon`.
- `atlasx/scripts`: gate and release evidence scripts plus their tests.
- `docs/reviews`: audit records and release evidence snapshots.

Keep production code under `atlasx/internal` or `atlasx/cmd`. Do not add ad hoc tooling or generated files to the repository root.

## Build, Test, and Development Commands

Run commands from `atlasx/` unless noted otherwise.

- `go build ./cmd/atlasctl ./cmd/atlasd`: build both binaries.
- `go test ./...`: run the full Go test suite.
- `go vet ./...`: run static checks before release work.
- `bash scripts/e2e_gate.sh`: execute the end-to-end gate and report uncovered items.
- `bash scripts/release_evidence.sh /tmp/atlasx-release`: collect release evidence and produce `SUMMARY.md`.
- `go run ./cmd/atlasd --once`: print the current machine/runtime/sidebar status snapshot.

## Coding Style & Naming Conventions

Use standard Go formatting and keep code `gofmt`-clean. Prefer short, explicit functions and package-local helpers over deep nesting. Package names stay lowercase and descriptive; exported identifiers use Go’s normal `CamelCase`, unexported names use `camelCase`.

Shell scripts should use `bash`, `set -euo pipefail`, and clear function names such as `refresh_status` or `run_runtime_smoke`.

## Testing Guidelines

This project uses Go’s built-in `testing` package. Place tests next to the package they validate using `*_test.go`. Name tests by behavior, for example `TestSidebarSetProviderWritesRegistryAndShowsStatus`.

For task completion and release work, run at least:

- `go test ./...`
- `go vet ./...`
- the relevant script test or smoke command if scripts or runtime flows changed

## Commit & Pull Request Guidelines

Recent history uses task-first commit subjects such as `T148 complete sidebar ask real smoke`. Follow the same pattern: `<TaskID> <imperative summary>`.

PRs should include:

- scope and affected paths
- validation commands and exit results
- linked task IDs from `tasks.csv`
- screenshots or logs only when UI/runtime behavior changed

## Security & Configuration Tips

Do not commit real API keys, tokens, or machine-specific secrets. Store only environment variable names such as `OPENAI_API_KEY` in config. Treat `tasks.csv` as the single task source of truth when updating status.
