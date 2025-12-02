package executor

import (
	"context"
	"fmt"
	"net/http"

	"yapi.run/cli/internal/domain"
)

// HTTPClient is an interface for a client that can send HTTP requests.
// It's implemented by *http.Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPExecutor handles HTTP requests.
type HTTPExecutor struct {
	client HTTPClient
}

// NewHTTPExecutor creates a new HTTPExecutor.
func NewHTTPExecutor(client HTTPClient) *HTTPExecutor {
	return &HTTPExecutor{client: client}
}

// Execute performs an HTTP request based on the provided domain.Request.
func (e *HTTPExecutor) Execute(ctx context.Context, req *domain.Request) (*domain.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set custom headers
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	res, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Convert http.Header to map[string]string
	headers := make(map[string]string)
	for k, v := range res.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &domain.Response{
		StatusCode: res.StatusCode,
		Headers:    headers,
		Body:       res.Body,
	}, nil
}
