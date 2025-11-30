package executor_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"cli/internal/config"
	"cli/internal/executor"
)

func TestHTTPExecutor_Execute_GETWithQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/test" {
			t.Errorf("Expected path /api/test, got %s", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("foo") != "bar" {
			t.Errorf("Expected query param foo=bar, got %s", query.Get("foo"))
		}
		if query.Get("baz") != "qux" {
			t.Errorf("Expected query param baz=qux, got %s", query.Get("baz"))
		}
		w.Write([]byte("OK"))
	}))
	defer srv.Close()

	cfg := &config.YapiConfig{
		URL:    srv.URL + "/api/test",
		Method: "GET",
		Query: map[string]string{
			"foo": "bar",
			"baz": "qux",
		},
	}

	exec := executor.NewHTTPExecutor()
	resp, err := exec.Execute(cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if resp != "OK" {
		t.Errorf("Expected response OK, got %s", resp)
	}
}

func TestHTTPExecutor_URLBuilding(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *config.YapiConfig
		expectedPath string
		expectedQuery string
	}{
		{
			name: "basic URL with path",
			cfg: &config.YapiConfig{
				URL:    "https://example.com",
				Path:   "/api/test",
				Method: "GET",
			},
			expectedPath: "/api/test",
			expectedQuery: "",
		},
		{
			name: "URL without path",
			cfg: &config.YapiConfig{
				URL:    "https://example.com",
				Method: "GET",
			},
			expectedPath: "/", // Root path if no path specified
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
			expectedPath: "/api",
			expectedQuery: "baz=qux&foo=bar", // Query params are sorted alphabetically for consistent testing
		},
		{
			name: "URL encodes special characters in path",
			cfg: &config.YapiConfig{
				URL:    "https://example.com",
				Path:   "/api/test with spaces",
				Method: "GET",
			},
			expectedPath: "/api/test with spaces",
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
			expectedPath: "/api",
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
			_, err := exec.Execute(tt.cfg)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
		})
	}
}

func TestHTTPExecutor_Execute_POSTWithJSONBody(t *testing.T) {
	expectedBody := `{"name":"test","value":123}`
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
		if string(bodyBytes) != expectedBody {
			t.Errorf("Expected request body %s, got %s", expectedBody, string(bodyBytes))
		}
		w.Write([]byte(`{"status":"received"}`))
	}))
	defer srv.Close()

	cfg := &config.YapiConfig{
		URL:    srv.URL,
		Method: "POST",
		Body: map[string]interface{}{
			"name":  "test",
			"value": 123,
		},
	}

	exec := executor.NewHTTPExecutor()
	resp, err := exec.Execute(cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	expectedResponse := `{"status":"received"}`
	if resp != expectedResponse {
		t.Errorf("Expected response %s, got %s", expectedResponse, resp)
	}
}
