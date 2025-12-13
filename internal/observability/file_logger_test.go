package observability

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFileLoggerClient_MergesEvents(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "history.json")

	client, err := NewFileLoggerClient(path, "v1.0.0", "abc123")
	if err != nil {
		t.Fatalf("NewFileLoggerClient: %v", err)
	}

	// Track multiple events
	client.Track("command", map[string]any{
		"command":  "yapi run foo.yml",
		"from_tui": true,
	})
	client.Track("cmd_run", map[string]any{
		"args_count":  1,
		"duration_ms": 100,
	})
	client.Track("request_executed", map[string]any{
		"is_chain":    true,
		"transport":   "http",
		"duration_ms": 150, // should override cmd_run's duration_ms
	})

	if err := client.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Read and parse the file
	data, err := os.ReadFile(path) //nolint:gosec // test temp file
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	// Check merged fields
	if entry["command"] != "yapi run foo.yml" {
		t.Errorf("command = %v, want 'yapi run foo.yml'", entry["command"])
	}
	if entry["from_tui"] != true {
		t.Errorf("from_tui = %v, want true", entry["from_tui"])
	}
	if entry["is_chain"] != true {
		t.Errorf("is_chain = %v, want true", entry["is_chain"])
	}
	if entry["transport"] != "http" {
		t.Errorf("transport = %v, want 'http'", entry["transport"])
	}

	// Later event should override earlier
	if entry["duration_ms"] != float64(150) {
		t.Errorf("duration_ms = %v, want 150", entry["duration_ms"])
	}

	// Check events array
	events, ok := entry["events"].([]any)
	if !ok {
		t.Fatalf("events not an array: %T", entry["events"])
	}
	if len(events) != 3 {
		t.Errorf("len(events) = %d, want 3", len(events))
	}

	// Check static fields
	if entry["version"] != "v1.0.0" {
		t.Errorf("version = %v, want 'v1.0.0'", entry["version"])
	}
	if entry["commit"] != "abc123" {
		t.Errorf("commit = %v, want 'abc123'", entry["commit"])
	}
	if entry["timestamp"] == nil {
		t.Error("timestamp is nil")
	}
}

func TestFileLoggerClient_NoEventsNoWrite(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "history.json")

	client, err := NewFileLoggerClient(path, "v1.0.0", "abc123")
	if err != nil {
		t.Fatalf("NewFileLoggerClient: %v", err)
	}

	// Close without tracking anything
	if err := client.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// File should exist but be empty
	data, err := os.ReadFile(path) //nolint:gosec // test temp file
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty file, got %d bytes", len(data))
	}
}
