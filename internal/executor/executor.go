package executor

import (
	"context"
	"fmt"

	"yapi.run/cli/internal/domain"
)

// Executor is the interface all protocol executors must implement.
type Executor interface {
	Execute(ctx context.Context, req *domain.Request) (*domain.Response, error)
}

// Factory creates executors for different transports.
type Factory struct {
	httpClient HTTPClient
}

// NewFactory creates a new executor factory with the given HTTP client.
func NewFactory(httpClient HTTPClient) *Factory {
	return &Factory{httpClient: httpClient}
}

// Create returns the appropriate executor for the given transport.
// The returned executor is wrapped with timing middleware.
func (f *Factory) Create(transport string) (Executor, error) {
	var exec Executor

	switch transport {
	case "http":
		exec = NewHTTPExecutor(f.httpClient)
	case "graphql":
		exec = NewGraphQLExecutor(f.httpClient)
	case "grpc":
		exec = NewGRPCExecutor()
	case "tcp":
		exec = NewTCPExecutor()
	default:
		return nil, fmt.Errorf("unsupported transport: %s", transport)
	}

	return WithTiming(exec), nil
}
