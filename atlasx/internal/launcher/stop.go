package launcher

import (
	"os"
	"syscall"
)

func terminatePID(pid int) error {
	return signalPID(pid, syscall.SIGTERM)
}

func killPID(pid int) error {
	return signalPID(pid, syscall.SIGKILL)
}

func signalPID(pid int, signal syscall.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(signal)
}
