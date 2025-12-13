package observability

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// FileLoggerClient logs events to a file
type FileLoggerClient struct {
	file    *os.File
	version string
	commit  string
	mu      sync.Mutex
	events  []map[string]any
}

// NewFileLoggerClient creates a new file logger client
func NewFileLoggerClient(path, version, commit string) (*FileLoggerClient, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil { //nolint:gosec // user config directory
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //nolint:gosec // user-provided log path
	if err != nil {
		return nil, err
	}
	return &FileLoggerClient{
		file:    file,
		version: version,
		commit:  commit,
	}, nil
}

// Track records an event with properties to be written on Close.
func (f *FileLoggerClient) Track(event string, props map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()

	entry := map[string]any{
		"event": event,
	}
	for k, v := range props {
		entry[k] = v
	}
	f.events = append(f.events, entry)
}

// Close merges all tracked events and writes them to the log file.
func (f *FileLoggerClient) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.events) == 0 {
		return f.file.Close()
	}

	// Merge all events into a single entry
	merged := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"version":   f.version,
		"commit":    f.commit,
	}

	// Merge props from all events (later events override earlier)
	for _, ev := range f.events {
		for k, v := range ev {
			if k != "event" {
				merged[k] = v
			}
		}
	}

	// Collect event names
	var events []string
	for _, ev := range f.events {
		if name, ok := ev["event"].(string); ok {
			events = append(events, name)
		}
	}
	merged["events"] = events

	jsonBytes, err := json.Marshal(merged)
	if err != nil {
		return f.file.Close()
	}

	_, _ = fmt.Fprintln(f.file, string(jsonBytes))
	return f.file.Close()
}
