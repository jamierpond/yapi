package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestFindVarPositionInYAML(t *testing.T) {
	// Create a temporary yapi.config.yml
	content := `yapi: v1
environments:
  dev:
    vars:
      API_KEY: "12345"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "yapi.config.yml")
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		varName string
		section []string
		want    bool // found?
		line    int
	}{
		{
			name:    "found variable",
			varName: "API_KEY",
			section: []string{"environments", "dev", "vars"},
			want:    true,
			line:    4, // 0-indexed: line 5 in file (API_KEY is on line 5)
		},
		{
			name:    "missing variable",
			varName: "MISSING_VAR",
			section: []string{"environments", "dev", "vars"},
			want:    false,
		},
		{
			name:    "missing section",
			varName: "API_KEY",
			section: []string{"environments", "prod", "vars"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := FindVarPositionInYAML(tmpDir, tt.varName, tt.section)
			if err != nil {
				// We expect error return for not found
				if tt.want {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}

			if !tt.want {
				if loc != nil {
					t.Errorf("expected nil location, got %+v", loc)
				}
				return
			}

			if loc == nil {
				t.Fatal("expected location, got nil")
			}

			if loc.Line != tt.line {
				t.Errorf("expected line %d, got %d", tt.line, loc.Line)
			}
			if loc.File != configPath {
				t.Errorf("expected file %s, got %s", configPath, loc.File)
			}
		})
	}
}

func TestFindVarPositionInYAML_YamlExtension(t *testing.T) {
	// Test with .yaml extension
	content := `yapi: v1
environments:
  prod:
    vars:
      SECRET: "abc123"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "yapi.config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	loc, err := FindVarPositionInYAML(tmpDir, "SECRET", []string{"environments", "prod", "vars"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loc == nil {
		t.Fatal("expected location, got nil")
	}
	if loc.File != configPath {
		t.Errorf("expected file %s, got %s", configPath, loc.File)
	}
}

func TestFindVarPositionInYAML_NoConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := FindVarPositionInYAML(tmpDir, "API_KEY", []string{"environments", "dev", "vars"})
	if err == nil {
		t.Error("expected error for missing config file, got nil")
	}
	// Check error message includes both file names
	if err != nil && (!contains(err.Error(), "yapi.config.yml") || !contains(err.Error(), "yapi.config.yaml")) {
		t.Errorf("error message should mention both config file names, got: %v", err)
	}
}

func TestFindNodeInMapping(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		key     string
		wantNil bool
		wantVal string
	}{
		{
			name:    "find existing key",
			yaml:    "foo: bar\nbaz: qux",
			key:     "foo",
			wantNil: false,
			wantVal: "bar",
		},
		{
			name:    "key not found",
			yaml:    "foo: bar",
			key:     "missing",
			wantNil: true,
		},
		{
			name:    "nil node",
			yaml:    "",
			key:     "foo",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.yaml == "" {
				result := FindNodeInMapping(nil, tt.key)
				if result != nil {
					t.Error("expected nil for nil node")
				}
				return
			}

			var node yamlNode
			if err := yamlUnmarshal([]byte(tt.yaml), &node); err != nil {
				t.Fatalf("failed to parse YAML: %v", err)
			}

			doc := node.Content[0]
			result := FindNodeInMapping(doc, tt.key)

			if tt.wantNil {
				if result != nil {
					t.Error("expected nil result")
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Value != tt.wantVal {
				t.Errorf("expected value %q, got %q", tt.wantVal, result.Value)
			}
		})
	}
}

func TestFindKeyNodeInMapping(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		key     string
		wantNil bool
	}{
		{
			name:    "find existing key node",
			yaml:    "foo: bar\nbaz: qux",
			key:     "foo",
			wantNil: false,
		},
		{
			name:    "key not found",
			yaml:    "foo: bar",
			key:     "missing",
			wantNil: true,
		},
		{
			name:    "nil node",
			yaml:    "",
			key:     "foo",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.yaml == "" {
				result := FindKeyNodeInMapping(nil, tt.key)
				if result != nil {
					t.Error("expected nil for nil node")
				}
				return
			}

			var node yamlNode
			if err := yamlUnmarshal([]byte(tt.yaml), &node); err != nil {
				t.Fatalf("failed to parse YAML: %v", err)
			}

			doc := node.Content[0]
			result := FindKeyNodeInMapping(doc, tt.key)

			if tt.wantNil {
				if result != nil {
					t.Error("expected nil result")
				}
				return
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}
			// The key node's value should be the key itself
			if result.Value != tt.key {
				t.Errorf("expected key node value %q, got %q", tt.key, result.Value)
			}
		})
	}
}

func TestFindSectionPosition(t *testing.T) {
	content := `yapi: v1
environments:
  dev:
    vars:
      API_KEY: "12345"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "yapi.config.yml")
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		section []string
		wantErr bool
	}{
		{
			name:    "find existing section",
			section: []string{"environments", "dev", "vars"},
			wantErr: false,
		},
		{
			name:    "empty section",
			section: []string{},
			wantErr: true,
		},
		{
			name:    "missing section",
			section: []string{"environments", "prod"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc, err := FindSectionPosition(tmpDir, tt.section)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if loc == nil {
				t.Error("expected location, got nil")
			}
		})
	}
}

// Helper to check string containment
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Type aliases for yaml.Node to avoid import in test
type yamlNode = yaml.Node

var yamlUnmarshal = yaml.Unmarshal
