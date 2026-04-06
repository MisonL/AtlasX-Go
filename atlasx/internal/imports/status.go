package imports

import (
	"encoding/json"
	"os"
	"path/filepath"

	"atlasx/internal/platform/macos"
)

const (
	importResultSucceeded = "succeeded"
	importResultFailed    = "failed"
)

type OperationStatus struct {
	GeneratedAt string `json:"generated_at"`
	Source      string `json:"source"`
	Result      string `json:"result"`
	Error       string `json:"error"`
}

func SaveChromeImportStatus(paths macos.Paths, status OperationStatus) error {
	return saveOperationStatus(chromeImportStatusPath(paths), paths, status)
}

func LoadChromeImportStatus(paths macos.Paths) (OperationStatus, error) {
	return loadOperationStatus(chromeImportStatusPath(paths))
}

func SaveSafariImportStatus(paths macos.Paths, status OperationStatus) error {
	return saveOperationStatus(safariImportStatusPath(paths), paths, status)
}

func LoadSafariImportStatus(paths macos.Paths) (OperationStatus, error) {
	return loadOperationStatus(safariImportStatusPath(paths))
}

func saveOperationStatus(path string, paths macos.Paths, status OperationStatus) error {
	if err := macos.EnsureDir(paths.StateRoot); err != nil {
		return err
	}

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func loadOperationStatus(path string) (OperationStatus, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return OperationStatus{}, err
	}

	var status OperationStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return OperationStatus{}, err
	}
	return status, nil
}

func chromeImportStatusPath(paths macos.Paths) string {
	return filepath.Join(paths.StateRoot, "chrome-import-status.json")
}

func safariImportStatusPath(paths macos.Paths) string {
	return filepath.Join(paths.StateRoot, "safari-import-status.json")
}
