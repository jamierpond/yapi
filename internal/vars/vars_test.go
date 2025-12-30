package vars

import (
	"fmt"
	"testing"
)

func TestExpandString(t *testing.T) {
	resolver := func(key string) (string, error) {
		switch key {
		case "FOO":
			return "bar", nil
		case "NESTED":
			return "nested_value", nil
		case "FOO123":
			return "bar123", nil
		default:
			return "", fmt.Errorf("unknown var: %s", key)
		}
	}

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"no vars", "hello world", "hello world", false},
		{"simple $VAR", "hello $FOO", "hello bar", false},
		{"simple ${VAR}", "hello ${FOO}", "hello bar", false},
		{"multiple vars", "$FOO and ${NESTED}", "bar and nested_value", false},
		{"unknown var", "$UNKNOWN", "", true},
		{"empty string", "", "", false},
		// Bcrypt hashes should NOT be treated as variables
		{"bcrypt hash", "$2a$12$k0LsiR40ZNcMxbyD80g5nebjB9R0/VqilwfFLLr6m/XTOc9WRf8Om", "$2a$12$k0LsiR40ZNcMxbyD80g5nebjB9R0/VqilwfFLLr6m/XTOc9WRf8Om", false},
		{"bcrypt hash $2y", "$2y$10$abcdefghijklmnopqrstuv1234567890ABCDEFGHIJKLMNOPQR", "$2y$10$abcdefghijklmnopqrstuv1234567890ABCDEFGHIJKLMNOPQR", false},
		// Dollar signs followed by digits should be preserved
		{"dollar digit", "price is $100", "price is $100", false},
		{"dollar digits only", "$123", "$123", false},
		{"dollar with numbers", "$1test", "$1test", false},
		// Valid variables should still work
		{"var starts with letter", "$FOO123", "bar123", false},
		{"var starts with underscore", "${_private}", "", true}, // _private not defined
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandString(tt.input, resolver)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExpandString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ExpandString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func FuzzExpandString(f *testing.F) {
	// Seed with various variable patterns
	f.Add("hello world")
	f.Add("$VAR")
	f.Add("${VAR}")
	f.Add("$VAR1 $VAR2 $VAR3")
	f.Add("${nested.value}")
	f.Add("$step.response.body")
	f.Add("prefix_${VAR}_suffix")
	f.Add("$")
	f.Add("${}")
	f.Add("${unclosed")
	f.Add("$$VAR")
	f.Add("$123")
	f.Add("${123}")
	f.Add("${a.b.c.d.e}")
	f.Add("https://example.com/$PATH")
	f.Add(`{"key": "${VALUE}"}`)

	// Resolver that always succeeds
	resolver := func(key string) (string, error) {
		return "RESOLVED:" + key, nil
	}

	f.Fuzz(func(t *testing.T, input string) {
		// ExpandString should not panic on any input
		_, _ = ExpandString(input, resolver)
	})
}

func FuzzHasChainVars(f *testing.F) {
	f.Add("$step.field")
	f.Add("${step.response.body}")
	f.Add("$VAR")
	f.Add("${VAR}")
	f.Add("no vars here")
	f.Add("$a.b.c.d")
	f.Add("")

	f.Fuzz(func(t *testing.T, input string) {
		// HasChainVars should not panic
		_ = HasChainVars(input)
	})
}

func FuzzHasEnvVars(f *testing.F) {
	f.Add("$HOME")
	f.Add("${PATH}")
	f.Add("no vars")
	f.Add("$123invalid")
	f.Add("")

	f.Fuzz(func(t *testing.T, input string) {
		// HasEnvVars should not panic
		_ = HasEnvVars(input)
	})
}
