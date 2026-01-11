//go:build windows

package process

import (
	"errors"
	"os"
	"os/exec"
)

// configurePlatform is a no-op on Windows.
// Future enhancement: could use Job Objects for process tree management.
func configurePlatform(cmd *exec.Cmd) {
	// No-op on Windows
}

// Stop terminates the process on Windows.
// Note: This kills the main process but may not kill all child processes.
// For dev servers, this is usually sufficient as the port gets freed.
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
	err := mp.cmd.Process.Kill()
	// Ignore "process already finished" errors
	if err != nil && !errors.Is(err, os.ErrProcessDone) {
		// On Windows, killing an already-dead process returns "Access is denied"
		// which we should ignore
		return nil
	}
	return nil
}
