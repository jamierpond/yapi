package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/posthog/posthog-go"
)

// PostHogClient wraps the real PostHog client
type PostHogClient struct {
	client     posthog.Client
	version    string
	commit     string
	printDebug bool
}

// NewPostHogClient creates a new PostHog client
func NewPostHogClient(apiKey, apiHost, version, commit string, printDebug bool) (*PostHogClient, error) {
	client, err := posthog.NewWithConfig(apiKey, posthog.Config{
		Endpoint: apiHost,
	})
	if err != nil {
		return nil, err
	}
	return &PostHogClient{
		client:     client,
		version:    version,
		commit:     commit,
		printDebug: printDebug,
	}, nil
}

func (p *PostHogClient) Track(event string, props map[string]interface{}) {
	// Build final properties with standard fields
	finalProps := map[string]interface{}{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"version": p.version,
		"commit":  p.commit,
	}
	for k, v := range props {
		finalProps[k] = v
	}

	// Print analytics event if debug enabled
	if p.printDebug {
		out := map[string]interface{}{
			"event":      event,
			"properties": finalProps,
		}
		jsonBytes, _ := json.MarshalIndent(out, "", "  ")
		fmt.Fprintf(os.Stderr, "[telemetry] %s\n", jsonBytes)
	}

	phProps := posthog.NewProperties()
	for k, v := range finalProps {
		phProps.Set(k, v)
	}

	distinctID := getMachineID()

	p.client.Enqueue(posthog.Capture{
		DistinctId: distinctID,
		Event:      event,
		Properties: phProps,
	})
}

func (p *PostHogClient) Close() error {
	return p.client.Close()
}
