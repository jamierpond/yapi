package analytics

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/posthog/posthog-go"
)

const (
	posthogAPIKey  = "phc_5gccNEZpJamQIIdMx97VQA1wuuDvqgISlLcSwoAGoeX"
	posthogAPIHost = "https://us.i.posthog.com"
)

var (
	client          posthog.Client
	disabled        bool
	printAnalytics  bool
	currentTracker  *CommandTracker
)

// CommandTracker tracks timing for a command execution.
type CommandTracker struct {
	command   string
	version   string
	startTime time.Time
}

// Init initializes the PostHog analytics client.
// Should be called once at startup.
// Respects YAPI_NO_ANALYTICS to disable and YAPI_PRINT_ANALYTICS to print events.
func Init() {
	// Check for opt-out
	if os.Getenv("YAPI_NO_ANALYTICS") != "" {
		disabled = true
		return
	}

	// Check for debug printing
	if os.Getenv("YAPI_PRINT_ANALYTICS") != "" {
		printAnalytics = true
	}

	var err error
	client, err = posthog.NewWithConfig(posthogAPIKey, posthog.Config{
		Endpoint: posthogAPIHost,
	})
	if err != nil {
		// Silently fail - analytics should never break the CLI
		return
	}
}

// Close flushes and closes the PostHog client.
// Should be called before the program exits.
func Close() {
	if client != nil {
		client.Close()
	}
}

// StartCommand begins tracking a command execution.
func StartCommand(command, version string) *CommandTracker {
	ct := &CommandTracker{
		command:   command,
		version:   version,
		startTime: time.Now(),
	}

	Track("command_started", map[string]interface{}{
		"command": command,
		"version": version,
	})

	return ct
}

// End completes the command tracking with success/failure status.
func (ct *CommandTracker) End(success bool, errorMsg string) {
	if ct == nil {
		return
	}

	duration := time.Since(ct.startTime)

	props := map[string]interface{}{
		"command":     ct.command,
		"version":     ct.version,
		"duration_ms": duration.Milliseconds(),
		"success":     success,
	}

	if errorMsg != "" {
		props["error"] = errorMsg
	}

	Track("command_completed", props)
}

// Track sends an event to PostHog.
func Track(event string, properties map[string]interface{}) {
	if disabled {
		return
	}

	distinctID := getDistinctID()

	// Add standard properties
	allProps := make(map[string]interface{})
	for k, v := range properties {
		allProps[k] = v
	}
	allProps["os"] = runtime.GOOS
	allProps["arch"] = runtime.GOARCH

	// Print analytics event if enabled
	if printAnalytics {
		out := map[string]interface{}{
			"event":       event,
			"distinct_id": distinctID,
			"properties":  allProps,
		}
		jsonBytes, _ := json.MarshalIndent(out, "", "  ")
		fmt.Fprintf(os.Stderr, "[analytics] %s\n", jsonBytes)
	}

	if client == nil {
		return
	}

	props := posthog.NewProperties()
	for k, v := range allProps {
		props.Set(k, v)
	}

	client.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      event,
		Properties: props,
	})
}

// getDistinctID returns a stable anonymous identifier for the machine.
func getDistinctID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}
