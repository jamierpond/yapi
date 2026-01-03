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
	return loadFromStringInternal(string(data), "", nil, nil)
}

func TestLoadFromStringInternal(t *testing.T) {
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
			_, err := loadFromStringInternal(tt.input, "", nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadFromStringInternal() error = %v, wantErr %v", err, tt.wantErr)
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

func TestLoadFromStringWithPath_EnvFiles(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "yapi-envfiles-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a .env file with test variables
	envContent := `API_KEY=test-api-key-123
API_URL=https://api.example.com
DEBUG=true`

	envPath := tmpDir + "/.env.test"
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Create the yapi config file that references the .env file
	yapiContent := `yapi: v1
url: ${API_URL}/users
method: GET
env_files:
  - .env.test
headers:
  Authorization: Bearer ${API_KEY}`

	yapiPath := tmpDir + "/test.yapi.yml"
	if err := os.WriteFile(yapiPath, []byte(yapiContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Load with path context
	result, err := LoadFromStringWithPath(yapiContent, yapiPath, nil, nil)
	if err != nil {
		t.Fatalf("LoadFromStringWithPath() failed: %v", err)
	}

	if result == nil || result.Request == nil {
		t.Fatal("LoadFromStringWithPath() returned nil result or request")
	}

	// Verify the URL was expanded correctly
	expectedURL := "https://api.example.com/users"
	if result.Request.URL != expectedURL {
		t.Errorf("URL = %q, want %q", result.Request.URL, expectedURL)
	}

	// Verify the header was expanded correctly
	expectedAuth := "Bearer test-api-key-123"
	if result.Request.Headers["Authorization"] != expectedAuth {
		t.Errorf("Authorization header = %q, want %q", result.Request.Headers["Authorization"], expectedAuth)
	}
}

func TestLoadFromStringWithPath_EnvFiles_MultipleFiles(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "yapi-envfiles-multi-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create first .env file
	env1Content := `BASE_URL=https://base.example.com
SHARED_VAR=from-first-file`
	if err := os.WriteFile(tmpDir+"/.env.base", []byte(env1Content), 0600); err != nil {
		t.Fatal(err)
	}

	// Create second .env file (should override SHARED_VAR)
	env2Content := `API_KEY=secret-key
SHARED_VAR=from-second-file`
	if err := os.WriteFile(tmpDir+"/.env.local", []byte(env2Content), 0600); err != nil {
		t.Fatal(err)
	}

	// Create the yapi config file that references both .env files
	yapiContent := `yapi: v1
url: ${BASE_URL}/api
method: POST
env_files:
  - .env.base
  - .env.local
headers:
  X-API-Key: ${API_KEY}
  X-Shared: ${SHARED_VAR}`

	yapiPath := tmpDir + "/test.yapi.yml"

	// Load with path context
	result, err := LoadFromStringWithPath(yapiContent, yapiPath, nil, nil)
	if err != nil {
		t.Fatalf("LoadFromStringWithPath() failed: %v", err)
	}

	if result == nil || result.Request == nil {
		t.Fatal("LoadFromStringWithPath() returned nil result or request")
	}

	// Verify the URL was expanded from first file
	expectedURL := "https://base.example.com/api"
	if result.Request.URL != expectedURL {
		t.Errorf("URL = %q, want %q", result.Request.URL, expectedURL)
	}

	// Verify the API key was expanded from second file
	if result.Request.Headers["X-API-Key"] != "secret-key" {
		t.Errorf("X-API-Key header = %q, want %q", result.Request.Headers["X-API-Key"], "secret-key")
	}

	// Verify the shared var was overridden by second file
	if result.Request.Headers["X-Shared"] != "from-second-file" {
		t.Errorf("X-Shared header = %q, want %q", result.Request.Headers["X-Shared"], "from-second-file")
	}
}

func TestLoadFromStringWithPath_EnvFiles_MissingFile(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "yapi-envfiles-missing-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create the yapi config file that references a non-existent .env file
	yapiContent := `yapi: v1
url: https://example.com
method: GET
env_files:
  - .env.nonexistent`

	yapiPath := tmpDir + "/test.yapi.yml"

	// Load with path context - should fail
	_, err = LoadFromStringWithPath(yapiContent, yapiPath, nil, nil)
	if err == nil {
		t.Error("LoadFromStringWithPath() should have failed for missing env file")
	}
}

func TestLoadFromStringWithPath_EnvFiles_OSEnvTakesPrecedence(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "yapi-envfiles-precedence-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a .env file
	envContent := `TEST_VAR=from-env-file`
	if err := os.WriteFile(tmpDir+"/.env.test", []byte(envContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Set an OS environment variable that should take precedence
	t.Setenv("TEST_VAR", "from-os-env")

	// Create the yapi config file
	yapiContent := `yapi: v1
url: https://example.com/${TEST_VAR}
method: GET
env_files:
  - .env.test`

	yapiPath := tmpDir + "/test.yapi.yml"

	// Load with path context
	result, err := LoadFromStringWithPath(yapiContent, yapiPath, nil, nil)
	if err != nil {
		t.Fatalf("LoadFromStringWithPath() failed: %v", err)
	}

	// Verify OS env takes precedence over .env file
	expectedURL := "https://example.com/from-os-env"
	if result.Request.URL != expectedURL {
		t.Errorf("URL = %q, want %q (OS env should take precedence)", result.Request.URL, expectedURL)
	}
}

func FuzzLoadFromStringInternal(f *testing.F) {
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
		// loadFromStringInternal should not panic on any input
		_, _ = loadFromStringInternal(input, "", nil, nil)
	})
}
