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
	"yapi.run/cli/internal/filter"
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
			expectation: config.Expectation{Status: []any{float64(200), float64(201)}},
			result:      &Result{StatusCode: 201},
			wantErr:     false,
		},
		{
			name:        "status not in array",
			expectation: config.Expectation{Status: []any{float64(200), float64(201)}},
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
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{`.status == "success"`}}},
			result:      &Result{Body: `{"status": "success"}`},
			wantErr:     false,
		},
		{
			name:        "assertion fails - value mismatch",
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{`.status == "error"`}}},
			result:      &Result{Body: `{"status": "success"}`},
			wantErr:     true,
		},
		{
			name:        "assertion passes - field exists",
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{`.status != null`}}},
			result:      &Result{Body: `{"status": "success"}`},
			wantErr:     false,
		},
		{
			name:        "assertion fails - field missing",
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{`.missing != null`}}},
			result:      &Result{Body: `{"status": "success"}`},
			wantErr:     true,
		},
		{
			name:        "multiple assertions - all pass",
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{`.status == "success"`, `.data == "test"`}}},
			result:      &Result{Body: `{"status": "success", "data": "test"}`},
			wantErr:     false,
		},
		{
			name:        "multiple assertions - one fails",
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{`.status == "success"`, `.data == "wrong"`}}},
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
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{`.items | length > 0`}}},
			result:      &Result{Body: `{"items": [1, 2, 3]}`},
			wantErr:     false,
		},
		{
			name:        "empty array fails length check",
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{`.items | length > 0`}}},
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
	ctx := NewChainContext(nil)
	ctx.Results["step1"] = StepResult{
		BodyJSON: map[string]any{
			"result": map[string]any{
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
		wantVal any
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
	ctx := NewChainContext(nil)
	ctx.Results["prev"] = StepResult{
		BodyJSON:   map[string]any{"token": "abc123"},
		StatusCode: 200,
	}
	// Add step with typed values for type preservation tests
	ctx.Results["step1"] = StepResult{
		BodyJSON: map[string]any{
			"result": map[string]any{
				"index": float64(7),
			},
		},
		StatusCode: 200,
	}

	tests := []struct {
		name     string
		body     map[string]any
		expected map[string]any
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
			body: map[string]any{
				"auth": "${prev.token}",
			},
			expected: map[string]any{
				"auth": "abc123",
			},
			wantErr: false,
		},
		{
			name: "non-string values unchanged",
			body: map[string]any{
				"count": 42,
				"flag":  true,
			},
			expected: map[string]any{
				"count": 42,
				"flag":  true,
			},
			wantErr: false,
		},
		{
			name: "nested body",
			body: map[string]any{
				"data": map[string]any{
					"token": "${prev.token}",
				},
			},
			expected: map[string]any{
				"data": map[string]any{
					"token": "abc123",
				},
			},
			wantErr: false,
		},
		{
			name: "type preservation - int",
			body: map[string]any{
				"track_index": "$step1.result.index",
			},
			expected: map[string]any{
				"track_index": float64(7), // Preserved as number, not string
			},
			wantErr: false,
		},
		{
			name: "mixed string stays string",
			body: map[string]any{
				"message": "Track $step1.result.index created",
			},
			expected: map[string]any{
				"message": "Track 7 created", // Interpolated as string
			},
			wantErr: false,
		},
		{
			name: "array with variable references",
			body: map[string]any{
				"track_indices": []any{
					"$step1.result.index",
					"$prev.token",
				},
			},
			expected: map[string]any{
				"track_indices": []any{
					float64(7),
					"abc123",
				},
			},
			wantErr: false,
		},
		{
			name: "nested array in object",
			body: map[string]any{
				"params": map[string]any{
					"indices": []any{
						"$step1.result.index",
					},
				},
			},
			expected: map[string]any{
				"params": map[string]any{
					"indices": []any{
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
					if expectedNested, ok := expectedVal.(map[string]any); ok {
						actualNested, ok := actualVal.(map[string]any)
						if !ok {
							t.Errorf("key '%s' expected map, got %T", k, actualVal)
							continue
						}
						for nk, nv := range expectedNested {
							// Handle arrays in nested maps
							if expectedArr, ok := nv.([]any); ok {
								actualArr, ok := actualNested[nk].([]any)
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
					} else if expectedArr, ok := expectedVal.([]any); ok {
						// Handle arrays
						actualArr, ok := actualVal.([]any)
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

func TestFormatAssertionError(t *testing.T) {
	tests := []struct {
		name        string
		detail      *filter.AssertionDetail
		wantContain []string
	}{
		{
			name: "equality operator",
			detail: &filter.AssertionDetail{
				Expression:    ".id == 999",
				LeftSide:      ".id",
				Operator:      "==",
				RightSide:     "999",
				ActualValue:   "1",
				ExpectedValue: "999",
			},
			wantContain: []string{
				"Expected: .id to equal 999",
				"Actual:   .id = 1",
				"Expression: .id == 999",
			},
		},
		{
			name: "not equal operator",
			detail: &filter.AssertionDetail{
				Expression:    ".userId != null",
				LeftSide:      ".userId",
				Operator:      "!=",
				RightSide:     "null",
				ActualValue:   "1",
				ExpectedValue: "null",
			},
			wantContain: []string{
				"Expected: .userId to not equal null",
				"Actual:   .userId = 1",
				"Expression: .userId != null",
			},
		},
		{
			name: "greater than operator",
			detail: &filter.AssertionDetail{
				Expression:    ".id > 100",
				LeftSide:      ".id",
				Operator:      ">",
				RightSide:     "100",
				ActualValue:   "1",
				ExpectedValue: "100",
			},
			wantContain: []string{
				"Expected: .id to be greater than 100",
				"Actual:   .id = 1",
				"Expression: .id > 100",
			},
		},
		{
			name: "greater than or equal operator",
			detail: &filter.AssertionDetail{
				Expression:    ".score >= 10",
				LeftSide:      ".score",
				Operator:      ">=",
				RightSide:     "10",
				ActualValue:   "5",
				ExpectedValue: "10",
			},
			wantContain: []string{
				"Expected: .score to be greater than or equal to 10",
				"Actual:   .score = 5",
				"Expression: .score >= 10",
			},
		},
		{
			name: "less than operator",
			detail: &filter.AssertionDetail{
				Expression:    ".value < 10",
				LeftSide:      ".value",
				Operator:      "<",
				RightSide:     "10",
				ActualValue:   "15",
				ExpectedValue: "10",
			},
			wantContain: []string{
				"Expected: .value to be less than 10",
				"Actual:   .value = 15",
				"Expression: .value < 10",
			},
		},
		{
			name: "less than or equal operator",
			detail: &filter.AssertionDetail{
				Expression:    ".count <= 5",
				LeftSide:      ".count",
				Operator:      "<=",
				RightSide:     "5",
				ActualValue:   "10",
				ExpectedValue: "5",
			},
			wantContain: []string{
				"Expected: .count to be less than or equal to 5",
				"Actual:   .count = 10",
				"Expression: .count <= 5",
			},
		},
		{
			name: "complex expression with pipe",
			detail: &filter.AssertionDetail{
				Expression:    ".title | length > 100",
				LeftSide:      ".title | length",
				Operator:      ">",
				RightSide:     "100",
				ActualValue:   "18",
				ExpectedValue: "100",
			},
			wantContain: []string{
				"Expected: .title | length to be greater than 100",
				"Actual:   .title | length = 18",
				"Expression: .title | length > 100",
			},
		},
		{
			name:   "nil detail",
			detail: nil,
			wantContain: []string{
				"assertion failed",
			},
		},
		{
			name: "detail without operator info",
			detail: &filter.AssertionDetail{
				Expression: ".some.complex.expression",
			},
			wantContain: []string{
				"assertion failed: .some.complex.expression",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAssertionError(tt.detail)
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("formatAssertionError() = %q\nwant to contain: %q", got, want)
				}
			}
		})
	}
}

func TestCheckExpectations_DetailedErrors(t *testing.T) {
	tests := []struct {
		name            string
		expectation     config.Expectation
		result          *Result
		wantErr         bool
		wantErrContains []string
	}{
		{
			name:        "equality assertion failure with details",
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{".id == 999"}}},
			result:      &Result{Body: `{"id": 1, "title": "test"}`},
			wantErr:     true,
			wantErrContains: []string{
				"Expected: .id to equal 999",
				"Actual:   .id = 1",
				"Expression: .id == 999",
			},
		},
		{
			name:        "greater than assertion failure",
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{".id > 100"}}},
			result:      &Result{Body: `{"id": 1}`},
			wantErr:     true,
			wantErrContains: []string{
				"Expected: .id to be greater than 100",
				"Actual:   .id = 1",
			},
		},
		{
			name:        "not equal assertion failure",
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{".completed != false"}}},
			result:      &Result{Body: `{"completed": false}`},
			wantErr:     true,
			wantErrContains: []string{
				"Expected: .completed to not equal false",
				"Actual:   .completed = false",
			},
		},
		{
			name:        "complex pipe expression failure",
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{".title | length > 100"}}},
			result:      &Result{Body: `{"title": "short"}`},
			wantErr:     true,
			wantErrContains: []string{
				"Expected: .title | length to be greater than 100",
				"Actual:   .title | length = 5",
			},
		},
		{
			name:        "null comparison failure",
			expectation: config.Expectation{Assert: config.AssertionSet{Body: []string{".userId == null"}}},
			result:      &Result{Body: `{"userId": 1}`},
			wantErr:     true,
			wantErrContains: []string{
				"Expected: .userId to equal null",
				"Actual:   .userId = 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := CheckExpectations(tt.expectation, tt.result)
			if (res.Error != nil) != tt.wantErr {
				t.Errorf("CheckExpectations() error = %v, wantErr %v", res.Error, tt.wantErr)
				return
			}

			if tt.wantErr && res.Error != nil {
				errMsg := res.Error.Error()
				for _, want := range tt.wantErrContains {
					if !strings.Contains(errMsg, want) {
						t.Errorf("CheckExpectations() error message:\n%s\nwant to contain: %q", errMsg, want)
					}
				}
			}
		})
	}
}

func TestCheckExpectations_HeaderAssertions(t *testing.T) {
	tests := []struct {
		name        string
		expectation config.Expectation
		result      *Result
		wantErr     bool
	}{
		{
			name: "header assertion passes - header exists",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Headers: []string{`.["Content-Type"] != null`},
				},
			},
			result: &Result{
				Body: `{}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantErr: false,
		},
		{
			name: "header assertion passes - header value matches",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Headers: []string{`.["Content-Type"] == "application/json"`},
				},
			},
			result: &Result{
				Body: `{}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantErr: false,
		},
		{
			name: "header assertion fails - header missing",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Headers: []string{`.["X-Custom-Header"] != null`},
				},
			},
			result: &Result{
				Body: `{}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantErr: true,
		},
		{
			name: "header assertion fails - value mismatch",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Headers: []string{`.["Content-Type"] == "text/html"`},
				},
			},
			result: &Result{
				Body: `{}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantErr: true,
		},
		{
			name: "multiple header assertions - all pass",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Headers: []string{
						`.["Content-Type"] != null`,
						`.["X-Custom"] == "value123"`,
					},
				},
			},
			result: &Result{
				Body: `{}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
					"X-Custom":     "value123",
				},
			},
			wantErr: false,
		},
		{
			name: "multiple header assertions - one fails",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Headers: []string{
						`.["Content-Type"] != null`,
						`.["X-Custom"] == "wrong"`,
					},
				},
			},
			result: &Result{
				Body: `{}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
					"X-Custom":     "value123",
				},
			},
			wantErr: true,
		},
		{
			name: "both body and header assertions - all pass",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Body:    []string{`.id == 1`},
					Headers: []string{`.["Content-Type"] == "application/json"`},
				},
			},
			result: &Result{
				Body: `{"id": 1}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantErr: false,
		},
		{
			name: "both body and header assertions - body fails",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Body:    []string{`.id == 999`},
					Headers: []string{`.["Content-Type"] == "application/json"`},
				},
			},
			result: &Result{
				Body: `{"id": 1}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantErr: true,
		},
		{
			name: "both body and header assertions - header fails",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Body:    []string{`.id == 1`},
					Headers: []string{`.["Content-Type"] == "text/html"`},
				},
			},
			result: &Result{
				Body: `{"id": 1}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantErr: true,
		},
		{
			name: "header assertion with complex expression",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Headers: []string{`(.["Content-Type"] // "") | contains("json")`},
				},
			},
			result: &Result{
				Body: `{}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantErr: false,
		},
		{
			name: "empty headers with assertion - fails",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Headers: []string{`.["Content-Type"] != null`},
				},
			},
			result: &Result{
				Body:    `{}`,
				Headers: map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "case-sensitive header names",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Headers: []string{`.["content-type"] != null`},
				},
			},
			result: &Result{
				Body: `{}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := CheckExpectations(tt.expectation, tt.result)
			if (res.Error != nil) != tt.wantErr {
				t.Errorf("CheckExpectations() error = %v, wantErr %v", res.Error, tt.wantErr)
			}

			expectedTotal := len(tt.expectation.Assert.Body) + len(tt.expectation.Assert.Headers)
			if res.AssertionsTotal != expectedTotal {
				t.Errorf("AssertionsTotal = %d, want %d", res.AssertionsTotal, expectedTotal)
			}

			if !tt.wantErr && res.AssertionsPassed != expectedTotal {
				t.Errorf("AssertionsPassed = %d, want %d", res.AssertionsPassed, expectedTotal)
			}
		})
	}
}

func TestCheckExpectations_HeaderAssertionErrors(t *testing.T) {
	tests := []struct {
		name        string
		expectation config.Expectation
		result      *Result
		wantErr     bool
	}{
		{
			name: "header assertion with invalid JQ expression",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Headers: []string{`.["Content-Type"] &`},
				},
			},
			result: &Result{
				Body: `{}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantErr: true,
		},
		{
			name: "header assertion with non-boolean result",
			expectation: config.Expectation{
				Assert: config.AssertionSet{
					Headers: []string{`.["Content-Type"]`},
				},
			},
			result: &Result{
				Body: `{}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
			wantErr: true,
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

func TestCheckExpectations_HeaderAssertionResults(t *testing.T) {
	expectation := config.Expectation{
		Assert: config.AssertionSet{
			Headers: []string{
				`.["Content-Type"] == "application/json"`,
				`.["X-Custom"] != null`,
			},
			Body: []string{
				`.id == 1`,
			},
		},
	}

	result := &Result{
		Body: `{"id": 1}`,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Custom":     "value",
		},
	}

	res := CheckExpectations(expectation, result)

	if res.Error != nil {
		t.Errorf("Expected no error, got: %v", res.Error)
	}

	if res.AssertionsTotal != 3 {
		t.Errorf("AssertionsTotal = %d, want 3", res.AssertionsTotal)
	}

	if res.AssertionsPassed != 3 {
		t.Errorf("AssertionsPassed = %d, want 3", res.AssertionsPassed)
	}

	if len(res.AssertionResults) != 3 {
		t.Fatalf("len(AssertionResults) = %d, want 3", len(res.AssertionResults))
	}

	for i, ar := range res.AssertionResults {
		if !ar.Passed {
			t.Errorf("AssertionResults[%d].Passed = false, want true", i)
		}
		if ar.Error != nil {
			t.Errorf("AssertionResults[%d].Error = %v, want nil", i, ar.Error)
		}
	}
}

func TestCheckExpectations_OnlyHeadersNoBody(t *testing.T) {
	expectation := config.Expectation{
		Assert: config.AssertionSet{
			Headers: []string{
				`.["Content-Type"] != null`,
				`.["Content-Length"] != null`,
			},
		},
	}

	result := &Result{
		Body: `{}`,
		Headers: map[string]string{
			"Content-Type":   "application/json",
			"Content-Length": "42",
		},
	}

	res := CheckExpectations(expectation, result)

	if res.Error != nil {
		t.Errorf("Expected no error, got: %v", res.Error)
	}

	if res.AssertionsTotal != 2 {
		t.Errorf("AssertionsTotal = %d, want 2 (only headers)", res.AssertionsTotal)
	}

	if res.AssertionsPassed != 2 {
		t.Errorf("AssertionsPassed = %d, want 2", res.AssertionsPassed)
	}
}
