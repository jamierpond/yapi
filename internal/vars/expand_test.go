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

func TestExpandAll_MapStringAny(t *testing.T) {
	os.Setenv("TEST_VAR", "expanded_value")
	os.Setenv("TEST_NUM", "42")
	defer func() {
		os.Unsetenv("TEST_VAR")
		os.Unsetenv("TEST_NUM")
	}()

	t.Run("expands string values in map[string]any", func(t *testing.T) {
		type Config struct {
			Body map[string]any
		}

		config := &Config{
			Body: map[string]any{
				"name":   "$TEST_VAR",
				"value":  "${TEST_VAR}",
				"static": "no_change",
			},
		}

		ExpandAll(config, EnvResolver)

		if config.Body["name"] != "expanded_value" {
			t.Errorf("Body[name] = %v, want expanded_value", config.Body["name"])
		}
		if config.Body["value"] != "expanded_value" {
			t.Errorf("Body[value] = %v, want expanded_value", config.Body["value"])
		}
		if config.Body["static"] != "no_change" {
			t.Errorf("Body[static] = %v, want no_change", config.Body["static"])
		}
	})

	t.Run("preserves non-string types in map[string]any", func(t *testing.T) {
		type Config struct {
			Body map[string]any
		}

		config := &Config{
			Body: map[string]any{
				"number": 42,
				"float":  3.14,
				"bool":   true,
				"null":   nil,
			},
		}

		ExpandAll(config, EnvResolver)

		if config.Body["number"] != 42 {
			t.Errorf("Body[number] = %v, want 42", config.Body["number"])
		}
		if config.Body["float"] != 3.14 {
			t.Errorf("Body[float] = %v, want 3.14", config.Body["float"])
		}
		if config.Body["bool"] != true {
			t.Errorf("Body[bool] = %v, want true", config.Body["bool"])
		}
		if config.Body["null"] != nil {
			t.Errorf("Body[null] = %v, want nil", config.Body["null"])
		}
	})

	t.Run("expands nested maps in map[string]any", func(t *testing.T) {
		type Config struct {
			Body map[string]any
		}

		config := &Config{
			Body: map[string]any{
				"nested": map[string]any{
					"key": "$TEST_VAR",
					"deep": map[string]any{
						"value": "${TEST_VAR}",
					},
				},
			},
		}

		ExpandAll(config, EnvResolver)

		nested := config.Body["nested"].(map[string]any)
		if nested["key"] != "expanded_value" {
			t.Errorf("Body[nested][key] = %v, want expanded_value", nested["key"])
		}

		deep := nested["deep"].(map[string]any)
		if deep["value"] != "expanded_value" {
			t.Errorf("Body[nested][deep][value] = %v, want expanded_value", deep["value"])
		}
	})

	t.Run("expands arrays in map[string]any", func(t *testing.T) {
		type Config struct {
			Body map[string]any
		}

		config := &Config{
			Body: map[string]any{
				"array": []any{
					"$TEST_VAR",
					"${TEST_VAR}",
					42,
					true,
					map[string]any{
						"nested": "$TEST_VAR",
					},
				},
			},
		}

		ExpandAll(config, EnvResolver)

		arr := config.Body["array"].([]any)
		if arr[0] != "expanded_value" {
			t.Errorf("Body[array][0] = %v, want expanded_value", arr[0])
		}
		if arr[1] != "expanded_value" {
			t.Errorf("Body[array][1] = %v, want expanded_value", arr[1])
		}
		if arr[2] != 42 {
			t.Errorf("Body[array][2] = %v, want 42", arr[2])
		}
		if arr[3] != true {
			t.Errorf("Body[array][3] = %v, want true", arr[3])
		}
		nested := arr[4].(map[string]any)
		if nested["nested"] != "expanded_value" {
			t.Errorf("Body[array][4][nested] = %v, want expanded_value", nested["nested"])
		}
	})

	t.Run("preserves nil map[string]any", func(t *testing.T) {
		type Config struct {
			Body map[string]any
		}

		config := &Config{
			Body: nil,
		}

		ExpandAll(config, EnvResolver)

		if config.Body != nil {
			t.Errorf("Body should be nil, got %v", config.Body)
		}
	})

	t.Run("handles empty map[string]any", func(t *testing.T) {
		type Config struct {
			Body map[string]any
		}

		config := &Config{
			Body: map[string]any{},
		}

		ExpandAll(config, EnvResolver)

		if config.Body == nil {
			t.Error("Body should not be nil")
		}
		if len(config.Body) != 0 {
			t.Errorf("Body should be empty, got %v", config.Body)
		}
	})

	t.Run("handles mixed string interpolation and literals", func(t *testing.T) {
		type Config struct {
			Body map[string]any
		}

		config := &Config{
			Body: map[string]any{
				"url":      "https://$TEST_VAR/api",
				"message":  "Value is: ${TEST_VAR}",
				"template": "Start ${TEST_VAR} end",
			},
		}

		ExpandAll(config, EnvResolver)

		if config.Body["url"] != "https://expanded_value/api" {
			t.Errorf("Body[url] = %v, want https://expanded_value/api", config.Body["url"])
		}
		if config.Body["message"] != "Value is: expanded_value" {
			t.Errorf("Body[message] = %v, want 'Value is: expanded_value'", config.Body["message"])
		}
		if config.Body["template"] != "Start expanded_value end" {
			t.Errorf("Body[template] = %v, want 'Start expanded_value end'", config.Body["template"])
		}
	})

	t.Run("handles undefined variables", func(t *testing.T) {
		type Config struct {
			Body map[string]any
		}

		config := &Config{
			Body: map[string]any{
				"undefined": "$UNDEFINED_VAR",
			},
		}

		ExpandAll(config, EnvResolver)

		// EnvResolver returns empty string for undefined vars
		if config.Body["undefined"] != "" {
			t.Errorf("Body[undefined] = %v, want empty string", config.Body["undefined"])
		}
	})

	t.Run("complex real-world scenario", func(t *testing.T) {
		type Config struct {
			Body      map[string]any
			Variables map[string]any // GraphQL variables
		}

		config := &Config{
			Body: map[string]any{
				"query": "query { user { name } }",
				"user": map[string]any{
					"name":  "$TEST_VAR",
					"id":    123,
					"email": "test@${TEST_VAR}.com",
					"settings": map[string]any{
						"theme": "$TEST_VAR",
						"lang":  "en",
					},
				},
				"tags": []any{"$TEST_VAR", "static", "${TEST_VAR}"},
			},
			Variables: map[string]any{
				"id":   "${TEST_NUM}",
				"name": "$TEST_VAR",
			},
		}

		ExpandAll(config, EnvResolver)

		// Check Body
		if config.Body["query"] != "query { user { name } }" {
			t.Errorf("Body[query] = %v, want 'query { user { name } }'", config.Body["query"])
		}

		user := config.Body["user"].(map[string]any)
		if user["name"] != "expanded_value" {
			t.Errorf("Body[user][name] = %v, want expanded_value", user["name"])
		}
		if user["id"] != 123 {
			t.Errorf("Body[user][id] = %v, want 123", user["id"])
		}
		if user["email"] != "test@expanded_value.com" {
			t.Errorf("Body[user][email] = %v, want test@expanded_value.com", user["email"])
		}

		settings := user["settings"].(map[string]any)
		if settings["theme"] != "expanded_value" {
			t.Errorf("Body[user][settings][theme] = %v, want expanded_value", settings["theme"])
		}
		if settings["lang"] != "en" {
			t.Errorf("Body[user][settings][lang] = %v, want en", settings["lang"])
		}

		tags := config.Body["tags"].([]any)
		if tags[0] != "expanded_value" {
			t.Errorf("Body[tags][0] = %v, want expanded_value", tags[0])
		}
		if tags[1] != "static" {
			t.Errorf("Body[tags][1] = %v, want static", tags[1])
		}
		if tags[2] != "expanded_value" {
			t.Errorf("Body[tags][2] = %v, want expanded_value", tags[2])
		}

		// Check Variables
		if config.Variables["id"] != "42" {
			t.Errorf("Variables[id] = %v, want 42", config.Variables["id"])
		}
		if config.Variables["name"] != "expanded_value" {
			t.Errorf("Variables[name] = %v, want expanded_value", config.Variables["name"])
		}
	})
}
