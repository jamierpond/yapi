package share

import (
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple", "hello world"},
		{"yapi config", `POST https://api.example.com/users
Content-Type: application/json

{
  "name": "test",
  "email": "test@example.com"
}`},
		{"unicode", "日本語テスト 🚀"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := Encode(tc.input)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}
			if encoded == "" {
				t.Error("Encode returned empty string")
			}
		})
	}
}

func TestEncode_Empty(t *testing.T) {
	encoded, err := Encode("")
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	// Empty input produces empty encoding (after gzip)
	if encoded == "" {
		t.Skip("empty input produces empty encoding")
	}
}
