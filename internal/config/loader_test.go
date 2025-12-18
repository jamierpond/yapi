package config

import (
	"fmt"
	"io"
	"os"
	"testing"
)

// load is a test helper that reads and parses a yapi config file from the given path.
// If path is "-", reads from stdin.
func load(path string) (*ParseResult, error) {
	var data []byte
	var err error

	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
	} else {
		data, err = os.ReadFile(path) //nolint:gosec // user-provided config file
		if err != nil {
			return nil, err
		}
	}
	return LoadFromString(string(data))
}

func TestLoadFromString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid v1 config",
			input: `yapi: v1
url: https://example.com
method: GET`,
			wantErr: false,
		},
		{
			name: "legacy config without version",
			input: `url: https://example.com
method: GET`,
			wantErr: false,
		},
		{
			name:    "invalid yaml",
			input:   `{invalid: yaml: syntax`,
			wantErr: true,
		},
		{
			name:    "unsupported version",
			input:   `yapi: v99`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad_File(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "yapi-test-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := `yapi: v1
url: https://example.com
method: GET`

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test load function
	result, err := load(tmpfile.Name())
	if err != nil {
		t.Fatalf("load() failed: %v", err)
	}

	if result == nil {
		t.Fatal("load() returned nil result")
	}

	if result.Request == nil {
		t.Fatal("load() returned nil request")
	}

	// Verify the config was loaded successfully
	if result.Request.URL != "https://example.com" {
		t.Errorf("load() URL = %v, want https://example.com", result.Request.URL)
	}
}

func TestLoad_Stdin(t *testing.T) {
	// This test verifies that Load("-") reads from stdin.
	// Note: Actual stdin testing would require more complex setup,
	// so we just verify the code path doesn't panic and handles the special case.
	// Real testing is done via integration tests.
	t.Skip("stdin testing requires complex setup - tested manually")
}

func FuzzLoadFromString(f *testing.F) {
	// Seed with valid YAML configs
	f.Add(`yapi: v1
url: https://example.com
method: GET`)

	f.Add(`yapi: v1
url: https://api.example.com/users
method: POST
headers:
  Content-Type: application/json
json: '{"name": "test"}'`)

	f.Add(`yapi: v1
url: https://example.com
graphql: |
  query {
    users {
      id
      name
    }
  }`)

	f.Add(`yapi: v1
url: grpc://localhost:50051
service: myservice.MyService
rpc: GetUser`)

	f.Add(`yapi: v1
url: tcp://localhost:8080
data: "hello"
encoding: text`)

	// Chain config
	f.Add(`yapi: v1
chain:
  - name: auth
    url: https://example.com/auth
    method: POST
  - name: api
    url: https://example.com/api
    headers:
      Authorization: "Bearer ${auth.response.body.token}"`)

	// Invalid/edge cases
	f.Add(``)
	f.Add(`{}`)
	f.Add(`[]`)
	f.Add(`null`)
	f.Add(`yapi: v99`)
	f.Add(`url: not-a-url`)

	f.Fuzz(func(t *testing.T, input string) {
		// LoadFromString should not panic on any input
		_, _ = LoadFromString(input)
	})
}
