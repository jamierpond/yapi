//go:build windows

package process

import (
	"os/exec"
	"strconv"
)

// configurePlatform is a no-op on Windows.
func configurePlatform(cmd *exec.Cmd) {
	// No-op on Windows
}

// Stop terminates the process and its children on Windows.
// Stop is idempotent and safe to call multiple times.
func (mp *ManagedProcess) Stop() error {
	var stopErr error
	mp.stopOnce.Do(func() {
		stopErr = mp.stop()
	})
	return stopErr
}

func (mp *ManagedProcess) stop() error {
	if mp.cmd == nil || mp.cmd.Process == nil {
		return nil
	}
	// Use taskkill to kill the entire process tree
	// /T = Tree (kill children), /F = Force, /PID = Process ID
	killCmd := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(mp.cmd.Process.Pid))
	// Ignore errors - process may already be dead
	_ = killCmd.Run()
	return nil
}
