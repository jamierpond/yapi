package importer

import (
	"testing"
)

func TestConvertVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple variable",
			input:    "{{apiKey}}",
			expected: "${apiKey}",
		},
		{
			name:     "multiple variables",
			input:    "{{baseUrl}}/users/{{userId}}",
			expected: "${baseUrl}/users/${userId}",
		},
		{
			name:     "no variables",
			input:    "https://api.example.com/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "postman dynamic variable - guid",
			input:    "{{$guid}}",
			expected: "${$guid}",
		},
		{
			name:     "postman dynamic variable - timestamp",
			input:    "{{$timestamp}}",
			expected: "${$timestamp}",
		},
		{
			name:     "postman dynamic variable - randomInt",
			input:    "{{$randomInt}}",
			expected: "${$randomInt}",
		},
		{
			name:     "postman dynamic variable - isoTimestamp",
			input:    "{{$isoTimestamp}}",
			expected: "${$isoTimestamp}",
		},
		{
			name:     "mixed regular and dynamic variables",
			input:    "{{baseUrl}}/api/{{$guid}}/{{userId}}",
			expected: "${baseUrl}/api/${$guid}/${userId}",
		},
		{
			name:     "variable with underscores",
			input:    "{{api_key}}",
			expected: "${api_key}",
		},
		{
			name:     "variable with numbers",
			input:    "{{api2Key}}",
			expected: "${api2Key}",
		},
		{
			name:     "variable with hyphens",
			input:    "{{api-key}}",
			expected: "${api-key}",
		},
		{
			name:     "nested braces in string",
			input:    "text {{{not-a-var}}} more text",
			expected: "text ${{not-a-var}} more text",
		},
		{
			name:     "special characters around variables",
			input:    "Authorization: Bearer {{token}}",
			expected: "Authorization: Bearer ${token}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertVariables(tt.input)
			if result != tt.expected {
				t.Errorf("convertVariables(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple name",
			input:    "users",
			expected: "users",
		},
		{
			name:     "name with spaces",
			input:    "Get Users",
			expected: "get-users",
		},
		{
			name:     "name with special characters",
			input:    "Get Users (v2)!",
			expected: "get-users-v2",
		},
		{
			name:     "directory traversal attempt with dots",
			input:    "../../etc/passwd",
			expected: "etcpasswd",
		},
		{
			name:     "directory traversal with slashes",
			input:    "../../../root",
			expected: "root",
		},
		{
			name:     "absolute path attempt",
			input:    "/etc/passwd",
			expected: "etcpasswd",
		},
		{
			name:     "windows path attempt",
			input:    "..\\..\\windows\\system32",
			expected: "windowssystem32",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "unnamed",
		},
		{
			name:     "only dots",
			input:    "...",
			expected: "unnamed",
		},
		{
			name:     "only special characters",
			input:    "!@#$%^&*()",
			expected: "unnamed",
		},
		{
			name:     "name with underscores",
			input:    "get_users",
			expected: "get_users",
		},
		{
			name:     "name with hyphens",
			input:    "get-users",
			expected: "get-users",
		},
		{
			name:     "mixed case",
			input:    "GetUsers",
			expected: "getusers",
		},
		{
			name:     "very long name",
			input:    "this-is-a-very-long-filename-that-exceeds-the-maximum-allowed-length-and-should-be-truncated-to-fit-within-the-limit-of-two-hundred-characters-which-is-quite-a-lot-but-we-want-to-make-sure-it-works-correctly-with-really-long-names-that-users-might-create",
			expected: "this-is-a-very-long-filename-that-exceeds-the-maximum-allowed-length-and-should-be-truncated-to-fit-within-the-limit-of-two-hundred-characters-which-is-quite-a-lot-but-we-want-to-make-sure-it-works-co",
		},
		{
			name:     "unicode characters",
			input:    "ユーザー取得",
			expected: "unnamed",
		},
		{
			name:     "name with dots (should be removed)",
			input:    "file.name.ext",
			expected: "filenameext",
		},
		{
			name:     "null bytes attempt",
			input:    "file\x00name",
			expected: "filename",
		},
		{
			name:     "current directory reference",
			input:    ".",
			expected: "unnamed",
		},
		{
			name:     "parent directory reference",
			input:    "..",
			expected: "unnamed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFileName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFileName(%q) = %q, want %q", tt.input, result, tt.expected)
			}

			// Additional security checks
			if len(result) > 200 {
				t.Errorf("sanitizeFileName(%q) resulted in name longer than 200 chars: %d", tt.input, len(result))
			}

			// Ensure no path separators
			if containsPathSeparator(result) {
				t.Errorf("sanitizeFileName(%q) = %q contains path separator", tt.input, result)
			}

			// Ensure no dots
			if containsDots(result) {
				t.Errorf("sanitizeFileName(%q) = %q contains dots (potential directory traversal)", tt.input, result)
			}
		})
	}
}

// Helper function to check for path separators
func containsPathSeparator(s string) bool {
	return len(s) > 0 && (s[0] == '/' || s[0] == '\\' ||
		len(s) > 1 && (s[len(s)-1] == '/' || s[len(s)-1] == '\\'))
}

// Helper function to check for dots
func containsDots(s string) bool {
	for _, c := range s {
		if c == '.' {
			return true
		}
	}
	return false
}

func TestSanitizeFileNameSecurity(t *testing.T) {
	// Additional security-focused tests
	maliciousInputs := []string{
		"../../root/.ssh/id_rsa",
		"../../../etc/shadow",
		"/var/log/../../etc/passwd",
		"..\\..\\windows\\system32\\config\\sam",
		"./.env",
		"../config.yml",
		"./../../secret.key",
	}

	for _, input := range maliciousInputs {
		t.Run("security_"+input, func(t *testing.T) {
			result := sanitizeFileName(input)

			// Ensure result doesn't contain path separators
			if containsPathSeparator(result) {
				t.Errorf("SECURITY: sanitizeFileName(%q) = %q still contains path separator", input, result)
			}

			// Ensure result doesn't contain dots
			if containsDots(result) {
				t.Errorf("SECURITY: sanitizeFileName(%q) = %q still contains dots", input, result)
			}

			// Ensure result is not empty or dangerous
			if result == "" || result == "." || result == ".." {
				t.Errorf("SECURITY: sanitizeFileName(%q) = %q is potentially dangerous", input, result)
			}
		})
	}
}
