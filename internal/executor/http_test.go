package executor_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"yapi/internal/config"
	"yapi/internal/executor"
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
