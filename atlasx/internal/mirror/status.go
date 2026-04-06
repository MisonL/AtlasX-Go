package mirror

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"atlasx/internal/platform/macos"
)

const (
	scanResultSucceeded = "succeeded"
	scanResultFailed    = "failed"
)

type ScanStatus struct {
	GeneratedAt string `json:"generated_at"`
	ProfileDir  string `json:"profile_dir"`
	Result      string `json:"result"`
	Error       string `json:"error"`
}

func Scan(paths macos.Paths, profileDir string) (Snapshot, error) {
	snapshot, err := Collect(profileDir)
	if err != nil {
		_ = SaveScanStatus(paths, ScanStatus{
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			ProfileDir:  profileDir,
			Result:      scanResultFailed,
			Error:       err.Error(),
		})
		return Snapshot{}, err
	}

	if err := Save(paths, snapshot); err != nil {
		_ = SaveScanStatus(paths, ScanStatus{
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
			ProfileDir:  profileDir,
			Result:      scanResultFailed,
			Error:       err.Error(),
		})
		return Snapshot{}, err
	}

	_ = SaveScanStatus(paths, ScanStatus{
		GeneratedAt: snapshot.GeneratedAt,
		ProfileDir:  profileDir,
		Result:      scanResultSucceeded,
	})
	return snapshot, nil
}

func SaveScanStatus(paths macos.Paths, status ScanStatus) error {
	if err := macos.EnsureDir(paths.StateRoot); err != nil {
		return err
	}

	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(scanStatusPath(paths), append(data, '\n'), 0o644)
}

func LoadScanStatus(paths macos.Paths) (ScanStatus, error) {
	data, err := os.ReadFile(scanStatusPath(paths))
	if err != nil {
		return ScanStatus{}, err
	}

	var status ScanStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return ScanStatus{}, err
	}
	return status, nil
}

func scanStatusPath(paths macos.Paths) string {
	return filepath.Join(paths.StateRoot, "mirror-scan-status.json")
}
