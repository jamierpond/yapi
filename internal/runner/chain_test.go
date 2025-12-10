package runner

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/executor"
)

func TestCheckExpectations_Status(t *testing.T) {
	tests := []struct {
		name        string
		expectation config.Expectation
		result      *Result
		wantErr     bool
	}{
		{
			name:        "status matches (int)",
			expectation: config.Expectation{Status: 200},
			result:      &Result{StatusCode: 200},
			wantErr:     false,
		},
		{
			name:        "status matches (float64)",
			expectation: config.Expectation{Status: float64(200)},
			result:      &Result{StatusCode: 200},
			wantErr:     false,
		},
		{
			name:        "status does not match",
			expectation: config.Expectation{Status: 200},
			result:      &Result{StatusCode: 404},
			wantErr:     true,
		},
		{
			name:        "status in array matches",
			expectation: config.Expectation{Status: []interface{}{float64(200), float64(201)}},
			result:      &Result{StatusCode: 201},
			wantErr:     false,
		},
		{
			name:        "status not in array",
			expectation: config.Expectation{Status: []interface{}{float64(200), float64(201)}},
			result:      &Result{StatusCode: 404},
			wantErr:     true,
		},
		{
			name:        "no status expectation",
			expectation: config.Expectation{},
			result:      &Result{StatusCode: 500},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := CheckExpectations(tt.expectation, tt.result)
			if (res.Error != nil) != tt.wantErr {
				t.Errorf("CheckExpectations() error = %v, wantErr %v", res.Error, tt.wantErr)
			}
		})
	}
}

func TestCheckExpectations_Assert(t *testing.T) {
	tests := []struct {
		name        string
		expectation config.Expectation
		result      *Result
		wantErr     bool
	}{
		{
			name:        "assertion passes - contains check",
			expectation: config.Expectation{Assert: []string{`.status == "success"`}},
			result:      &Result{Body: `{"status": "success"}`},
			wantErr:     false,
		},
		{
			name:        "assertion fails - value mismatch",
			expectation: config.Expectation{Assert: []string{`.status == "error"`}},
			result:      &Result{Body: `{"status": "success"}`},
			wantErr:     true,
		},
		{
			name:        "assertion passes - field exists",
			expectation: config.Expectation{Assert: []string{`.status != null`}},
			result:      &Result{Body: `{"status": "success"}`},
			wantErr:     false,
		},
		{
			name:        "assertion fails - field missing",
			expectation: config.Expectation{Assert: []string{`.missing != null`}},
			result:      &Result{Body: `{"status": "success"}`},
			wantErr:     true,
		},
		{
			name:        "multiple assertions - all pass",
			expectation: config.Expectation{Assert: []string{`.status == "success"`, `.data == "test"`}},
			result:      &Result{Body: `{"status": "success", "data": "test"}`},
			wantErr:     false,
		},
		{
			name:        "multiple assertions - one fails",
			expectation: config.Expectation{Assert: []string{`.status == "success"`, `.data == "wrong"`}},
			result:      &Result{Body: `{"status": "success", "data": "test"}`},
			wantErr:     true,
		},
		{
			name:        "no assertions",
			expectation: config.Expectation{},
			result:      &Result{Body: "anything"},
			wantErr:     false,
		},
		{
			name:        "array length check",
			expectation: config.Expectation{Assert: []string{`.items | length > 0`}},
			result:      &Result{Body: `{"items": [1, 2, 3]}`},
			wantErr:     false,
		},
		{
			name:        "empty array fails length check",
			expectation: config.Expectation{Assert: []string{`.items | length > 0`}},
			result:      &Result{Body: `{"items": []}`},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := CheckExpectations(tt.expectation, tt.result)
			if (res.Error != nil) != tt.wantErr {
				t.Errorf("CheckExpectations() error = %v, wantErr %v", res.Error, tt.wantErr)
			}
		})
	}
}

func TestResolveVariableRaw(t *testing.T) {
	ctx := NewChainContext()
	ctx.Results["step1"] = StepResult{
		BodyJSON: map[string]interface{}{
			"result": map[string]interface{}{
				"index":   float64(7), // JSON numbers are float64
				"enabled": true,
				"ratio":   3.14,
				"name":    "test",
			},
		},
		StatusCode: 200,
	}

	tests := []struct {
		name    string
		input   string
		wantVal interface{}
		wantOk  bool
	}{
		{
			name:    "pure int reference",
			input:   "$step1.result.index",
			wantVal: float64(7),
			wantOk:  true,
		},
		{
			name:    "pure bool reference",
			input:   "$step1.result.enabled",
			wantVal: true,
			wantOk:  true,
		},
		{
			name:    "pure float reference",
			input:   "$step1.result.ratio",
			wantVal: 3.14,
			wantOk:  true,
		},
		{
			name:    "pure string reference",
			input:   "$step1.result.name",
			wantVal: "test",
			wantOk:  true,
		},
		{
			name:    "strict format reference",
			input:   "${step1.result.index}",
			wantVal: float64(7),
			wantOk:  true,
		},
		{
			name:    "mixed string not resolved",
			input:   "prefix-$step1.result.index",
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "env var not resolved",
			input:   "$HOME",
			wantVal: nil,
			wantOk:  false,
		},
		{
			name:    "no variable",
			input:   "plain text",
			wantVal: nil,
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := ctx.ResolveVariableRaw(tt.input)
			if ok != tt.wantOk {
				t.Errorf("ResolveVariableRaw() ok = %v, wantOk %v", ok, tt.wantOk)
				return
			}
			if ok && val != tt.wantVal {
				t.Errorf("ResolveVariableRaw() = %v (%T), want %v (%T)", val, val, tt.wantVal, tt.wantVal)
			}
		})
	}
}

func TestInterpolateBody(t *testing.T) {
	ctx := NewChainContext()
	ctx.Results["prev"] = StepResult{
		BodyJSON:   map[string]interface{}{"token": "abc123"},
		StatusCode: 200,
	}
	// Add step with typed values for type preservation tests
	ctx.Results["step1"] = StepResult{
		BodyJSON: map[string]interface{}{
			"result": map[string]interface{}{
				"index": float64(7),
			},
		},
		StatusCode: 200,
	}

	tests := []struct {
		name     string
		body     map[string]interface{}
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "nil body",
			body:     nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name: "simple string interpolation",
			body: map[string]interface{}{
				"auth": "${prev.token}",
			},
			expected: map[string]interface{}{
				"auth": "abc123",
			},
			wantErr: false,
		},
		{
			name: "non-string values unchanged",
			body: map[string]interface{}{
				"count": 42,
				"flag":  true,
			},
			expected: map[string]interface{}{
				"count": 42,
				"flag":  true,
			},
			wantErr: false,
		},
		{
			name: "nested body",
			body: map[string]interface{}{
				"data": map[string]interface{}{
					"token": "${prev.token}",
				},
			},
			expected: map[string]interface{}{
				"data": map[string]interface{}{
					"token": "abc123",
				},
			},
			wantErr: false,
		},
		{
			name: "type preservation - int",
			body: map[string]interface{}{
				"track_index": "$step1.result.index",
			},
			expected: map[string]interface{}{
				"track_index": float64(7), // Preserved as number, not string
			},
			wantErr: false,
		},
		{
			name: "mixed string stays string",
			body: map[string]interface{}{
				"message": "Track $step1.result.index created",
			},
			expected: map[string]interface{}{
				"message": "Track 7 created", // Interpolated as string
			},
			wantErr: false,
		},
		{
			name: "array with variable references",
			body: map[string]interface{}{
				"track_indices": []interface{}{
					"$step1.result.index",
					"$prev.token",
				},
			},
			expected: map[string]interface{}{
				"track_indices": []interface{}{
					float64(7),
					"abc123",
				},
			},
			wantErr: false,
		},
		{
			name: "nested array in object",
			body: map[string]interface{}{
				"params": map[string]interface{}{
					"indices": []interface{}{
						"$step1.result.index",
					},
				},
			},
			expected: map[string]interface{}{
				"params": map[string]interface{}{
					"indices": []interface{}{
						float64(7),
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := interpolateBody(ctx, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("interpolateBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Simple comparison for nil case
				if tt.expected == nil && result != nil {
					t.Errorf("expected nil, got %v", result)
					return
				}
				if tt.expected == nil {
					return
				}
				// Compare specific keys
				for k, expectedVal := range tt.expected {
					actualVal, ok := result[k]
					if !ok {
						t.Errorf("key '%s' not found in result", k)
						continue
					}
					// Handle nested maps
					if expectedNested, ok := expectedVal.(map[string]interface{}); ok {
						actualNested, ok := actualVal.(map[string]interface{})
						if !ok {
							t.Errorf("key '%s' expected map, got %T", k, actualVal)
							continue
						}
						for nk, nv := range expectedNested {
							// Handle arrays in nested maps
							if expectedArr, ok := nv.([]interface{}); ok {
								actualArr, ok := actualNested[nk].([]interface{})
								if !ok {
									t.Errorf("nested key '%s.%s' expected array, got %T", k, nk, actualNested[nk])
									continue
								}
								if len(actualArr) != len(expectedArr) {
									t.Errorf("nested key '%s.%s' array length = %d, want %d", k, nk, len(actualArr), len(expectedArr))
									continue
								}
								for i, ev := range expectedArr {
									if actualArr[i] != ev {
										t.Errorf("nested key '%s.%s[%d]' = %v, want %v", k, nk, i, actualArr[i], ev)
									}
								}
							} else if actualNested[nk] != nv {
								t.Errorf("nested key '%s.%s' = %v, want %v", k, nk, actualNested[nk], nv)
							}
						}
					} else if expectedArr, ok := expectedVal.([]interface{}); ok {
						// Handle arrays
						actualArr, ok := actualVal.([]interface{})
						if !ok {
							t.Errorf("key '%s' expected array, got %T", k, actualVal)
							continue
						}
						if len(actualArr) != len(expectedArr) {
							t.Errorf("key '%s' array length = %d, want %d", k, len(actualArr), len(expectedArr))
							continue
						}
						for i, ev := range expectedArr {
							if actualArr[i] != ev {
								t.Errorf("key '%s[%d]' = %v, want %v", k, i, actualArr[i], ev)
							}
						}
					} else if actualVal != expectedVal {
						t.Errorf("key '%s' = %v, want %v", k, actualVal, expectedVal)
					}
				}
			}
		})
	}
}

// mockExecutorFactory is a test helper that returns a configurable transport
type mockExecutorFactory struct {
	transport executor.TransportFunc
}

func (m *mockExecutorFactory) Create(transportType string) (executor.TransportFunc, error) {
	return m.transport, nil
}

func TestRunChain_Delay(t *testing.T) {
	// Test that delay waits before executing the request
	base := &config.ConfigV1{URL: "http://example.com"}
	steps := []config.ChainStep{
		{
			Name: "delayed_request",
			ConfigV1: config.ConfigV1{
				URL:   "http://example.com",
				Delay: "100ms",
			},
		},
	}

	transportCalled := false
	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		transportCalled = true
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       io.NopCloser(strings.NewReader(`{"status": "ok"}`)),
		}, nil
	}

	factory := &mockExecutorFactory{transport: mockTransport}

	start := time.Now()
	result, err := RunChain(context.Background(), factory, base, steps, Options{})
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("RunChain() returned unexpected error: %v", err)
	}

	// Verify timing - should have delayed at least 100ms
	if elapsed < 100*time.Millisecond {
		t.Errorf("execution was too fast (%v), delay didn't work", elapsed)
	}

	// Verify transport WAS called (delay doesn't skip request)
	if !transportCalled {
		t.Error("transport should be called for delay steps")
	}

	// Verify we got the actual response body
	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}
	if !strings.Contains(result.Results[0].Body, "status") {
		t.Errorf("expected actual response body, got: %s", result.Results[0].Body)
	}
}

func TestRunChain_DelayInvalidDuration(t *testing.T) {
	// Test error handling for invalid delay duration format
	base := &config.ConfigV1{URL: "http://example.com"}
	steps := []config.ChainStep{
		{
			Name: "bad_delay",
			ConfigV1: config.ConfigV1{
				URL:   "http://example.com",
				Delay: "abc", // Invalid format
			},
		},
	}

	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{},
			Body:       io.NopCloser(strings.NewReader(`{}`)),
		}, nil
	}

	factory := &mockExecutorFactory{transport: mockTransport}

	_, err := RunChain(context.Background(), factory, base, steps, Options{})
	if err == nil {
		t.Error("expected error for invalid duration, got nil")
	}
	if !strings.Contains(err.Error(), "invalid delay") {
		t.Errorf("expected 'invalid delay' error, got: %v", err)
	}
}

func TestRunChain_DelayContextCancellation(t *testing.T) {
	// Test that delay respects context cancellation
	base := &config.ConfigV1{URL: "http://example.com"}
	steps := []config.ChainStep{
		{
			Name: "long_delay",
			ConfigV1: config.ConfigV1{
				URL:   "http://example.com",
				Delay: "5s", // Long delay
			},
		},
	}

	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{},
			Body:       io.NopCloser(strings.NewReader(`{}`)),
		}, nil
	}

	factory := &mockExecutorFactory{transport: mockTransport}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := RunChain(ctx, factory, base, steps, Options{})
	elapsed := time.Since(start)

	// Should have cancelled quickly, not waited 5 seconds
	if elapsed > 500*time.Millisecond {
		t.Errorf("context cancellation didn't work, elapsed: %v", elapsed)
	}

	if err == nil {
		t.Error("expected context error, got nil")
	}
	if !strings.Contains(err.Error(), "context") {
		t.Errorf("expected context error, got: %v", err)
	}
}

func TestRunChain_NegativeDelay(t *testing.T) {
	// Test that negative delay doesn't cause issues (should be skipped)
	base := &config.ConfigV1{URL: "http://example.com"}
	steps := []config.ChainStep{
		{
			Name: "negative_delay",
			ConfigV1: config.ConfigV1{
				URL:   "http://example.com",
				Delay: "-5s",
			},
		},
	}

	mockTransport := func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		return &domain.Response{
			StatusCode: 200,
			Headers:    map[string]string{},
			Body:       io.NopCloser(strings.NewReader(`{}`)),
		}, nil
	}

	factory := &mockExecutorFactory{transport: mockTransport}

	start := time.Now()
	_, err := RunChain(context.Background(), factory, base, steps, Options{})
	elapsed := time.Since(start)

	// Should complete quickly since negative duration is skipped
	if elapsed > 500*time.Millisecond {
		t.Errorf("negative duration caused unexpected delay: %v", elapsed)
	}

	// Should not error - negative duration is just skipped
	if err != nil {
		t.Errorf("unexpected error for negative duration: %v", err)
	}
}
