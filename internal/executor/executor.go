// Package executor provides transport implementations for HTTP, gRPC, TCP, and GraphQL.
package executor

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"yapi.run/cli/internal/constants"
	"yapi.run/cli/internal/domain"
)

// TransportFunc is the functional signature for all transport implementations.
type TransportFunc func(ctx context.Context, req *domain.Request) (*domain.Response, error)

// HTTPClient is an interface for a client that can send HTTP requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Factory creates transport functions for different protocols.
type Factory struct {
	Client HTTPClient
}

// NewFactory creates a new executor factory with the given HTTP client.
func NewFactory(client HTTPClient) *Factory {
	return &Factory{Client: client}
}

// Create returns the appropriate transport function for the given transport type.
// The returned function is wrapped with timing middleware.
func (f *Factory) Create(transport string) (TransportFunc, error) {
	var fn TransportFunc

	switch transport {
	case constants.TransportHTTP:
		fn = HTTPTransport(f.Client)
	case constants.TransportGraphQL:
		fn = GraphQLTransport(f.Client)
	case constants.TransportGRPC:
		fn = GRPCTransport
	case constants.TransportTCP:
		fn = TCPTransport
	default:
		return nil, fmt.Errorf("unsupported transport: %s", transport)
	}

	return WithTiming(fn), nil
}

// WithTiming wraps a transport function to measure execution duration.
func WithTiming(next TransportFunc) TransportFunc {
	return func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		start := time.Now()
		resp, err := next(ctx, req)
		if err != nil {
			return nil, err
		}
		resp.Duration = time.Since(start)
		return resp, err
	}
}
