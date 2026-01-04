package config

import (
	"fmt"
	"io"
	"os"
	"strings"
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
	return loadFromStringInternal(string(data), "", nil, nil, ResolverOptions{})
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
			_, err := loadFromStringInternal(tt.input, "", nil, nil, ResolverOptions{})
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

	// Load with path context - should succeed with warning (not error)
	result, err := LoadFromStringWithPath(yapiContent, yapiPath, nil, nil)
	if err != nil {
		t.Fatalf("LoadFromStringWithPath() should not fail for missing env file, got: %v", err)
	}

	// Should have a warning about the missing file
	if len(result.Warnings) == 0 {
		t.Error("Expected warning for missing env file, got none")
	}

	// Check warning mentions the missing file
	foundWarning := false
	for _, w := range result.Warnings {
		if strings.Contains(w, ".env.nonexistent") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Errorf("Expected warning to mention '.env.nonexistent', got: %v", result.Warnings)
	}
}

func TestLoadFromStringWithPath_EnvFiles_EnvFileTakesPrecedence(t *testing.T) {
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

	// Set an OS environment variable - env_files should take priority over this
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

	// Verify env_files takes precedence over OS env
	expectedURL := "https://example.com/from-env-file"
	if result.Request.URL != expectedURL {
		t.Errorf("URL = %q, want %q (env_files should take precedence over OS env)", result.Request.URL, expectedURL)
	}
}

func TestLoadFromStringWithPath_EnvFiles_MultipleMissingFiles(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "yapi-envfiles-multi-missing-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a yapi config file that references multiple non-existent .env files
	yapiContent := `yapi: v1
url: https://example.com
method: GET
env_files:
  - .env.first
  - .env.second
  - .env.third`

	yapiPath := tmpDir + "/test.yapi.yml"

	// Load with path context - should succeed with multiple warnings
	result, err := LoadFromStringWithPath(yapiContent, yapiPath, nil, nil)
	if err != nil {
		t.Fatalf("LoadFromStringWithPath() should not fail for missing env files, got: %v", err)
	}

	// Should have 3 warnings (one for each missing file)
	if len(result.Warnings) != 3 {
		t.Errorf("Expected 3 warnings, got %d: %v", len(result.Warnings), result.Warnings)
	}

	// Check each warning mentions its respective file
	expectedFiles := []string{".env.first", ".env.second", ".env.third"}
	for i, expectedFile := range expectedFiles {
		found := false
		for _, w := range result.Warnings {
			if strings.Contains(w, expectedFile) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected warning %d to mention '%s', got: %v", i, expectedFile, result.Warnings)
		}
	}
}

func TestLoadFromStringWithPath_EnvFiles_DuplicateEntries(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "yapi-envfiles-duplicate-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a yapi config file with duplicate env_files entries
	yapiContent := `yapi: v1
url: https://example.com
method: GET
env_files:
  - .env.missing
  - .env.missing
  - .env.missing`

	yapiPath := tmpDir + "/test.yapi.yml"

	// Load with path context - duplicates should be deduplicated
	result, err := LoadFromStringWithPath(yapiContent, yapiPath, nil, nil)
	if err != nil {
		t.Fatalf("LoadFromStringWithPath() should not fail, got: %v", err)
	}

	// Should have only 1 warning (duplicates are deduplicated)
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning (duplicates deduplicated), got %d: %v", len(result.Warnings), result.Warnings)
	}
}

func TestLoadFromStringWithPath_EnvFiles_PermissionError(t *testing.T) {
	// Skip on Windows as permission handling is different
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("Skipping permission test on Windows")
	}

	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "yapi-envfiles-permission-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a .env file with no read permissions
	envPath := tmpDir + "/.env.noaccess"
	if err := os.WriteFile(envPath, []byte("SECRET=value"), 0000); err != nil {
		t.Fatal(err)
	}

	// Create the yapi config file
	yapiContent := `yapi: v1
url: https://example.com
method: GET
env_files:
  - .env.noaccess`

	yapiPath := tmpDir + "/test.yapi.yml"

	// Load with path context - should fail with permission error
	_, err = LoadFromStringWithPath(yapiContent, yapiPath, nil, nil)
	if err == nil {
		t.Error("LoadFromStringWithPath() should have failed for permission error")
	}

	// Verify error message mentions the file
	if !strings.Contains(err.Error(), ".env.noaccess") {
		t.Errorf("Error message should mention the file, got: %v", err)
	}
}

func TestLoadFromStringWithPath_EnvFiles_DirectoryNotFile(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "yapi-envfiles-directory-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a directory instead of a file
	envPath := tmpDir + "/.env.isdir"
	if err := os.Mkdir(envPath, 0750); err != nil { // #nosec G301 -- test directory
		t.Fatal(err)
	}

	// Create the yapi config file
	yapiContent := `yapi: v1
url: https://example.com
method: GET
env_files:
  - .env.isdir`

	yapiPath := tmpDir + "/test.yapi.yml"

	// Load with path context - should fail because it's a directory
	_, err = LoadFromStringWithPath(yapiContent, yapiPath, nil, nil)
	if err == nil {
		t.Error("LoadFromStringWithPath() should have failed when env_files references a directory")
	}

	// Verify error message mentions it's a directory
	if !strings.Contains(err.Error(), "directory") {
		t.Errorf("Error message should mention 'directory', got: %v", err)
	}
}

func TestLoadFromStringWithOptions_StrictEnv_MissingFile(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "yapi-strictenv-test-*")
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

	// Load with strict mode - should fail
	_, err = LoadFromStringWithOptions(yapiContent, LoadOptions{
		ConfigPath: yapiPath,
		StrictEnv:  true,
	})
	if err == nil {
		t.Error("LoadFromStringWithOptions with StrictEnv should have failed for missing env file")
	}

	// Verify error message mentions the file
	if !strings.Contains(err.Error(), ".env.nonexistent") {
		t.Errorf("Error message should mention missing file, got: %v", err)
	}
}

func TestLoadFromStringWithOptions_StrictEnv_NoOSFallback(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "yapi-strictenv-osfallback-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an empty .env file (exists but doesn't define TEST_VAR)
	if err := os.WriteFile(tmpDir+"/.env.test", []byte("OTHER_VAR=other"), 0600); err != nil {
		t.Fatal(err)
	}

	// Set an OS environment variable
	t.Setenv("TEST_VAR", "from-os-env")

	// Create the yapi config file
	yapiContent := `yapi: v1
url: https://example.com/${TEST_VAR}
method: GET
env_files:
  - .env.test`

	yapiPath := tmpDir + "/test.yapi.yml"

	// Load with strict mode - OS env should NOT be used
	result, err := LoadFromStringWithOptions(yapiContent, LoadOptions{
		ConfigPath: yapiPath,
		StrictEnv:  true,
	})
	if err != nil {
		t.Fatalf("LoadFromStringWithOptions() failed: %v", err)
	}

	// In strict mode, the variable should be empty (not resolved from OS)
	expectedURL := "https://example.com/"
	if result.Request.URL != expectedURL {
		t.Errorf("URL = %q, want %q (strict mode should not use OS env fallback)", result.Request.URL, expectedURL)
	}
}

func TestBuildEnvFileResolverWithOptions(t *testing.T) {
	// Set up test OS environment variable
	t.Setenv("OS_VAR", "from-os")

	tests := []struct {
		name             string
		envFileVars      map[string]string
		existingResolver func(string) (string, error)
		opts             ResolverOptions
		key              string
		expectedValue    string
	}{
		{
			name:          "env file var takes priority",
			envFileVars:   map[string]string{"VAR": "from-env-file"},
			opts:          ResolverOptions{},
			key:           "VAR",
			expectedValue: "from-env-file",
		},
		{
			name:        "existing resolver used when no env file var",
			envFileVars: map[string]string{},
			existingResolver: func(key string) (string, error) {
				if key == "VAR" {
					return "from-existing", nil
				}
				return "", nil
			},
			opts:          ResolverOptions{},
			key:           "VAR",
			expectedValue: "from-existing",
		},
		{
			name:          "OS env fallback when not strict",
			envFileVars:   map[string]string{},
			opts:          ResolverOptions{StrictEnv: false},
			key:           "OS_VAR",
			expectedValue: "from-os",
		},
		{
			name:          "no OS env fallback when strict",
			envFileVars:   map[string]string{},
			opts:          ResolverOptions{StrictEnv: true},
			key:           "OS_VAR",
			expectedValue: "",
		},
		{
			name:          "undefined var returns empty string",
			envFileVars:   map[string]string{},
			opts:          ResolverOptions{StrictEnv: true},
			key:           "UNDEFINED_VAR",
			expectedValue: "",
		},
		{
			name:        "env file var overrides existing resolver",
			envFileVars: map[string]string{"VAR": "from-env-file"},
			existingResolver: func(key string) (string, error) {
				if key == "VAR" {
					return "from-existing", nil
				}
				return "", nil
			},
			opts:          ResolverOptions{},
			key:           "VAR",
			expectedValue: "from-env-file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := buildEnvFileResolverWithOptions(tt.envFileVars, tt.existingResolver, tt.opts)
			value, err := resolver(tt.key)
			if err != nil {
				t.Fatalf("resolver error: %v", err)
			}
			if value != tt.expectedValue {
				t.Errorf("resolver(%q) = %q, want %q", tt.key, value, tt.expectedValue)
			}
		})
	}
}

func TestBuildEnvFileResolverWithTracking(t *testing.T) {
	// Set up test OS environment variable
	t.Setenv("OS_VAR", "from-os")

	tests := []struct {
		name             string
		envFileVars      map[string]string
		existingResolver func(string) (string, error)
		opts             ResolverOptions
		key              string
		expectedValue    string
		expectedSource   ResolutionSource
	}{
		{
			name:           "tracks env file source",
			envFileVars:    map[string]string{"VAR": "from-env-file"},
			opts:           ResolverOptions{},
			key:            "VAR",
			expectedValue:  "from-env-file",
			expectedSource: SourceEnvFile,
		},
		{
			name:        "tracks project config source",
			envFileVars: map[string]string{},
			existingResolver: func(key string) (string, error) {
				if key == "VAR" {
					return "from-project", nil
				}
				return "", nil
			},
			opts:           ResolverOptions{},
			key:            "VAR",
			expectedValue:  "from-project",
			expectedSource: SourceProjectConfig,
		},
		{
			name:           "tracks OS env source",
			envFileVars:    map[string]string{},
			opts:           ResolverOptions{StrictEnv: false},
			key:            "OS_VAR",
			expectedValue:  "from-os",
			expectedSource: SourceOSEnv,
		},
		{
			name:           "tracks not found",
			envFileVars:    map[string]string{},
			opts:           ResolverOptions{StrictEnv: true},
			key:            "UNDEFINED_VAR",
			expectedValue:  "",
			expectedSource: SourceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracking := buildEnvFileResolverWithTracking(tt.envFileVars, tt.existingResolver, tt.opts)
			value, err := tracking.Resolver(tt.key)
			if err != nil {
				t.Fatalf("resolver error: %v", err)
			}
			if value != tt.expectedValue {
				t.Errorf("resolver(%q) = %q, want %q", tt.key, value, tt.expectedValue)
			}
			if tracking.ResolutionSource[tt.key] != tt.expectedSource {
				t.Errorf("source = %v, want %v", tracking.ResolutionSource[tt.key], tt.expectedSource)
			}
		})
	}
}

func TestBuildEnvFileResolverWithTracking_OSFallbackTracking(t *testing.T) {
	t.Setenv("FALLBACK_VAR", "fallback-value")

	tracking := buildEnvFileResolverWithTracking(map[string]string{}, nil, ResolverOptions{})

	// Resolve a variable that falls back to OS env
	_, _ = tracking.Resolver("FALLBACK_VAR")

	// Check it's tracked in OSFallbackVars
	if len(tracking.OSFallbackVars) != 1 {
		t.Fatalf("expected 1 OS fallback var, got %d", len(tracking.OSFallbackVars))
	}
	if tracking.OSFallbackVars[0] != "FALLBACK_VAR" {
		t.Errorf("OSFallbackVars[0] = %q, want %q", tracking.OSFallbackVars[0], "FALLBACK_VAR")
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
		_, _ = loadFromStringInternal(input, "", nil, nil, ResolverOptions{})
	})
}
