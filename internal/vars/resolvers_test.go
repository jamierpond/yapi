package vars

import (
	"os"
	"testing"
)

func TestEnvResolver(t *testing.T) {
	// Set up test environment variables
	os.Setenv("TEST_VAR", "test_value")
	os.Setenv("EMPTY_VAR", "")
	defer func() {
		os.Unsetenv("TEST_VAR")
		os.Unsetenv("EMPTY_VAR")
	}()

	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{
			name:    "existing variable",
			key:     "TEST_VAR",
			want:    "test_value",
			wantErr: false,
		},
		{
			name:    "empty variable",
			key:     "EMPTY_VAR",
			want:    "",
			wantErr: false,
		},
		{
			name:    "non-existent variable",
			key:     "NONEXISTENT",
			want:    "",
			wantErr: false,
		},
		{
			name:    "PATH variable should exist",
			key:     "PATH",
			want:    "", // Don't check exact value, just that it doesn't error
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnvResolver(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnvResolver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// For PATH, just check it doesn't error
			if tt.key == "PATH" {
				return
			}
			if got != tt.want {
				t.Errorf("EnvResolver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvResolver_Integration(t *testing.T) {
	os.Setenv("API_KEY", "secret123")
	defer os.Unsetenv("API_KEY")

	val, err := EnvResolver("API_KEY")
	if err != nil {
		t.Fatalf("EnvResolver() error = %v", err)
	}
	if val != "secret123" {
		t.Errorf("EnvResolver() = %v, want secret123", val)
	}
}

func TestMockResolver(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{
			name:    "port key",
			key:     "PORT",
			want:    "8080",
			wantErr: false,
		},
		{
			name:    "port in lowercase",
			key:     "port",
			want:    "8080",
			wantErr: false,
		},
		{
			name:    "port with prefix",
			key:     "API_PORT",
			want:    "8080",
			wantErr: false,
		},
		{
			name:    "host key",
			key:     "HOST",
			want:    "localhost",
			wantErr: false,
		},
		{
			name:    "host in mixed case",
			key:     "Host",
			want:    "localhost",
			wantErr: false,
		},
		{
			name:    "host with suffix",
			key:     "DB_HOST",
			want:    "localhost",
			wantErr: false,
		},
		{
			name:    "url key",
			key:     "URL",
			want:    "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "url in lowercase",
			key:     "url",
			want:    "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "url with prefix",
			key:     "API_URL",
			want:    "http://localhost:8080",
			wantErr: false,
		},
		{
			name:    "unknown key returns placeholder",
			key:     "UNKNOWN_VAR",
			want:    "PLACEHOLDER",
			wantErr: false,
		},
		{
			name:    "empty key returns placeholder",
			key:     "",
			want:    "PLACEHOLDER",
			wantErr: false,
		},
		{
			name:    "random string returns placeholder",
			key:     "SOME_RANDOM_KEY",
			want:    "PLACEHOLDER",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MockResolver(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("MockResolver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MockResolver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMockResolver_CaseInsensitive(t *testing.T) {
	variations := []string{"PORT", "port", "Port", "pOrT"}

	for _, key := range variations {
		t.Run(key, func(t *testing.T) {
			val, err := MockResolver(key)
			if err != nil {
				t.Errorf("MockResolver(%v) error = %v", key, err)
			}
			if val != "8080" {
				t.Errorf("MockResolver(%v) = %v, want 8080", key, val)
			}
		})
	}
}

func TestMockResolver_Priority(t *testing.T) {
	// Test that specific keywords take precedence
	tests := []struct {
		key  string
		want string
	}{
		{"port", "8080"},                      // port has priority
		{"host", "localhost"},                 // host has priority
		{"url", "http://localhost:8080"},      // url has priority
		{"api_port", "8080"},                  // port keyword in string
		{"server_host", "localhost"},          // host keyword in string
		{"base_url", "http://localhost:8080"}, // url keyword in string
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			val, _ := MockResolver(tt.key)
			if val != tt.want {
				t.Errorf("MockResolver(%v) = %v, want %v", tt.key, val, tt.want)
			}
		})
	}
}

func TestResolvers_WithExpandString(t *testing.T) {
	os.Setenv("TEST_ENV", "env_value")
	defer os.Unsetenv("TEST_ENV")

	t.Run("EnvResolver with ExpandString", func(t *testing.T) {
		input := "Value is $TEST_ENV"
		result, err := ExpandString(input, EnvResolver)
		if err != nil {
			t.Fatalf("ExpandString() error = %v", err)
		}
		if result != "Value is env_value" {
			t.Errorf("ExpandString() = %v, want 'Value is env_value'", result)
		}
	})

	t.Run("MockResolver with ExpandString", func(t *testing.T) {
		input := "Server at $HOST:$PORT"
		result, err := ExpandString(input, MockResolver)
		if err != nil {
			t.Fatalf("ExpandString() error = %v", err)
		}
		if result != "Server at localhost:8080" {
			t.Errorf("ExpandString() = %v, want 'Server at localhost:8080'", result)
		}
	})
}
