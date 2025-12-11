// Package observability provides fail-safe analytics using an interface-based design.
// If YAPI_NO_ANALYTICS is set, all operations become no-ops.
package observability

import (
	"os"
	"path/filepath"
)

// Default log file path
var LogFilePath = filepath.Join(os.Getenv("HOME"), "yapi-log.txt")

// Init initializes the observability client.
// Should be called once at startup with version info.
// Respects YAPI_NO_ANALYTICS env var and user config preference.
func Init(version, commit string) {
	// impl defaults to NoopClient, so we only need to upgrade if conditions are met

	// Environment variable opt-out takes highest priority
	if os.Getenv("YAPI_NO_ANALYTICS") != "" {
		return // impl stays NoopClient
	}

	// Check user's saved preference from config.json
	// Default to disabled unless user explicitly enabled
	pref := GetObservabilityPreference()
	if pref == nil || !*pref {
		return // No preference or explicitly disabled = no observability
	}

	// Initialize file logger
	client, err := NewFileLoggerClient(LogFilePath, version, commit)
	if err != nil {
		return // impl stays NoopClient
	}

	impl = client
}
