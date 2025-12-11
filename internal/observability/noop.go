package observability

// NoopClient is the default client that does nothing.
// Used when:
// - The user sets YAPI_NO_ANALYTICS=1
// - Initialization fails
type NoopClient struct{}

func (n *NoopClient) Track(event string, props map[string]interface{}) {
	// Intentionally empty - the compiler will inline this
}

func (n *NoopClient) Close() error {
	return nil
}
