package runner

import (
	"testing"

	"yapi.run/cli/internal/config"
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
	ctx := NewChainContext()
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
