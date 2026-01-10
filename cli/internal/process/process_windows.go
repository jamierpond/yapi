//go:build windows

package process

import "os/exec"

// configurePlatform is a no-op on Windows.
// Future enhancement: could use Job Objects for process tree management.
func configurePlatform(cmd *exec.Cmd) {
	// No-op on Windows
}

// Stop terminates the process on Windows.
// Note: This kills the main process but may not kill all child processes.
// For dev servers, this is usually sufficient as the port gets freed.
func (mp *ManagedProcess) Stop() error {
	if mp.cmd == nil || mp.cmd.Process == nil {
		return nil
	}
	return mp.cmd.Process.Kill()
}
