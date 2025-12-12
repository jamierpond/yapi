// Package observability provides local file logging.
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

// Init initializes observability (file logging).
// Should be called once at startup with version info.
func Init(version, commit string) {
	if fileLogger, err := NewFileLoggerClient(LogFilePath, version, commit); err == nil {
		AddProvider(fileLogger)
	}
}
