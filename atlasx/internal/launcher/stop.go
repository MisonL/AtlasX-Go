package launcher

import (
	"os"
	"syscall"
)

func terminatePID(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(syscall.SIGTERM)
}
