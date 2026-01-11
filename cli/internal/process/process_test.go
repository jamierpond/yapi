package process

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"
)

// sleepCommand returns a platform-appropriate sleep command.
func sleepCommand(seconds int) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("powershell -Command Start-Sleep -Seconds %d", seconds)
	}
	return fmt.Sprintf("sleep %d", seconds)
}

// quickCommand returns a platform-appropriate command that exits quickly.
func quickCommand() string {
	if runtime.GOOS == "windows" {
		return "cmd /C echo done"
	}
	return "true"
}

func TestStart_SimpleCommand(t *testing.T) {
	proc, err := Start(context.Background(), sleepCommand(2), false)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}
	defer func() { _ = proc.Stop() }()

	if proc.Pid() == 0 {
		t.Error("expected non-zero PID")
	}
}

func TestStop_GracefulTermination(t *testing.T) {
	proc, err := Start(context.Background(), sleepCommand(2), false)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	err = proc.Stop()
	if err != nil {
		t.Fatalf("failed to stop process: %v", err)
	}
}

func TestStop_AlreadyStopped(t *testing.T) {
	// Start a quick command that exits immediately
	proc, err := Start(context.Background(), quickCommand(), false)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	// Wait for it to finish
	time.Sleep(200 * time.Millisecond)

	// Should not error when stopping already-stopped process
	err = proc.Stop()
	if err != nil {
		t.Fatalf("unexpected error stopping already-exited process: %v", err)
	}
}

func TestManagedProcess_NilCmd(t *testing.T) {
	mp := &ManagedProcess{}

	// Should handle nil cmd gracefully
	if mp.Pid() != 0 {
		t.Error("expected 0 PID for nil cmd")
	}

	err := mp.Stop()
	if err != nil {
		t.Fatalf("unexpected error stopping nil process: %v", err)
	}

	err = mp.Wait()
	if err != nil {
		t.Fatalf("unexpected error waiting on nil process: %v", err)
	}
}
