// Package telemetry provides fail-safe analytics using an interface-based design.
// If API keys are missing or YAPI_NO_ANALYTICS is set, all operations become no-ops.
package telemetry

import (
	"os"
)

// Set at build time via ldflags:
// go build -ldflags "-X yapi.run/cli/internal/telemetry.PosthogAPIKey=... -X yapi.run/cli/internal/telemetry.PosthogAPIHost=..."
var (
	PosthogAPIKey  string
	PosthogAPIHost string
)

// Init initializes the telemetry client.
// Should be called once at startup with version info.
// Respects YAPI_NO_ANALYTICS to disable and YAPI_PRINT_ANALYTICS to print events.
func Init(version, commit string) {
	// impl defaults to NoopClient, so we only need to upgrade if conditions are met

	// Check for opt-out first
	if os.Getenv("YAPI_NO_ANALYTICS") != "" {
		return // impl stays NoopClient
	}

	// Disable if keys not set at build time
	if PosthogAPIKey == "" || PosthogAPIHost == "" {
		return // impl stays NoopClient
	}

	// Check for debug printing
	printDebug := os.Getenv("YAPI_PRINT_ANALYTICS") != ""

	// Initialize real backend
	client, err := NewPostHogClient(PosthogAPIKey, PosthogAPIHost, version, commit, printDebug)
	if err != nil {
		return // impl stays NoopClient
	}

	impl = client
}
