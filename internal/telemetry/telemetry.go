// Package telemetry provides fail-safe analytics using Context propagation.
// If API keys are missing or YAPI_NO_ANALYTICS is set, all operations become no-ops.
package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/posthog/posthog-go"
)

// Set at build time via ldflags:
// go build -ldflags "-X yapi.run/cli/internal/telemetry.PosthogAPIKey=... -X yapi.run/cli/internal/telemetry.PosthogAPIHost=..."
var (
	PosthogAPIKey  string
	PosthogAPIHost string
)

var (
	client         posthog.Client
	disabled       bool
	printAnalytics bool
	appVersion     string
	appCommit      string
)

// Init initializes the telemetry client.
// Should be called once at startup with version info.
// Respects YAPI_NO_ANALYTICS to disable and YAPI_PRINT_ANALYTICS to print events.
func Init(version, commit string) {
	appVersion = version
	appCommit = commit

	// Check for opt-out first
	if os.Getenv("YAPI_NO_ANALYTICS") != "" {
		disabled = true
		return
	}

	// Disable if keys not set at build time
	if PosthogAPIKey == "" || PosthogAPIHost == "" {
		disabled = true
		return
	}

	// Check for debug printing
	if os.Getenv("YAPI_PRINT_ANALYTICS") != "" {
		printAnalytics = true
	}

	var err error
	client, err = posthog.NewWithConfig(PosthogAPIKey, posthog.Config{
		Endpoint: PosthogAPIHost,
	})
	if err != nil {
		// Silently fail - analytics should never break the CLI
		disabled = true
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

// Enabled returns true if telemetry is active.
func Enabled() bool {
	return !disabled
}

// Track sends an event to PostHog with the given properties.
// This is a fire-and-forget function that never returns errors.
func Track(event string, props map[string]interface{}) {
	if disabled {
		return
	}

	// Build final properties with standard fields
	finalProps := map[string]interface{}{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"version": appVersion,
		"commit":  appCommit,
	}
	for k, v := range props {
		finalProps[k] = v
	}

	// Print analytics event if enabled
	if printAnalytics {
		out := map[string]interface{}{
			"event":      event,
			"properties": finalProps,
		}
		jsonBytes, _ := json.MarshalIndent(out, "", "  ")
		fmt.Fprintf(os.Stderr, "[telemetry] %s\n", jsonBytes)
	}

	if client == nil {
		return
	}

	phProps := posthog.NewProperties()
	for k, v := range finalProps {
		phProps.Set(k, v)
	}

	distinctID := getMachineID()

	client.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      event,
		Properties: phProps,
	})
}

// contextKey is the key type for storing Trace in context
type contextKey struct{}

// Trace holds timing and property information for a traced operation
type Trace struct {
	Command   string
	StartTime time.Time
	Props     map[string]interface{}
}

// WithContext starts a new trace and returns a context containing it.
func WithContext(ctx context.Context, command string, props map[string]interface{}) context.Context {
	if props == nil {
		props = make(map[string]interface{})
	}
	return context.WithValue(ctx, contextKey{}, &Trace{
		Command:   command,
		StartTime: time.Now(),
		Props:     props,
	})
}

// EndFromContext completes the trace stored in context and emits the event.
func EndFromContext(ctx context.Context, err error) {
	tr, ok := ctx.Value(contextKey{}).(*Trace)
	if !ok {
		return
	}

	tr.Props["duration_ms"] = time.Since(tr.StartTime).Milliseconds()
	tr.Props["success"] = (err == nil)
	if err != nil {
		tr.Props["error"] = err.Error()
	}
	Track(tr.Command, tr.Props)
}

