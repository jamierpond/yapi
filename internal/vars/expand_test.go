package vars

import (
	"os"
	"testing"
)

func TestExpandAll(t *testing.T) {
	// Set up test environment variables
	os.Setenv("TEST_VAR", "expanded_value")
	os.Setenv("TEST_PORT", "8080")
	defer func() {
		os.Unsetenv("TEST_VAR")
		os.Unsetenv("TEST_PORT")
	}()

	t.Run("expands string fields", func(t *testing.T) {
		type Config struct {
			URL  string
			Path string
		}

		config := &Config{
			URL:  "http://$TEST_VAR/api",
			Path: "/v1/${TEST_VAR}",
		}

		ExpandAll(config, EnvResolver)

		if config.URL != "http://expanded_value/api" {
			t.Errorf("URL = %v, want http://expanded_value/api", config.URL)
		}
		if config.Path != "/v1/expanded_value" {
			t.Errorf("Path = %v, want /v1/expanded_value", config.Path)
		}
	})

	t.Run("expands map values", func(t *testing.T) {
		type Config struct {
			Headers map[string]string
		}

		config := &Config{
			Headers: map[string]string{
				"Host":   "$TEST_VAR",
				"Port":   "${TEST_PORT}",
				"Static": "no_expansion",
			},
		}

		ExpandAll(config, EnvResolver)

		if config.Headers["Host"] != "expanded_value" {
			t.Errorf("Headers[Host] = %v, want expanded_value", config.Headers["Host"])
		}
		if config.Headers["Port"] != "8080" {
			t.Errorf("Headers[Port] = %v, want 8080", config.Headers["Port"])
		}
		if config.Headers["Static"] != "no_expansion" {
			t.Errorf("Headers[Static] = %v, want no_expansion", config.Headers["Static"])
		}
	})

	t.Run("expands nested structs", func(t *testing.T) {
		type Inner struct {
			Value string
		}
		type Config struct {
			Outer Inner
		}

		config := &Config{
			Outer: Inner{Value: "$TEST_VAR"},
		}

		ExpandAll(config, EnvResolver)

		if config.Outer.Value != "expanded_value" {
			t.Errorf("Outer.Value = %v, want expanded_value", config.Outer.Value)
		}
	})

	t.Run("expands pointer to struct", func(t *testing.T) {
		type Inner struct {
			Value string
		}
		type Config struct {
			Ptr *Inner
		}

		config := &Config{
			Ptr: &Inner{Value: "$TEST_VAR"},
		}

		ExpandAll(config, EnvResolver)

		if config.Ptr.Value != "expanded_value" {
			t.Errorf("Ptr.Value = %v, want expanded_value", config.Ptr.Value)
		}
	})

	t.Run("handles nil pointer", func(t *testing.T) {
		type Inner struct {
			Value string
		}
		type Config struct {
			Ptr *Inner
		}

		config := &Config{
			Ptr: nil,
		}

		// Should not panic
		ExpandAll(config, EnvResolver)
	})

	t.Run("handles unexported fields", func(t *testing.T) {
		type Config struct {
			Public  string
			private string
		}

		config := &Config{
			Public:  "$TEST_VAR",
			private: "$TEST_VAR",
		}

		// Should not panic on unexported field
		ExpandAll(config, EnvResolver)

		if config.Public != "expanded_value" {
			t.Errorf("Public = %v, want expanded_value", config.Public)
		}
		// private field should remain unchanged (can't be set via reflection)
		if config.private != "$TEST_VAR" {
			t.Errorf("private = %v, want $TEST_VAR", config.private)
		}
	})

	t.Run("handles non-string map values", func(t *testing.T) {
		type Config struct {
			IntMap map[string]int
		}

		config := &Config{
			IntMap: map[string]int{
				"key": 42,
			},
		}

		// Should not panic
		ExpandAll(config, EnvResolver)

		if config.IntMap["key"] != 42 {
			t.Errorf("IntMap[key] = %v, want 42", config.IntMap["key"])
		}
	})

	t.Run("handles non-struct input", func(t *testing.T) {
		str := "test"
		// Should not panic
		ExpandAll(&str, EnvResolver)
	})

	t.Run("handles nil input", func(t *testing.T) {
		// Should not panic
		ExpandAll(nil, EnvResolver)
	})

	t.Run("uses custom resolver", func(t *testing.T) {
		type Config struct {
			Value string
		}

		config := &Config{
			Value: "$CUSTOM",
		}

		customResolver := func(key string) (string, error) {
			if key == "CUSTOM" {
				return "custom_value", nil
			}
			return "", nil
		}

		ExpandAll(config, customResolver)

		if config.Value != "custom_value" {
			t.Errorf("Value = %v, want custom_value", config.Value)
		}
	})

	t.Run("handles resolver errors gracefully", func(t *testing.T) {
		type Config struct {
			Value string
		}

		config := &Config{
			Value: "$NONEXISTENT",
		}

		// EnvResolver returns empty string for non-existent vars
		ExpandAll(config, EnvResolver)

		// Should expand to empty string
		if config.Value != "" {
			t.Errorf("Value = %v, want empty string", config.Value)
		}
	})
}

func TestExpandAll_ComplexStruct(t *testing.T) {
	os.Setenv("TEST_URL", "example.com")
	os.Setenv("TEST_KEY", "secret")
	defer func() {
		os.Unsetenv("TEST_URL")
		os.Unsetenv("TEST_KEY")
	}()

	type NestedConfig struct {
		API string
	}

	type ComplexConfig struct {
		URL     string
		Headers map[string]string
		Nested  NestedConfig
		Ptr     *NestedConfig
	}

	config := &ComplexConfig{
		URL: "https://$TEST_URL",
		Headers: map[string]string{
			"Authorization": "Bearer $TEST_KEY",
			"Host":          "$TEST_URL",
		},
		Nested: NestedConfig{
			API: "https://$TEST_URL/api",
		},
		Ptr: &NestedConfig{
			API: "https://$TEST_URL/v2",
		},
	}

	ExpandAll(config, EnvResolver)

	if config.URL != "https://example.com" {
		t.Errorf("URL = %v, want https://example.com", config.URL)
	}
	if config.Headers["Authorization"] != "Bearer secret" {
		t.Errorf("Headers[Authorization] = %v, want Bearer secret", config.Headers["Authorization"])
	}
	if config.Headers["Host"] != "example.com" {
		t.Errorf("Headers[Host] = %v, want example.com", config.Headers["Host"])
	}
	if config.Nested.API != "https://example.com/api" {
		t.Errorf("Nested.API = %v, want https://example.com/api", config.Nested.API)
	}
	if config.Ptr.API != "https://example.com/v2" {
		t.Errorf("Ptr.API = %v, want https://example.com/v2", config.Ptr.API)
	}
}
