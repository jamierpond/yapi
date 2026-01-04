package langserver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Note: TestFindEnvFilePathAtPosition was removed as part of Principle VII simplification.
// The functionality is now tested via validation.FindEnvFilesInConfig in the validation package.

func TestFindVarPositionInEnvFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a test .env file
	envContent := `# This is a comment
API_KEY=secret123
DATABASE_URL=postgres://localhost/db
# Another comment
DEBUG=true
`
	envPath := filepath.Join(tmpDir, ".env.test")
	if err := os.WriteFile(envPath, []byte(envContent), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		varName      string
		expectFound  bool
		expectedLine int
		expectedCol  int
	}{
		{
			name:         "find API_KEY",
			varName:      "API_KEY",
			expectFound:  true,
			expectedLine: 1,
			expectedCol:  0,
		},
		{
			name:         "find DATABASE_URL",
			varName:      "DATABASE_URL",
			expectFound:  true,
			expectedLine: 2,
			expectedCol:  0,
		},
		{
			name:         "find DEBUG",
			varName:      "DEBUG",
			expectFound:  true,
			expectedLine: 4,
			expectedCol:  0,
		},
		{
			name:        "variable not found",
			varName:     "NONEXISTENT",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location, err := findVarPositionInEnvFile(tmpDir, ".env.test", tt.varName)

			if tt.expectFound {
				if err != nil {
					t.Fatalf("findVarPositionInEnvFile() error = %v", err)
				}
				if location == nil {
					t.Fatal("findVarPositionInEnvFile() returned nil location")
				}
				if int(location.Range.Start.Line) != tt.expectedLine {
					t.Errorf("line = %d, want %d", location.Range.Start.Line, tt.expectedLine)
				}
				if int(location.Range.Start.Character) != tt.expectedCol {
					t.Errorf("col = %d, want %d", location.Range.Start.Character, tt.expectedCol)
				}
			} else if location != nil {
				t.Errorf("expected nil location for nonexistent variable, got %v", location)
			}
		})
	}
}

func TestResolveEnvFileLocation(t *testing.T) {
	// Create a temporary directory with an env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env.exists")
	if err := os.WriteFile(envPath, []byte("VAR=value"), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		envFilePath string
		expectFound bool
	}{
		{
			name:        "existing file",
			envFilePath: ".env.exists",
			expectFound: true,
		},
		{
			name:        "nonexistent file",
			envFilePath: ".env.notfound",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location := resolveEnvFileLocation(tt.envFilePath, tmpDir)

			if tt.expectFound {
				if location == nil {
					t.Error("expected location, got nil")
					return
				}
				// Should point to line 0, char 0
				if location.Range.Start.Line != 0 || location.Range.Start.Character != 0 {
					t.Errorf("expected position (0, 0), got (%d, %d)",
						location.Range.Start.Line, location.Range.Start.Character)
				}
			} else if location != nil {
				t.Errorf("expected nil for nonexistent file, got %v", location)
			}
		})
	}
}

func TestFindVariableInConfigEnvFiles(t *testing.T) {
	// Create a temp directory with env files
	tmpDir := t.TempDir()

	// Create .env.base with BASE_URL
	baseEnvPath := filepath.Join(tmpDir, ".env.base")
	if err := os.WriteFile(baseEnvPath, []byte("BASE_URL=http://localhost:3000\n"), 0600); err != nil {
		t.Fatal(err)
	}

	// Create .env.override with API_KEY
	overrideEnvPath := filepath.Join(tmpDir, ".env.override")
	if err := os.WriteFile(overrideEnvPath, []byte("API_KEY=secret123\n"), 0600); err != nil {
		t.Fatal(err)
	}

	configText := `yapi: v1
url: ${BASE_URL}/api
headers:
  Authorization: Bearer ${API_KEY}
env_files:
  - .env.base
  - .env.override`

	tests := []struct {
		name        string
		varName     string
		expectFound bool
		expectFile  string
	}{
		{
			name:        "find BASE_URL in .env.base",
			varName:     "BASE_URL",
			expectFound: true,
			expectFile:  ".env.base",
		},
		{
			name:        "find API_KEY in .env.override",
			varName:     "API_KEY",
			expectFound: true,
			expectFile:  ".env.override",
		},
		{
			name:        "variable not found",
			varName:     "NONEXISTENT",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location := findVariableInConfigEnvFiles(configText, tt.varName, tmpDir)

			if tt.expectFound {
				if location == nil {
					t.Error("expected location, got nil")
					return
				}
				// Check that the URI contains the expected file
				if !strings.Contains(location.URI, tt.expectFile) {
					t.Errorf("expected URI to contain %q, got %q", tt.expectFile, location.URI)
				}
			} else if location != nil {
				t.Errorf("expected nil for nonexistent variable, got %v", location)
			}
		})
	}
}
