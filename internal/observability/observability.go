// Package observability provides fail-safe analytics using an interface-based design.
// If YAPI_NO_ANALYTICS is set, all operations become no-ops.
package observability

import (
	"os"
	"path/filepath"
)

// Default log file path
var LogFilePath = filepath.Join(os.Getenv("HOME"), "yapi-log.txt")

// PostHog config - set at build time via ldflags
var (
	PosthogAPIKey  string
	PosthogAPIHost string
)

// Init initializes observability providers.
// Should be called once at startup with version info.
// Respects YAPI_NO_ANALYTICS env var and user config preferences.
func Init(version, commit string) {
	// Environment variable opt-out takes highest priority
	if os.Getenv("YAPI_NO_ANALYTICS") != "" {
		return
	}

	// Add file logger if enabled
	if pref := GetFileLoggingPreference(); pref != nil && *pref {
		if fileLogger, err := NewFileLoggerClient(LogFilePath, version, commit); err == nil {
			AddProvider(fileLogger)
		}
	}

	// Add PostHog if enabled and keys are set
	if pref := GetPosthogPreference(); pref != nil && *pref {
		if PosthogAPIKey != "" && PosthogAPIHost != "" {
			printDebug := os.Getenv("YAPI_PRINT_ANALYTICS") != ""
			if posthog, err := NewPostHogClient(PosthogAPIKey, PosthogAPIHost, version, commit, printDebug); err == nil {
				AddProvider(posthog)
			}
		}
	}
}
