// Package observability provides local file logging for usage stats.
// If YAPI_NO_ANALYTICS is set, all operations become no-ops.
package observability

import (
	"os"
	"path/filepath"
)

// Log file name constant
const LogFileName = "yapi.log"

// Default log file path
var LogFilePath = filepath.Join(os.Getenv("HOME"), LogFileName)

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
