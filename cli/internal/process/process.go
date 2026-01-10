// Package process provides managed process lifecycle for spawning and cleaning up subprocesses.
package process

import (
	"context"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"
)

// ManagedProcess wraps an exec.Cmd with lifecycle management.
type ManagedProcess struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
}

// Start spawns a new process using the shell.
// The process runs in the background and can be stopped with Stop().
// If verbose is true, stdout/stderr are piped to os.Stdout/os.Stderr.
func Start(ctx context.Context, command string, verbose bool) (*ManagedProcess, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	// Set process group so we can kill all children
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	mp := &ManagedProcess{cmd: cmd}

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// Capture but discard output
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		mp.stdout = stdout

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nil, err
		}
		mp.stderr = stderr

		// Drain pipes to prevent blocking
		go func() { _, _ = io.Copy(io.Discard, stdout) }()
		go func() { _, _ = io.Copy(io.Discard, stderr) }()
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return mp, nil
}

// Stop gracefully terminates the process.
// Sends SIGTERM first, waits up to 5 seconds, then SIGKILL.
func (mp *ManagedProcess) Stop() error {
	if mp.cmd == nil || mp.cmd.Process == nil {
		return nil
	}

	// Get the process group ID
	pgid, err := syscall.Getpgid(mp.cmd.Process.Pid)
	if err != nil {
		// Fall back to killing just the process
		return mp.cmd.Process.Kill()
	}

	// Send SIGTERM to the process group
	// Ignore error - process might already be dead (ESRCH)
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
		// Force kill
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		<-done
		return nil
	}
}

// Wait waits for the process to exit and returns any error.
func (mp *ManagedProcess) Wait() error {
	if mp.cmd == nil {
		return nil
	}
	return mp.cmd.Wait()
}

// Pid returns the process ID, or 0 if not started.
func (mp *ManagedProcess) Pid() int {
	if mp.cmd == nil || mp.cmd.Process == nil {
		return 0
	}
	return mp.cmd.Process.Pid
}
