package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseEvidenceScriptWritesExpectedArtifacts(t *testing.T) {
	if os.Getenv("ATLASX_RELEASE_EVIDENCE_ACTIVE") == "1" {
		t.Skip("skip nested release evidence invocation")
	}

	outputDir := t.TempDir()
	repoDir := filepath.Clean(filepath.Join(".."))
	stubDir := createReleaseEvidenceStubDir(t)

	command := exec.Command("/bin/bash", "scripts/release_evidence.sh", outputDir)
	command.Dir = repoDir
	command.Env = append(os.Environ(),
		"PATH="+stubDir+":"+os.Getenv("PATH"),
		"STUB_DATE_OUTPUT=20260409T120000Z",
		"STUB_GO_TEST_EXIT=0",
		"STUB_GO_TEST_OUTPUT=go test output\n",
		"STUB_ATLASD_ONCE_EXIT=0",
		`STUB_ATLASD_ONCE_OUTPUT={"runtime_manifest_version":"136.0.7103.114","runtime_manifest_channel":"stable","sidebar_qa_default_provider":"primary"}`+"\n",
		"STUB_E2E_GATE_EXIT=0",
		"STUB_E2E_GATE_OUTPUT=E2E gate finished\nUNCOVERED summary: none\n",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("release evidence script failed: %v output=%s", err, string(output))
	}

	assertFileContains(t, filepath.Join(outputDir, "go-test.log"), "go test output")
	assertFileContains(t, filepath.Join(outputDir, "e2e-gate.log"), "E2E gate finished")
	assertFileContains(t, filepath.Join(outputDir, "atlasd-once.log"), `"runtime_manifest_version":"136.0.7103.114"`)
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "success=true")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "go-test.log")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "e2e-gate.log")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "atlasd-once.log")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "collected_at=20260409T120000Z")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "runtime_manifest_version=136.0.7103.114")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "runtime_manifest_channel=stable")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "sidebar_default_provider=primary")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "uncovered_count=0")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "uncovered_items=none")
}

func TestReleaseEvidenceScriptReturnsFailureAndSummaryWhenStepFails(t *testing.T) {
	if os.Getenv("ATLASX_RELEASE_EVIDENCE_ACTIVE") == "1" {
		t.Skip("skip nested release evidence invocation")
	}

	outputDir := t.TempDir()
	repoDir := filepath.Clean(filepath.Join(".."))
	stubDir := createReleaseEvidenceStubDir(t)

	command := exec.Command("/bin/bash", "scripts/release_evidence.sh", outputDir)
	command.Dir = repoDir
	command.Env = append(os.Environ(),
		"PATH="+stubDir+":"+os.Getenv("PATH"),
		"STUB_DATE_OUTPUT=20260409T120500Z",
		"STUB_GO_TEST_EXIT=0",
		"STUB_GO_TEST_OUTPUT=go test output\n",
		"STUB_ATLASD_ONCE_EXIT=0",
		"STUB_ATLASD_ONCE_OUTPUT={}\n",
		"STUB_E2E_GATE_EXIT=9",
		"STUB_E2E_GATE_OUTPUT=e2e gate failure\nUNCOVERED summary:\n  - tabs capture smoke: 当前没有受管浏览器会话\n  - sidebar ask real smoke: sidebar_qa_ready=false\n",
	)
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatalf("expected release evidence script to fail, output=%s", string(output))
	}

	assertFileContains(t, filepath.Join(outputDir, "go-test.log"), "go test output")
	assertFileContains(t, filepath.Join(outputDir, "e2e-gate.log"), "e2e gate failure")
	assertFileContains(t, filepath.Join(outputDir, "atlasd-once.log"), "{}")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "success=false")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "failed_steps=bash scripts/e2e_gate.sh")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "exit_code=9")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "runtime_manifest_version=none")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "runtime_manifest_channel=none")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "sidebar_default_provider=none")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "uncovered_count=2")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "tabs capture smoke: 当前没有受管浏览器会话")
	assertFileContains(t, filepath.Join(outputDir, "SUMMARY.md"), "sidebar ask real smoke: sidebar_qa_ready=false")
}

func createReleaseEvidenceStubDir(t *testing.T) string {
	t.Helper()

	stubDir := t.TempDir()

	stubGo := `#!/bin/bash
set -euo pipefail
if [[ "${1-}" == "test" && "${2-}" == "./..." ]]; then
  printf '%s' "${STUB_GO_TEST_OUTPUT:-}"
  exit "${STUB_GO_TEST_EXIT:-0}"
fi
if [[ "${1-}" == "run" && "${2-}" == "./cmd/atlasd" && "${3-}" == "--once" ]]; then
  printf '%s' "${STUB_ATLASD_ONCE_OUTPUT:-}"
  exit "${STUB_ATLASD_ONCE_EXIT:-0}"
fi
printf 'unexpected go invocation: %s\n' "$*" >&2
exit 1
`

	stubBash := `#!/bin/bash
set -euo pipefail
if [[ "${1-}" == "scripts/e2e_gate.sh" ]]; then
  printf '%s' "${STUB_E2E_GATE_OUTPUT:-}"
  exit "${STUB_E2E_GATE_EXIT:-0}"
fi
exec /bin/bash "$@"
`

	stubDate := `#!/bin/bash
set -euo pipefail
printf '%s\n' "${STUB_DATE_OUTPUT:-20260409T000000Z}"
`

	writeExecutableFile(t, filepath.Join(stubDir, "go"), stubGo)
	writeExecutableFile(t, filepath.Join(stubDir, "bash"), stubBash)
	writeExecutableFile(t, filepath.Join(stubDir, "date"), stubDate)

	return stubDir
}

func writeExecutableFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write executable file %s: %v", path, err)
	}
}

func assertFileContains(t *testing.T, path string, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}
	if !strings.Contains(string(content), want) {
		t.Fatalf("expected %s to contain %q, got %s", path, want, string(content))
	}
}
