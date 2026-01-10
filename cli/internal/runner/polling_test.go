package runner_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/executor"
	"yapi.run/cli/internal/runner"
)

func TestRunWithPolling_SuccessOnFirstAttempt(t *testing.T) {
	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       io.NopCloser(strings.NewReader(`{"status": "completed"}`)),
			Duration:   10 * time.Millisecond,
		}, nil
	}

	req := &domain.Request{URL: "http://example.com/api", Method: "GET"}
	waitFor := &config.WaitFor{
		Until:   []string{`.status == "completed"`},
		Period:  "100ms",
		Timeout: "5s",
	}

	var execFn executor.TransportFunc = mockTransport
	result, err := runner.RunWithPolling(context.Background(), execFn, req, waitFor, nil, runner.Options{}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", result.Attempts)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", result.StatusCode)
	}
}

func TestRunWithPolling_SuccessAfterRetries(t *testing.T) {
	attempts := 0
	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		attempts++
		status := "pending"
		if attempts >= 3 {
			status = "completed"
		}
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       io.NopCloser(strings.NewReader(`{"status": "` + status + `"}`)),
			Duration:   10 * time.Millisecond,
		}, nil
	}

	req := &domain.Request{URL: "http://example.com/api", Method: "GET"}
	waitFor := &config.WaitFor{
		Until:   []string{`.status == "completed"`},
		Period:  "10ms",
		Timeout: "5s",
	}

	var execFn executor.TransportFunc = mockTransport
	result, err := runner.RunWithPolling(context.Background(), execFn, req, waitFor, nil, runner.Options{}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", result.Attempts)
	}
}

func TestRunWithPolling_Timeout(t *testing.T) {
	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       io.NopCloser(strings.NewReader(`{"status": "pending"}`)),
			Duration:   10 * time.Millisecond,
		}, nil
	}

	req := &domain.Request{URL: "http://example.com/api", Method: "GET"}
	waitFor := &config.WaitFor{
		Until:   []string{`.status == "completed"`},
		Period:  "50ms",
		Timeout: "100ms",
	}

	var execFn executor.TransportFunc = mockTransport
	_, err := runner.RunWithPolling(context.Background(), execFn, req, waitFor, nil, runner.Options{}, nil)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

func TestRunWithPolling_RequestError(t *testing.T) {
	attempts := 0
	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("connection refused")
		}
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       io.NopCloser(strings.NewReader(`{"status": "completed"}`)),
			Duration:   10 * time.Millisecond,
		}, nil
	}

	req := &domain.Request{URL: "http://example.com/api", Method: "GET"}
	waitFor := &config.WaitFor{
		Until:   []string{`.status == "completed"`},
		Period:  "10ms",
		Timeout: "5s",
	}

	var execFn executor.TransportFunc = mockTransport
	result, err := runner.RunWithPolling(context.Background(), execFn, req, waitFor, nil, runner.Options{}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", result.Attempts)
	}
}

func TestRunWithPolling_Backoff(t *testing.T) {
	attempts := 0
	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		attempts++
		status := "pending"
		if attempts >= 3 {
			status = "completed"
		}
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       io.NopCloser(strings.NewReader(`{"status": "` + status + `"}`)),
			Duration:   10 * time.Millisecond,
		}, nil
	}

	req := &domain.Request{URL: "http://example.com/api", Method: "GET"}
	waitFor := &config.WaitFor{
		Until: []string{`.status == "completed"`},
		Backoff: &config.Backoff{
			Seed:       "10ms",
			Multiplier: 2,
		},
		Timeout: "5s",
	}

	var execFn executor.TransportFunc = mockTransport
	result, err := runner.RunWithPolling(context.Background(), execFn, req, waitFor, nil, runner.Options{}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", result.Attempts)
	}
}

func TestRunWithPolling_ValidationErrors(t *testing.T) {
	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{},
			Body:       io.NopCloser(strings.NewReader(`{}`)),
		}, nil
	}

	req := &domain.Request{URL: "http://example.com/api", Method: "GET"}
	var execFn executor.TransportFunc = mockTransport

	tests := []struct {
		name    string
		waitFor *config.WaitFor
		wantErr string
	}{
		{
			name:    "empty until",
			waitFor: &config.WaitFor{Until: []string{}, Period: "1s", Timeout: "5s"},
			wantErr: "until is required",
		},
		{
			name:    "both period and backoff",
			waitFor: &config.WaitFor{Until: []string{".x"}, Period: "1s", Backoff: &config.Backoff{Seed: "1s", Multiplier: 2}, Timeout: "5s"},
			wantErr: "mutually exclusive",
		},
		{
			name:    "neither period nor backoff",
			waitFor: &config.WaitFor{Until: []string{".x"}, Timeout: "5s"},
			wantErr: "either period or backoff",
		},
		{
			name:    "invalid period",
			waitFor: &config.WaitFor{Until: []string{".x"}, Period: "invalid", Timeout: "5s"},
			wantErr: "invalid wait_for.period",
		},
		{
			name:    "invalid timeout",
			waitFor: &config.WaitFor{Until: []string{".x"}, Period: "1s", Timeout: "invalid"},
			wantErr: "invalid wait_for.timeout",
		},
		{
			name:    "invalid backoff seed",
			waitFor: &config.WaitFor{Until: []string{".x"}, Backoff: &config.Backoff{Seed: "invalid", Multiplier: 2}, Timeout: "5s"},
			wantErr: "invalid wait_for.backoff.seed",
		},
		{
			name:    "backoff multiplier <= 1",
			waitFor: &config.WaitFor{Until: []string{".x"}, Backoff: &config.Backoff{Seed: "1s", Multiplier: 1}, Timeout: "5s"},
			wantErr: "multiplier must be > 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runner.RunWithPolling(context.Background(), execFn, req, tt.waitFor, nil, runner.Options{}, nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestRunWithPolling_BodyPreservedAcrossAttempts(t *testing.T) {
	// This test verifies that request body is available on every polling attempt.
	// Previously, the body io.Reader was exhausted after the first attempt,
	// causing subsequent attempts to send empty bodies (breaking gRPC polling).
	expectedBody := `{"query": "test"}`
	attempts := 0

	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		attempts++

		// Read and verify body on each attempt
		var bodyContent string
		if req.Body != nil {
			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatalf("attempt %d: failed to read body: %v", attempts, err)
			}
			bodyContent = string(bodyBytes)
		}

		if bodyContent != expectedBody {
			t.Errorf("attempt %d: expected body %q, got %q", attempts, expectedBody, bodyContent)
		}

		// Return "pending" for first 2 attempts, then "completed"
		status := "pending"
		if attempts >= 3 {
			status = "completed"
		}
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       io.NopCloser(strings.NewReader(`{"status": "` + status + `"}`)),
			Duration:   10 * time.Millisecond,
		}, nil
	}

	req := &domain.Request{
		URL:    "http://example.com/api",
		Method: "POST",
		Body:   strings.NewReader(expectedBody),
	}
	waitFor := &config.WaitFor{
		Until:   []string{`.status == "completed"`},
		Period:  "10ms",
		Timeout: "5s",
	}

	var execFn executor.TransportFunc = mockTransport
	result, err := runner.RunWithPolling(context.Background(), execFn, req, waitFor, nil, runner.Options{}, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", result.Attempts)
	}
}

func TestRunWithPolling_ContextCancellation(t *testing.T) {
	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       io.NopCloser(strings.NewReader(`{"status": "pending"}`)),
			Duration:   10 * time.Millisecond,
		}, nil
	}

	req := &domain.Request{URL: "http://example.com/api", Method: "GET"}
	waitFor := &config.WaitFor{
		Until:   []string{`.status == "completed"`},
		Period:  "100ms",
		Timeout: "10s",
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	var execFn executor.TransportFunc = mockTransport
	_, err := runner.RunWithPolling(ctx, execFn, req, waitFor, nil, runner.Options{}, nil)

	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}
