package executor_test

import (
	"context"
	"encoding/json"
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
url: https://example.com/
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
			name: "URL with special characters in query requiring encoding",
			yaml: `
yapi: v1
url: https://example.com
path: /search
method: GET
query:
  q: "fish in:name"
  sort: stars`,
			expectedPath: "/search",
			expectedQuery: url.Values{
				"q":    {"fish in:name"},
				"sort": {"stars"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := config.LoadFromString(tt.yaml)
			if err != nil {
				t.Fatalf("LoadFromString failed: %v", err)
			}
			req := res.Request

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

			// Parse the original URL to extract path and query
			parsedURL, err := url.Parse(req.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}
			// Replace with test server URL + path + query
			req.URL = srv.URL + parsedURL.Path
			if parsedURL.RawQuery != "" {
				req.URL += "?" + parsedURL.RawQuery
			}

			client := &http.Client{}
			execFn := executor.HTTPTransport(client)
			resp, err := execFn(context.Background(), req)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
			if resp == nil {
				t.Fatal("Execute returned nil response")
			}
		})
	}
}

func TestHTTPTransport_BodyAndJSON(t *testing.T) {
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
			req := res.Request

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

				var actual, expected any
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

			req.URL = srv.URL

			client := &http.Client{}
			execFn := executor.HTTPTransport(client)
			resp, err := execFn(context.Background(), req)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			expectedResponse := `{"status":"received"}`
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}
			if string(bodyBytes) != expectedResponse {
				t.Errorf("Expected response %s, got %s", expectedResponse, string(bodyBytes))
			}
		})
	}
}
