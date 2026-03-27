package chrome

import (
	"os"
	"os/exec"
	"path/filepath"
)

func DefaultUserDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Library", "Application Support", "Google", "Chrome-Atlas-X64")
}

func BuildLaunchArgs(appPath, url, userDataDir string, sharedProfile bool) []string {
	if sharedProfile {
		return []string{"-a", appPath, "--args", "--app=" + url}
	}
	return []string{
		"-na", appPath,
		"--args",
		"--app=" + url,
		"--user-data-dir=" + userDataDir,
	}
}

func Launch(url, userDataDir string, sharedProfile bool) error {
	detection, err := Detect("")
	if err != nil {
		return err
	}
	if !sharedProfile && userDataDir != "" {
		if err := os.MkdirAll(userDataDir, 0o755); err != nil {
			return err
		}
	}
	args := BuildLaunchArgs(detection.BinaryPath, url, userDataDir, sharedProfile)
	cmd := exec.Command("open", args...)
	return cmd.Run()
}
