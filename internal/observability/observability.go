// Package observability provides fail-safe analytics using an interface-based design.
// If YAPI_NO_ANALYTICS is set, all operations become no-ops.
package observability

import (
	"os"
	"path/filepath"
)

// Default log file path
var LogFilePath = filepath.Join(os.Getenv("HOME"), "yapi-log.txt")

// Init initializes observability providers.
// Should be called once at startup with version info.
// Respects YAPI_NO_ANALYTICS env var and user config preference.
func Init(version, commit string) {
	// Environment variable opt-out takes highest priority
	if os.Getenv("YAPI_NO_ANALYTICS") != "" {
		return
	}

	// Check user's saved preference from config.json
	// Default to disabled unless user explicitly enabled
	pref := GetObservabilityPreference()
	if pref == nil || !*pref {
		return
	}

	// Add file logger provider
	if fileLogger, err := NewFileLoggerClient(LogFilePath, version, commit); err == nil {
		AddProvider(fileLogger)
	}

	// Future: Add PostHog provider here
	// if PosthogAPIKey != "" && PosthogAPIHost != "" {
	//     if posthog, err := NewPostHogClient(...); err == nil {
	//         AddProvider(posthog)
	//     }
	// }
}
