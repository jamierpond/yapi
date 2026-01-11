//go:build !windows

package process

import (
	"os/exec"
	"syscall"
	"time"
)

// configurePlatform sets up Unix process groups for clean termination.
func configurePlatform(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// Stop gracefully terminates the process and its children.
// Sends SIGTERM to process group, waits up to 5 seconds, then SIGKILL.
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

	pgid, err := syscall.Getpgid(mp.cmd.Process.Pid)
	if err != nil {
		// Can't get process group, fall back to killing just the process
		return mp.cmd.Process.Kill()
	}

	// Send SIGTERM to process group (negative pgid targets the group)
	// Ignore error - process might already be dead
	_ = syscall.Kill(-pgid, syscall.SIGTERM)

	// Wait for process to exit with timeout
	done := make(chan error, 1)
	go func() {
		done <- mp.cmd.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		// Force kill the process group
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		<-done
		return nil
	}
}
