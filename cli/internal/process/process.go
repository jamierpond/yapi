// Package process provides managed process lifecycle for spawning and cleaning up subprocesses.
package process

import (
	"context"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

// ManagedProcess wraps an exec.Cmd with lifecycle management.
type ManagedProcess struct {
	cmd      *exec.Cmd
	stdout   io.ReadCloser
	stderr   io.ReadCloser
	stopOnce sync.Once
}

// Start spawns a new process using the shell.
// The process runs in the background and can be stopped with Stop().
// If verbose is true, stdout/stderr are piped to os.Stdout/os.Stderr.
func Start(ctx context.Context, command string, verbose bool) (*ManagedProcess, error) {
	cmd := buildCommand(ctx, command)

	// Platform-specific setup (process groups on Unix, no-op on Windows)
	configurePlatform(cmd)

	mp := &ManagedProcess{cmd: cmd}

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
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

// buildCommand creates the appropriate shell command for the platform.
func buildCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/C", command)
	}
	return exec.CommandContext(ctx, "sh", "-c", command)
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
