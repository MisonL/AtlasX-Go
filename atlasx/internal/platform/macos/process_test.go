package macos

import "testing"

func TestParsePSOutput(t *testing.T) {
	output := " 101 /Applications/Google Chrome.app/Contents/MacOS/Google Chrome --user-data-dir=/tmp/atlasx\n 202 /usr/bin/ssh-agent\n"
	processes := parsePSOutput(output)

	if len(processes) != 2 {
		t.Fatalf("unexpected process count: %d", len(processes))
	}
	if processes[0].PID != 101 {
		t.Fatalf("unexpected pid: %d", processes[0].PID)
	}
	if processes[0].Command == "" {
		t.Fatal("expected command")
	}
}

func TestFilterProcesses(t *testing.T) {
	processes := []ProcessInfo{
		{PID: 10, Command: "/a --user-data-dir=/tmp/atlasx"},
		{PID: 11, Command: "/b --user-data-dir=/tmp/other"},
	}

	matches := filterProcesses(processes, "--user-data-dir=/tmp/atlasx")
	if len(matches) != 1 {
		t.Fatalf("unexpected match count: %d", len(matches))
	}
	if matches[0].PID != 10 {
		t.Fatalf("unexpected matched pid: %d", matches[0].PID)
	}
}
