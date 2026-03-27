package macos

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type ProcessInfo struct {
	PID     int
	Command string
}

func FindProcessesByUserDataDir(userDataDir string) ([]ProcessInfo, error) {
	processes, err := listProcesses()
	if err != nil {
		return nil, err
	}

	needle := "--user-data-dir=" + filepath.Clean(userDataDir)
	return filterProcesses(processes, needle), nil
}

func listProcesses() ([]ProcessInfo, error) {
	output, err := exec.Command("ps", "-ax", "-o", "pid=,command=").Output()
	if err != nil {
		return nil, err
	}
	return parsePSOutput(string(output)), nil
}

func parsePSOutput(output string) []ProcessInfo {
	lines := strings.Split(output, "\n")
	processes := make([]ProcessInfo, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}

		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		command := strings.TrimSpace(strings.TrimPrefix(trimmed, fields[0]))
		processes = append(processes, ProcessInfo{PID: pid, Command: command})
	}
	return processes
}

func filterProcesses(processes []ProcessInfo, needle string) []ProcessInfo {
	matches := make([]ProcessInfo, 0, len(processes))
	for _, process := range processes {
		if strings.Contains(process.Command, needle) {
			matches = append(matches, process)
		}
	}
	return matches
}
