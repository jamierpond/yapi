package executor_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"yapi.run/internal/config"
	"yapi.run/internal/executor"
)

func TestHTTPExecutor_URLBuilding(t *testing.T) {
	tests := []struct {
		name          string
		cfg           *config.YapiConfig
		expectedPath  string
		expectedQuery string
	}{
		{
			name: "basic URL with path",
			cfg: &config.YapiConfig{
				URL:    "https://example.com",
				Path:   "/api/test",
				Method: "GET",
			},
			expectedPath:  "/api/test",
			expectedQuery: "",
		},
		{
			name: "URL without path",
			cfg: &config.YapiConfig{
				URL:    "https://example.com",
				Method: "GET",
			},
			expectedPath:  "/", // Root path if no path specified
			expectedQuery: "",
		},
		{
			name: "URL with query string",
			cfg: &config.YapiConfig{
				URL:    "https://example.com",
				Path:   "/api",
				Method: "GET",
				Query: map[string]string{
					"foo": "bar",
					"baz": "qux",
				},
			},
			expectedPath:  "/api",
			expectedQuery: "baz=qux&foo=bar", // Query params are sorted alphabetically for consistent testing
		},
		{
			name: "URL encodes special characters in path",
			cfg: &config.YapiConfig{
				URL:    "https://example.com",
				Path:   "/api/test with spaces",
				Method: "GET",
			},
			expectedPath:  "/api/test with spaces",
			expectedQuery: "",
		},
		{
			name: "URL encodes special characters in query",
			cfg: &config.YapiConfig{
				URL:    "https://example.com",
				Path:   "/api",
				Method: "GET",
				Query: map[string]string{
					"q": "hello world!",
				},
			},
			expectedPath:  "/api",
			expectedQuery: "q=hello+world%21", // url.Values.Encode() uses %21 for '!' and '+' for ' '
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.expectedPath {
					t.Errorf("Expected path %q, got %q", tt.expectedPath, r.URL.Path)
				}

				// For query, construct expected query string using url.Values.Encode() for consistent comparison
				expectedQueryValues := make(url.Values)
				for k, v := range tt.cfg.Query {
					expectedQueryValues.Add(k, v)
				}
				actualQuery := r.URL.Query().Encode()
				if actualQuery != expectedQueryValues.Encode() {
					t.Errorf("Expected query %q, got %q", expectedQueryValues.Encode(), actualQuery)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			tt.cfg.URL = srv.URL // Update config URL to point to mock server

			exec := executor.NewHTTPExecutor()
			resp, err := exec.Execute(tt.cfg)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
			if resp == nil {
				t.Fatal("Execute returned nil response")
			}
		})
	}
}

func TestHTTPExecutor_Execute_BodyAndJSON(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *config.YapiConfig
		expectedBody   string
		expectedStatus int
	}{
		{
			name: "POST with simple JSON body",
			cfg: &config.YapiConfig{
				URL:    "", // Will be set to mock server URL
				Method: "POST",
				Body: map[string]interface{}{
					"name":  "test",
					"value": 123,
				},
			},
			expectedBody:   `{"name":"test","value":123}`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "POST with complex nested JSON body",
			cfg: &config.YapiConfig{
				URL:    "", // Will be set to mock server URL
				Method: "POST",
				Body: map[string]interface{}{
					"title":       "Testing yapi - YAML API Testing Tool",
					"description": "This demo shows nested objects, arrays, and various data types",
					"userId":      123,
					"isPublished": true,
					"tags":        []interface{}{"testing", "api", "yaml"},
					"metadata": map[string]interface{}{
						"source":    "yapi",
						"version":   "1.0",
						"timestamp": "2024-01-15T10:30:00Z",
					},
					"author": map[string]interface{}{
						"name":  "Test User",
						"email": "test@example.com",
					},
				},
			},
			expectedBody:   `{"author":{"email":"test@example.com","name":"Test User"},"description":"This demo shows nested objects, arrays, and various data types","isPublished":true,"metadata":{"source":"yapi","timestamp":"2024-01-15T10:30:00Z","version":"1.0"},"tags":["testing","api","yaml"],"title":"Testing yapi - YAML API Testing Tool","userId":123}`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "POST with raw JSON string",
			cfg: &config.YapiConfig{
				URL:    "", // Will be set to mock server URL
				Method: "POST",
				JSON:   `{"status":"active","code":42}`, // Raw JSON directly
			},
			expectedBody:   `{"status":"active","code":42}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST method, got %s", r.Method)
				}
				// Content-Type should be application/json by default if body/json is present
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("Failed to read request body: %v", err)
				}

				// Compare JSON bodies robustly by unmarshalling and marshalling again
				var actual, expected interface{}
				err = json.Unmarshal(bodyBytes, &actual)
				if err != nil {
					t.Fatalf("Failed to unmarshal actual request body: %v, body: %s", err, string(bodyBytes))
				}
				err = json.Unmarshal([]byte(tt.expectedBody), &expected)
				if err != nil {
					t.Fatalf("Failed to unmarshal expected request body: %v, body: %s", err, tt.expectedBody)
				}

				if !reflect.DeepEqual(actual, expected) {
					t.Errorf("Expected request body %v, got %v", expected, actual)
				}

				w.WriteHeader(tt.expectedStatus)
				w.Write([]byte(`{"status":"received"}`)) // Generic response
			}))
			defer srv.Close()

			tt.cfg.URL = srv.URL // Update config URL to point to mock server

			exec := executor.NewHTTPExecutor()
			resp, err := exec.Execute(tt.cfg)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			// Verify generic response
			expectedResponse := `{"status":"received"}`
			if resp.Body != expectedResponse {
				t.Errorf("Expected response %s, got %s", expectedResponse, resp.Body)
			}
		})
	}
}
