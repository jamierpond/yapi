package utils

import "testing"

func TestIsBinaryContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "empty string",
			content:  "",
			expected: false,
		},
		{
			name:     "plain text",
			content:  "Hello, World!",
			expected: false,
		},
		{
			name:     "JSON content",
			content:  `{"key": "value", "number": 123}`,
			expected: false,
		},
		{
			name:     "text with newlines",
			content:  "Line 1\nLine 2\nLine 3",
			expected: false,
		},
		{
			name:     "null byte",
			content:  "Hello\x00World",
			expected: true,
		},
		{
			name:     "PNG signature",
			content:  string([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}),
			expected: true,
		},
		{
			name:     "JPEG signature",
			content:  string([]byte{0xFF, 0xD8, 0xFF, 0xE0}),
			expected: true,
		},
		{
			name:     "UTF-8 with emojis",
			content:  "Hello 👋 World 🌍",
			expected: false,
		},
		{
			name:     "high bytes (binary-like)",
			content:  string([]byte{0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89}),
			expected: true,
		},
		{
			name:     "control characters",
			content:  "Text\x01\x02\x03\x04\x05",
			expected: true,
		},
		{
			name:     "tabs and newlines are ok",
			content:  "Line 1\tColumn 2\nLine 2\tColumn 2\r\n",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBinaryContent(tt.content)
			if result != tt.expected {
				t.Errorf("IsBinaryContent(%q) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestIsBinaryContentLargeFile(t *testing.T) {
	// Test with a large text file (> 8KB sample size)
	largeText := make([]byte, 10000)
	for i := range largeText {
		largeText[i] = 'a'
	}
	if IsBinaryContent(string(largeText)) {
		t.Error("Large text file should not be detected as binary")
	}

	// Test with a large binary file
	largeBinary := make([]byte, 10000)
	for i := range largeBinary {
		largeBinary[i] = byte(i % 256)
	}
	if !IsBinaryContent(string(largeBinary)) {
		t.Error("Large binary file should be detected as binary")
	}
}
