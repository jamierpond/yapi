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
func (f *Factory) Create(transport string) (Executor, error) {
	switch transport {
	case "http":
		return NewHTTPExecutor(f.httpClient), nil
	case "graphql":
		return NewGraphQLExecutor(f.httpClient), nil
	case "grpc":
		return NewGRPCExecutor(), nil
	case "tcp":
		return NewTCPExecutor(), nil
	default:
		return nil, fmt.Errorf("unsupported transport: %s", transport)
	}
}
