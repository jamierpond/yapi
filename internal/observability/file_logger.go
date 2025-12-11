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
}

// NewFileLoggerClient creates a new file logger client
func NewFileLoggerClient(path, version, commit string) (*FileLoggerClient, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileLoggerClient{
		file:    file,
		version: version,
		commit:  commit,
	}, nil
}

func (f *FileLoggerClient) Track(event string, props map[string]any) {
	f.mu.Lock()
	defer f.mu.Unlock()

	entry := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"event":     event,
		"os":        runtime.GOOS,
		"arch":      runtime.GOARCH,
		"version":   f.version,
		"commit":    f.commit,
	}
	for k, v := range props {
		entry[k] = v
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		return
	}

	fmt.Fprintln(f.file, string(jsonBytes))
}

func (f *FileLoggerClient) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.file.Close()
}
