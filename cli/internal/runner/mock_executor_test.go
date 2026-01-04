package runner_test

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/runner"
)

func TestRun(t *testing.T) {
	// 1. Setup mock transport function
	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		// Simulate a successful response
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       io.NopCloser(strings.NewReader(`{"message": "success"}`)),
			Duration:   100 * time.Millisecond,
		}, nil
	}

	// 2. Create a sample domain request
	req := &domain.Request{
		URL:    "http://example.com/api",
		Method: "GET",
	}

	// 3. Define runner options
	opts := runner.Options{
		NoColor: true,
	}

	// 4. Call the runner with the mock transport function
	var execFn executor.TransportFunc = mockTransport
	result, err := runner.Run(context.Background(), execFn, req, nil, opts)

	// 5. Assert the results
	if err != nil {
		t.Fatalf("runner.Run() returned an unexpected error: %v", err)
	}

	if result.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", result.StatusCode)
	}

	expectedBody := `{"message": "success"}`
	if result.Body != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, result.Body)
	}

	if result.ContentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", result.ContentType)
	}
}
