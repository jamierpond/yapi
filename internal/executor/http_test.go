package executor_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/executor"
)

func TestHTTPExecutor_URLBuilding(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		expectedPath  string
		expectedQuery url.Values
	}{
		{
			name: "basic URL with path",
			yaml: `
yapi: v1
url: https://example.com
path: /api/test
method: GET`,
			expectedPath:  "/api/test",
			expectedQuery: url.Values{},
		},
		{
			name: "URL without path",
			yaml: `
yapi: v1
url: https://example.com
method: GET`,
			expectedPath:  "/",
			expectedQuery: url.Values{},
		},
		{
			name: "URL with query string",
			yaml: `
yapi: v1
url: https://example.com
path: /api
method: GET
query:
  foo: bar
  baz: qux`,
			expectedPath: "/api",
			expectedQuery: url.Values{
				"foo": {"bar"},
				"baz": {"qux"},
			},
		},
		{
			name: "URL encodes special characters in path",
			yaml: `
yapi: v1
url: https://example.com
path: "/api/test with spaces"
method: GET`,
			expectedPath:  "/api/test with spaces",
			expectedQuery: url.Values{},
		},
		{
			name: "URL encodes special characters in query",
			yaml: `
yapi: v1
url: https://example.com
path: /api
method: GET
query:
  q: "hello world!"`,
			expectedPath: "/api",
			expectedQuery: url.Values{
				"q": {"hello world!"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := config.LoadFromString(tt.yaml)
			if err != nil {
				t.Fatalf("LoadFromString failed: %v", err)
			}
			cfg := res.Config

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.expectedPath {
					t.Errorf("Expected path %q, got %q", tt.expectedPath, r.URL.Path)
				}
				if r.URL.Query().Encode() != tt.expectedQuery.Encode() {
					t.Errorf("Expected query %q, got %q", tt.expectedQuery.Encode(), r.URL.Query().Encode())
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			cfg.URL = srv.URL

			exec := executor.NewHTTPExecutor()
			resp, err := exec.Execute(cfg)
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
		yaml           string
		expectedBody   string
		expectedStatus int
	}{
		{
			name: "POST with simple JSON body",
			yaml: `
yapi: v1
url: ""
method: POST
body:
  name: test
  value: 123`,
			expectedBody:   `{"name":"test","value":123}`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "POST with complex nested JSON body",
			yaml: `
yapi: v1
url: ""
method: POST
body:
  title: "Testing yapi - YAML API Testing Tool"
  description: "This demo shows nested objects, arrays, and various data types"
  userId: 123
  isPublished: true
  tags:
    - testing
    - api
    - yaml
  metadata:
    source: yapi
    version: "1.0"
    timestamp: "2024-01-15T10:30:00Z"
  author:
    name: "Test User"
    email: "test@example.com"`,
			expectedBody:   `{"author":{"email":"test@example.com","name":"Test User"},"description":"This demo shows nested objects, arrays, and various data types","isPublished":true,"metadata":{"source":"yapi","timestamp":"2024-01-15T10:30:00Z","version":"1.0"},"tags":["testing","api","yaml"],"title":"Testing yapi - YAML API Testing Tool","userId":123}`,
			expectedStatus: http.StatusOK,
		},
		{
			name: "POST with raw JSON string",
			yaml: `
yapi: v1
url: ""
method: POST
json: '{"status":"active","code":42}'`,
			expectedBody:   `{"status":"active","code":42}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := config.LoadFromString(tt.yaml)
			if err != nil {
				t.Fatalf("LoadFromString failed: %v", err)
			}
			cfg := res.Config

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST method, got %s", r.Method)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("Failed to read request body: %v", err)
				}

				var actual, expected interface{}
				if err := json.Unmarshal(bodyBytes, &actual); err != nil {
					t.Fatalf("Failed to unmarshal actual request body: %v, body: %s", err, string(bodyBytes))
				}
				if err := json.Unmarshal([]byte(tt.expectedBody), &expected); err != nil {
					t.Fatalf("Failed to unmarshal expected request body: %v, body: %s", err, tt.expectedBody)
				}

				if !reflect.DeepEqual(actual, expected) {
					t.Errorf("Expected request body %v, got %v", expected, actual)
				}

				w.WriteHeader(tt.expectedStatus)
				w.Write([]byte(`{"status":"received"}`))
			}))
			defer srv.Close()

			cfg.URL = srv.URL

			exec := executor.NewHTTPExecutor()
			resp, err := exec.Execute(cfg)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			expectedResponse := `{"status":"received"}`
			if resp.Body != expectedResponse {
				t.Errorf("Expected response %s, got %s", expectedResponse, resp.Body)
			}
		})
	}
}
