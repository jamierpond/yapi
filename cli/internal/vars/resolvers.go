package vars

import (
	"os"
	"strings"
)

// EnvResolver resolves environment variables only.
func EnvResolver(key string) (string, error) {
	if val, ok := os.LookupEnv(key); ok {
		return val, nil
	}
	// Return empty string if not found - os.ExpandEnv behavior
	return "", nil
}

// MockResolver provides placeholder values for variable interpolation in LSP validation.
// This allows the compiler to validate the config structure even without real env vars.
func MockResolver(key string) (string, error) {
	keyLower := strings.ToLower(key)
	if strings.Contains(keyLower, "port") {
		return "8080", nil
	}
	if strings.Contains(keyLower, "host") {
		return "localhost", nil
	}
	if strings.Contains(keyLower, "url") {
		return "http://localhost:8080", nil
	}
	// Return a placeholder for anything else - don't fail
	return "PLACEHOLDER", nil
}
