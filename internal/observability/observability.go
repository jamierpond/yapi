// Package observability provides local file logging for usage stats.
// If YAPI_NO_ANALYTICS is set, all operations become no-ops.
package observability

import (
	"os"
	"path/filepath"
)

// LogDir is the yapi data directory
var LogDir = filepath.Join(os.Getenv("HOME"), ".yapi")

// LogFileName is the log file name
const LogFileName = "log.json"

// LogFilePath is the full path to the log file
var LogFilePath = filepath.Join(LogDir, LogFileName)

// HistoryFileName is the history file name
const HistoryFileName = "history.json"

// HistoryFilePath is the full path to the history file
var HistoryFilePath = filepath.Join(LogDir, HistoryFileName)

// Init initializes observability (file logging only).
// Should be called once at startup with version info.
// Respects YAPI_NO_ANALYTICS env var.
func Init(version, commit string) {
	// Environment variable opt-out takes highest priority
	if os.Getenv("YAPI_NO_ANALYTICS") != "" {
		return
	}

	// Enable file logging
	if fileLogger, err := NewFileLoggerClient(LogFilePath, version, commit); err == nil {
		AddProvider(fileLogger)
	}
}
