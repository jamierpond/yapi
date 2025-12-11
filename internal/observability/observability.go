// Package observability provides fail-safe analytics using an interface-based design.
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

// PostHog config - set at build time via ldflags
var (
	PosthogAPIKey  string
	PosthogAPIHost string
)

// Init initializes observability providers based on the telemetry mode.
// Should be called once at startup with version info.
// Respects YAPI_NO_ANALYTICS env var and user config preferences.
func Init(version, commit string) {
	// Environment variable opt-out takes highest priority
	if os.Getenv("YAPI_NO_ANALYTICS") != "" {
		return
	}

	mode := GetMode()

	// Mode "off" disables all telemetry
	if mode == ModeOff {
		return
	}

	// File logging is enabled for both "local" and "on" modes
	if mode == ModeLocal || mode == ModeOn {
		if fileLogger, err := NewFileLoggerClient(LogFilePath, version, commit); err == nil {
			AddProvider(fileLogger)
		}
	}

	// PostHog is only enabled when mode is "on"
	if mode == ModeOn {
		if PosthogAPIKey != "" && PosthogAPIHost != "" {
			printDebug := os.Getenv("YAPI_PRINT_ANALYTICS") != ""
			if posthog, err := NewPostHogClient(PosthogAPIKey, PosthogAPIHost, version, commit, printDebug); err == nil {
				AddProvider(posthog)
			}
		}
	}
}
