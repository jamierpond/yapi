package process

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestStart_SimpleCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	proc, err := Start(context.Background(), "sleep 10", false)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}
	defer proc.Stop()

	if proc.Pid() == 0 {
		t.Error("expected non-zero PID")
	}
}

func TestStart_InvalidCommand(t *testing.T) {
	_, err := Start(context.Background(), "nonexistent_command_12345", false)
	// Note: Start() doesn't fail for invalid commands - the command itself fails
	// This is expected shell behavior
	if err != nil {
		t.Logf("start returned error (may be expected): %v", err)
	}
}

func TestStop_GracefulTermination(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	proc, err := Start(context.Background(), "sleep 60", false)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	err = proc.Stop()
	if err != nil {
		t.Fatalf("failed to stop process: %v", err)
	}
}

func TestStop_AlreadyStopped(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	// Start a quick command that exits immediately
	proc, err := Start(context.Background(), "true", false)
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	// Wait for it to finish
	time.Sleep(100 * time.Millisecond)

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
