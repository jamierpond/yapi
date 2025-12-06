package analytics

import "time"

// RequestStats holds feature usage data for a single request/chain execution.
type RequestStats struct {
	// From config (collected before execution)
	Transport       string
	IsChain         bool
	ChainStepCount  int
	HasExpectations bool
	AssertionCount  int
	HasStatusCheck  bool
	UsesChainVars   bool
	UsesEnvVars     bool

	// From execution (collected after)
	Success    bool
	DurationMs int64
	ErrorType  string // e.g. "validation", "network", "assertion_failed"
}

// TrackRequest sends a request_executed event with feature usage stats.
func TrackRequest(stats RequestStats) {
	props := map[string]interface{}{
		"transport":        stats.Transport,
		"is_chain":         stats.IsChain,
		"chain_step_count": stats.ChainStepCount,
		"has_expectations": stats.HasExpectations,
		"assertion_count":  stats.AssertionCount,
		"has_status_check": stats.HasStatusCheck,
		"uses_chain_vars":  stats.UsesChainVars,
		"uses_env_vars":    stats.UsesEnvVars,
		"success":          stats.Success,
		"duration_ms":      stats.DurationMs,
	}

	if stats.ErrorType != "" {
		props["error_type"] = stats.ErrorType
	}

	Track("request_executed", props)
}

// RequestTracker helps collect stats and track on completion.
type RequestTracker struct {
	Stats     RequestStats
	StartTime time.Time
}

// NewRequestTracker creates a tracker with the start time set.
func NewRequestTracker() *RequestTracker {
	return &RequestTracker{
		StartTime: time.Now(),
	}
}

// Complete finalizes the stats and sends the event.
func (rt *RequestTracker) Complete() {
	rt.Stats.DurationMs = time.Since(rt.StartTime).Milliseconds()
	TrackRequest(rt.Stats)
}

// SetSuccess marks the request as successful.
func (rt *RequestTracker) SetSuccess() {
	rt.Stats.Success = true
}

// SetError marks the request as failed with an error type.
func (rt *RequestTracker) SetError(errorType string) {
	rt.Stats.ErrorType = errorType
}

// SetStats sets the collected config stats.
func (rt *RequestTracker) SetStats(stats RequestStats) {
	// Preserve timing from tracker
	startTime := rt.StartTime
	rt.Stats = stats
	rt.StartTime = startTime
}
