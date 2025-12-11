package observability

// Provider defines the behavior for any observability backend
type Provider interface {
	Track(event string, props map[string]any)
	Close() error
}

// providers holds all registered observability backends
var providers []Provider

// Track sends an event to all registered providers
func Track(event string, props map[string]any) {
	for _, p := range providers {
		p.Track(event, props)
	}
}

// Close flushes all providers
func Close() {
	for _, p := range providers {
		_ = p.Close()
	}
}

// Enabled returns true if any providers are registered
func Enabled() bool {
	return len(providers) > 0
}

// AddProvider registers a new observability provider
func AddProvider(p Provider) {
	providers = append(providers, p)
}
