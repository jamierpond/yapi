package telemetry

// Client defines the behavior for any analytics backend
type Client interface {
	Track(event string, props map[string]interface{})
	Close() error
}

// Global instance defaults to Noop so it is ALWAYS safe to call
var impl Client = &NoopClient{}

// Track is the public facade
func Track(event string, props map[string]interface{}) {
	impl.Track(event, props)
}

// Close flushes the backend
func Close() {
	_ = impl.Close()
}

// Enabled returns true if telemetry is active (not using NoopClient)
func Enabled() bool {
	_, isNoop := impl.(*NoopClient)
	return !isNoop
}
