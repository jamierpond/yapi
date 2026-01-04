package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"yapi.run/cli/internal/config"
)

func hasDiagnostic(diags []Diagnostic, substr string) bool {
	for _, d := range diags {
		if strings.Contains(d.Message, substr) {
			return true
		}
	}
	return false
}

func TestAnalyzeConfig_ValidHTTP(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
method: GET`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if len(a.Diagnostics) != 0 {
		t.Errorf("expected no diagnostics for valid HTTP config, got %d: %+v", len(a.Diagnostics), a.Diagnostics)
	}

	if a.Request == nil {
		t.Fatal("expected Request to be populated")
	}
}

func TestAnalyzeConfig_MissingURL(t *testing.T) {
	yaml := `yapi: v1
method: GET`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for missing URL")
	}

	if !hasDiagnostic(a.Diagnostics, "missing required field") {
		t.Errorf("expected 'missing required field' message, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_BadYAML(t *testing.T) {
	yaml := `yapi: v1
url: [invalid yaml`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for invalid YAML")
	}

	if !hasDiagnostic(a.Diagnostics, "invalid YAML") {
		t.Errorf("expected 'invalid YAML' message, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_BadGraphQL(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com/graphql
graphql: |
  query { foo( }`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if !hasDiagnostic(a.Diagnostics, "GraphQL syntax error") {
		t.Fatalf("expected GraphQL syntax error, got %+v", a.Diagnostics)
	}

	// Verify line number is set for GraphQL diagnostic
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "GraphQL") && d.Line < 0 {
			t.Errorf("expected GraphQL diagnostic to have line number set")
		}
	}
}

func TestAnalyzeConfig_ValidGraphQL(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com/graphql
graphql: |
  query { foo }`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	// Should have no GraphQL syntax errors
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "GraphQL") {
			t.Errorf("unexpected GraphQL diagnostic: %s", d.Message)
		}
	}
}

func TestAnalyzeConfig_BadJQ(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
jq_filter: .foo[`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if !hasDiagnostic(a.Diagnostics, "JQ syntax error") {
		t.Fatalf("expected JQ syntax error, got %+v", a.Diagnostics)
	}

	// Verify line number is set for JQ diagnostic
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "JQ") && d.Line < 0 {
			t.Errorf("expected JQ diagnostic to have line number set")
		}
	}
}

func TestAnalyzeConfig_ValidJQ(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
jq_filter: .data.items[]`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	// Should have no JQ syntax errors
	for _, d := range a.Diagnostics {
		if strings.Contains(d.Message, "JQ") {
			t.Errorf("unexpected JQ diagnostic: %s", d.Message)
		}
	}
}

func TestAnalyzeConfig_MissingVersion(t *testing.T) {
	yaml := `url: http://example.com
method: GET`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if len(a.Warnings) == 0 {
		t.Error("expected warning for missing yapi version")
	}

	found := false
	for _, w := range a.Warnings {
		if strings.Contains(w, "Missing") && strings.Contains(w, "v1") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected warning about missing version, got %+v", a.Warnings)
	}
}

func TestAnalyzeConfig_GRPCMissingRequirements(t *testing.T) {
	yaml := `yapi: v1
url: grpc://localhost:50051`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for gRPC missing service/rpc")
	}

	if !hasDiagnostic(a.Diagnostics, "service") {
		t.Errorf("expected error about missing service, got %+v", a.Diagnostics)
	}

	if !hasDiagnostic(a.Diagnostics, "rpc") {
		t.Errorf("expected error about missing rpc, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_TCPInvalidEncoding(t *testing.T) {
	yaml := `yapi: v1
url: tcp://localhost:9000
data: hello
encoding: invalid`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected errors for invalid TCP encoding")
	}

	if !hasDiagnostic(a.Diagnostics, "unsupported TCP encoding") {
		t.Errorf("expected error about unsupported encoding, got %+v", a.Diagnostics)
	}
}

func TestAnalyzeConfig_UnknownHTTPMethod(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
method: FOOBAR`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	// Should have a warning, not an error
	if a.HasErrors() {
		t.Error("expected warning, not error, for unknown HTTP method")
	}

	var found bool
	for _, d := range a.Diagnostics {
		if d.Severity == SeverityWarning && strings.Contains(d.Message, "unknown HTTP method") {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected warning about unknown HTTP method")
	}
}

func TestAnalyzeConfig_GraphQLWithBody(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com/graphql
graphql: query { foo }
body:
  key: value`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	if !a.HasErrors() {
		t.Fatal("expected error for graphql + body")
	}

	if !hasDiagnostic(a.Diagnostics, "cannot be used with") {
		t.Errorf("expected error about graphql/body conflict, got %+v", a.Diagnostics)
	}
}

func TestHasErrors(t *testing.T) {
	tests := []struct {
		name       string
		diags      []Diagnostic
		wantErrors bool
	}{
		{
			name:       "no diagnostics",
			diags:      nil,
			wantErrors: false,
		},
		{
			name: "only warnings",
			diags: []Diagnostic{
				{Severity: SeverityWarning, Message: "warning"},
			},
			wantErrors: false,
		},
		{
			name: "only info",
			diags: []Diagnostic{
				{Severity: SeverityInfo, Message: "info"},
			},
			wantErrors: false,
		},
		{
			name: "has errors",
			diags: []Diagnostic{
				{Severity: SeverityError, Message: "error"},
			},
			wantErrors: true,
		},
		{
			name: "mixed",
			diags: []Diagnostic{
				{Severity: SeverityWarning, Message: "warning"},
				{Severity: SeverityError, Message: "error"},
				{Severity: SeverityInfo, Message: "info"},
			},
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Analysis{Diagnostics: tt.diags}
			if got := a.HasErrors(); got != tt.wantErrors {
				t.Errorf("HasErrors() = %v, want %v", got, tt.wantErrors)
			}
		})
	}
}

func TestFindFieldLine(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
method: GET
graphql: |
  query { foo }
jq_filter: .data`

	tests := []struct {
		field    string
		wantLine int
	}{
		{"yapi", 0},
		{"url", 1},
		{"method", 2},
		{"graphql", 3},
		{"jq_filter", 5},
		{"nonexistent", -1},
		{"", -1},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got := findFieldLine(yaml, tt.field)
			if got != tt.wantLine {
				t.Errorf("findFieldLine(%q) = %d, want %d", tt.field, got, tt.wantLine)
			}
		})
	}
}

func TestFindFieldLine_EmptyText(t *testing.T) {
	got := findFieldLine("", "field")
	if got != -1 {
		t.Errorf("findFieldLine with empty text = %d, want -1", got)
	}
}

func TestAnalyzeConfig_UnknownKeys(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
method: GET
unknown_field: value
another_bad_key: 123`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	// Should not have errors - unknown keys are warnings
	if a.HasErrors() {
		t.Error("expected no errors for unknown keys")
	}

	// Should have warning diagnostics about unknown keys
	var unknownKeyDiags []Diagnostic
	for _, d := range a.Diagnostics {
		if d.Severity == SeverityWarning && strings.Contains(d.Message, "unknown key") {
			unknownKeyDiags = append(unknownKeyDiags, d)
		}
	}

	if len(unknownKeyDiags) < 2 {
		t.Errorf("expected at least 2 unknown key warnings, got %d", len(unknownKeyDiags))
	}

	// Check that line numbers are set correctly
	for _, d := range unknownKeyDiags {
		if d.Line < 0 {
			t.Errorf("expected line number for unknown key '%s', got %d", d.Field, d.Line)
		}
	}

	// Verify specific keys are detected
	if !hasDiagnostic(unknownKeyDiags, "unknown_field") {
		t.Errorf("expected warning about 'unknown_field', got %v", unknownKeyDiags)
	}

	if !hasDiagnostic(unknownKeyDiags, "another_bad_key") {
		t.Errorf("expected warning about 'another_bad_key', got %v", unknownKeyDiags)
	}
}

func TestAnalyzeConfig_NoUnknownKeys(t *testing.T) {
	yaml := `yapi: v1
url: http://example.com
method: GET
headers:
  Authorization: Bearer token`

	a, err := Analyze(yaml, AnalyzeOptions{})
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}

	// Should have no warnings about unknown keys
	for _, w := range a.Warnings {
		if strings.Contains(w, "unknown key") {
			t.Errorf("unexpected unknown key warning: %s", w)
		}
	}
}

// Tests for GraphQL variables vs environment variables detection

func TestFindEnvVarRefs_GraphQLVariablesNotDetected(t *testing.T) {
	// GraphQL $variables should NOT be detected as environment variables
	yaml := `yapi: v1
url: https://api.example.com/graphql
graphql: |
  query getUser($id: ID!, $includeEmail: Boolean) {
    user(id: $id) {
      name
      email @include(if: $includeEmail)
    }
  }
variables:
  id: "123"
  includeEmail: true`

	refs := FindEnvVarRefs(yaml)

	// Should find NO env var refs - all $vars are GraphQL variables
	for _, ref := range refs {
		t.Errorf("unexpected env var detected: $%s at line %d (should be recognized as GraphQL variable)", ref.Name, ref.Line)
	}
}

func TestFindEnvVarRefs_GraphQLBlockWithMultipleVariables(t *testing.T) {
	yaml := `yapi: v1
url: https://countries.trevorblades.com/graphql
graphql: |
  query getCountry($code: ID!) {
    country(code: $code) {
      name
      native
      capital
      currency
      languages {
        name
      }
    }
  }
variables:
  code: "GB"`

	refs := FindEnvVarRefs(yaml)

	// Should find no env vars - $code is a GraphQL variable
	if len(refs) > 0 {
		for _, ref := range refs {
			t.Errorf("GraphQL variable $%s incorrectly detected as env var at line %d", ref.Name, ref.Line)
		}
	}
}

func TestFindEnvVarRefs_MixedGraphQLAndEnvVars(t *testing.T) {
	yaml := `yapi: v1
url: ${API_URL}
headers:
  Authorization: Bearer ${TOKEN}
graphql: |
  query getUser($userId: ID!) {
    user(id: $userId) {
      name
    }
  }
variables:
  userId: ${USER_ID}`

	refs := FindEnvVarRefs(yaml)

	// Should find API_URL, TOKEN, and USER_ID but NOT userId (GraphQL variable)
	foundVars := make(map[string]bool)
	for _, ref := range refs {
		foundVars[ref.Name] = true
	}

	// These should be found (env vars)
	expectedEnvVars := []string{"API_URL", "TOKEN", "USER_ID"}
	for _, expected := range expectedEnvVars {
		if !foundVars[expected] {
			t.Errorf("expected env var $%s to be detected", expected)
		}
	}

	// These should NOT be found (GraphQL variables)
	unexpectedVars := []string{"userId"}
	for _, unexpected := range unexpectedVars {
		if foundVars[unexpected] {
			t.Errorf("GraphQL variable $%s should not be detected as env var", unexpected)
		}
	}
}

func TestFindEnvVarRefs_GraphQLInlineQuery(t *testing.T) {
	// Test with inline graphql (no multiline block indicator)
	yaml := `yapi: v1
url: https://api.example.com/graphql
graphql: "query { user(id: $id) { name } }"
variables:
  id: "123"`

	refs := FindEnvVarRefs(yaml)

	// The $id inside graphql should not be detected
	for _, ref := range refs {
		if ref.Name == "id" {
			t.Errorf("GraphQL variable $id incorrectly detected as env var")
		}
	}
}

func TestFindEnvVarRefs_GraphQLWithFoldedStyle(t *testing.T) {
	// Test with > folded style
	yaml := `yapi: v1
url: https://api.example.com/graphql
graphql: >
  query getUser($id: ID!) {
    user(id: $id) { name }
  }
variables:
  id: "123"`

	refs := FindEnvVarRefs(yaml)

	// Should not detect $id as env var
	for _, ref := range refs {
		if ref.Name == "id" {
			t.Errorf("GraphQL variable $id in folded block incorrectly detected as env var")
		}
	}
}

func TestFindEnvVarRefs_EnvVarAfterGraphQLBlock(t *testing.T) {
	yaml := `yapi: v1
url: https://api.example.com/graphql
graphql: |
  query { users { name } }
headers:
  X-Custom: ${CUSTOM_HEADER}`

	refs := FindEnvVarRefs(yaml)

	// Should find CUSTOM_HEADER after the graphql block ends
	found := false
	for _, ref := range refs {
		if ref.Name == "CUSTOM_HEADER" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected env var $CUSTOM_HEADER to be detected after graphql block")
	}
}

func TestFindEnvVarRefs_NoGraphQLBlock(t *testing.T) {
	yaml := `yapi: v1
url: ${BASE_URL}/api/users
method: GET
headers:
  Authorization: Bearer ${AUTH_TOKEN}
  X-Request-ID: ${REQUEST_ID}`

	refs := FindEnvVarRefs(yaml)

	expectedVars := []string{"BASE_URL", "AUTH_TOKEN", "REQUEST_ID"}
	foundVars := make(map[string]bool)
	for _, ref := range refs {
		foundVars[ref.Name] = true
	}

	for _, expected := range expectedVars {
		if !foundVars[expected] {
			t.Errorf("expected env var $%s to be detected", expected)
		}
	}

	if len(refs) != len(expectedVars) {
		t.Errorf("expected %d env vars, found %d", len(expectedVars), len(refs))
	}
}

func TestFindEnvVarRefs_BracedEnvVars(t *testing.T) {
	yaml := `yapi: v1
url: ${BASE_URL}/api
headers:
  Authorization: Bearer ${TOKEN}`

	refs := FindEnvVarRefs(yaml)

	expectedVars := []string{"BASE_URL", "TOKEN"}
	foundVars := make(map[string]bool)
	for _, ref := range refs {
		foundVars[ref.Name] = true
	}

	for _, expected := range expectedVars {
		if !foundVars[expected] {
			t.Errorf("expected braced env var ${%s} to be detected", expected)
		}
	}
}

func TestFindEnvVarRefs_GraphQLMutation(t *testing.T) {
	yaml := `yapi: v1
url: https://api.example.com/graphql
graphql: |
  mutation createUser($input: CreateUserInput!) {
    createUser(input: $input) {
      id
      name
    }
  }
variables:
  input:
    name: "Test User"
    email: "test@example.com"`

	refs := FindEnvVarRefs(yaml)

	// Should not detect $input as env var
	for _, ref := range refs {
		if ref.Name == "input" {
			t.Errorf("GraphQL variable $input incorrectly detected as env var")
		}
	}
}

func TestFindEnvVarRefs_GraphQLFragment(t *testing.T) {
	yaml := `yapi: v1
url: https://api.example.com/graphql
graphql: |
  fragment UserFields on User {
    id
    name
  }
  query getUser($id: ID!) {
    user(id: $id) {
      ...UserFields
    }
  }
variables:
  id: "123"`

	refs := FindEnvVarRefs(yaml)

	for _, ref := range refs {
		if ref.Name == "id" {
			t.Errorf("GraphQL variable $id incorrectly detected as env var")
		}
	}
}

func TestFindEnvVarRefs_NestedGraphQLVariables(t *testing.T) {
	yaml := `yapi: v1
url: https://api.example.com/graphql
graphql: |
  query search($query: String!, $first: Int, $after: String) {
    search(query: $query, first: $first, after: $after) {
      edges {
        node {
          ... on User {
            id
            name
          }
        }
        cursor
      }
      pageInfo {
        hasNextPage
      }
    }
  }
variables:
  query: "test"
  first: 10`

	refs := FindEnvVarRefs(yaml)

	graphqlVars := []string{"query", "first", "after"}
	for _, ref := range refs {
		for _, gqlVar := range graphqlVars {
			if ref.Name == gqlVar {
				t.Errorf("GraphQL variable $%s incorrectly detected as env var", gqlVar)
			}
		}
	}
}

func TestFindEnvFilesInConfig(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected []EnvFileInfo
	}{
		{
			name: "single env file",
			yaml: `yapi: v1
url: http://example.com
env_files:
  - .env.local`,
			expected: []EnvFileInfo{
				{Path: ".env.local", Line: 3, Col: 4},
			},
		},
		{
			name: "multiple env files",
			yaml: `yapi: v1
url: http://example.com
env_files:
  - .env.local
  - .env.prod
  - secrets.env`,
			expected: []EnvFileInfo{
				{Path: ".env.local", Line: 3, Col: 4},
				{Path: ".env.prod", Line: 4, Col: 4},
				{Path: "secrets.env", Line: 5, Col: 4},
			},
		},
		{
			name: "no env files",
			yaml: `yapi: v1
url: http://example.com`,
			expected: []EnvFileInfo{},
		},
		{
			name: "empty env files",
			yaml: `yapi: v1
url: http://example.com
env_files: []`,
			expected: []EnvFileInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindEnvFilesInConfig(tt.yaml)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d env files, got %d", len(tt.expected), len(result))
			}

			for i, exp := range tt.expected {
				if result[i].Path != exp.Path {
					t.Errorf("env file %d: expected path %q, got %q", i, exp.Path, result[i].Path)
				}
				if result[i].Line != exp.Line {
					t.Errorf("env file %d: expected line %d, got %d", i, exp.Line, result[i].Line)
				}
				if result[i].Col != exp.Col {
					t.Errorf("env file %d: expected col %d, got %d", i, exp.Col, result[i].Col)
				}
			}
		})
	}
}

func TestValidateEnvFilesExist(t *testing.T) {
	// Create a temp directory with a test env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env.exists")
	if err := os.WriteFile(envPath, []byte("VAR=value"), 0600); err != nil {
		t.Fatal(err)
	}

	// Create a config file in the temp directory
	configPath := filepath.Join(tmpDir, "test.yapi.yml")

	tests := []struct {
		name            string
		yaml            string
		expectedCount   int
		expectedMessage string
	}{
		{
			name: "existing file no diagnostics",
			yaml: `yapi: v1
url: http://example.com
env_files:
  - .env.exists`,
			expectedCount: 0,
		},
		{
			name: "missing file generates warning",
			yaml: `yapi: v1
url: http://example.com
env_files:
  - .env.missing`,
			expectedCount:   1,
			expectedMessage: ".env.missing",
		},
		{
			name: "mixed existing and missing",
			yaml: `yapi: v1
url: http://example.com
env_files:
  - .env.exists
  - .env.missing1
  - .env.missing2`,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := ValidateEnvFilesExist(tt.yaml, configPath)

			if len(diags) != tt.expectedCount {
				t.Errorf("expected %d diagnostics, got %d: %+v", tt.expectedCount, len(diags), diags)
			}

			if tt.expectedMessage != "" && len(diags) > 0 {
				if !strings.Contains(diags[0].Message, tt.expectedMessage) {
					t.Errorf("expected message containing %q, got %q", tt.expectedMessage, diags[0].Message)
				}
			}

			// Check that diagnostics have proper line numbers
			for _, d := range diags {
				if d.Line < 0 {
					t.Errorf("diagnostic has invalid line number: %d", d.Line)
				}
				if d.Severity != SeverityWarning {
					t.Errorf("expected warning severity, got %v", d.Severity)
				}
			}
		})
	}
}

func TestValidateEnvFilesExistFromProject(t *testing.T) {
	// Create a temp directory with test env files
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env.exists")
	if err := os.WriteFile(envPath, []byte("VAR=value"), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		yaml          string
		project       *config.ProjectConfigV1
		expectedCount int
	}{
		{
			name: "nil project returns no diagnostics",
			yaml: `yapi: v1
url: http://example.com
env_files:
  - .env.missing`,
			project:       nil,
			expectedCount: 0,
		},
		{
			name: "project with existing env file",
			yaml: `yapi: v1
url: http://example.com`,
			project: &config.ProjectConfigV1{
				Environments: map[string]config.Environment{
					"default": {
						ConfigV1: config.ConfigV1{
							EnvFiles: []string{".env.exists"},
						},
					},
				},
				DefaultEnvironment: "default",
			},
			expectedCount: 0,
		},
		{
			name: "project with missing env file",
			yaml: `yapi: v1
url: http://example.com`,
			project: &config.ProjectConfigV1{
				Environments: map[string]config.Environment{
					"default": {
						ConfigV1: config.ConfigV1{
							EnvFiles: []string{".env.missing"},
						},
					},
				},
				DefaultEnvironment: "default",
			},
			expectedCount: 1,
		},
		{
			name: "config-level env_files missing",
			yaml: `yapi: v1
url: http://example.com
env_files:
  - .env.config-missing`,
			project: &config.ProjectConfigV1{
				Environments: map[string]config.Environment{
					"default": {
						ConfigV1: config.ConfigV1{
							EnvFiles: []string{".env.exists"},
						},
					},
				},
				DefaultEnvironment: "default",
			},
			expectedCount: 1,
		},
		{
			name: "both project and config-level missing",
			yaml: `yapi: v1
url: http://example.com
env_files:
  - .env.config-missing`,
			project: &config.ProjectConfigV1{
				Environments: map[string]config.Environment{
					"default": {
						ConfigV1: config.ConfigV1{
							EnvFiles: []string{".env.project-missing"},
						},
					},
				},
				DefaultEnvironment: "default",
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := ValidateEnvFilesExistFromProject(tt.yaml, tt.project, tmpDir, "")

			if len(diags) != tt.expectedCount {
				t.Errorf("expected %d diagnostics, got %d: %+v", tt.expectedCount, len(diags), diags)
			}

			// Check that all diagnostics are warnings
			for _, d := range diags {
				if d.Severity != SeverityWarning {
					t.Errorf("expected warning severity, got %v", d.Severity)
				}
			}
		})
	}
}

func TestAnalyzeWithOptions(t *testing.T) {
	// Create temp directory with env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env.test")
	if err := os.WriteFile(envPath, []byte("TEST_VAR=test-value"), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		yaml       string
		opts       AnalyzeOptions
		wantErrors bool
	}{
		{
			name: "basic config without project",
			yaml: `yapi: v1
url: http://example.com
method: GET`,
			opts:       AnalyzeOptions{FilePath: filepath.Join(tmpDir, "test.yapi.yml")},
			wantErrors: false,
		},
		{
			name: "project with no default environment picks first",
			yaml: `yapi: v1
url: http://example.com
method: GET`,
			opts: AnalyzeOptions{
				FilePath: filepath.Join(tmpDir, "test.yapi.yml"),
				Project: &config.ProjectConfigV1{
					Environments: map[string]config.Environment{
						"dev": {
							ConfigV1: config.ConfigV1{
								EnvFiles: []string{".env.test"},
							},
						},
					},
					DefaultEnvironment: "", // No default set
				},
				ProjectRoot: tmpDir,
			},
			wantErrors: false,
		},
		{
			name: "project with default environment",
			yaml: `yapi: v1
url: http://example.com
method: GET`,
			opts: AnalyzeOptions{
				FilePath: filepath.Join(tmpDir, "test.yapi.yml"),
				Project: &config.ProjectConfigV1{
					Environments: map[string]config.Environment{
						"prod": {
							ConfigV1: config.ConfigV1{
								EnvFiles: []string{".env.test"},
							},
						},
					},
					DefaultEnvironment: "prod",
				},
				ProjectRoot: tmpDir,
			},
			wantErrors: false,
		},
		{
			name: "invalid YAML returns error diagnostic",
			yaml: `yapi: v1
url: [invalid yaml`,
			opts:       AnalyzeOptions{FilePath: filepath.Join(tmpDir, "test.yapi.yml")},
			wantErrors: true,
		},
		{
			name: "strict env mode with missing env file",
			yaml: `yapi: v1
url: http://example.com
env_files:
  - .env.nonexistent`,
			opts:       AnalyzeOptions{FilePath: filepath.Join(tmpDir, "test.yapi.yml"), StrictEnv: true},
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := Analyze(tt.yaml, tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if analysis == nil {
				t.Fatal("expected analysis, got nil")
			}
			if analysis.HasErrors() != tt.wantErrors {
				t.Errorf("HasErrors() = %v, want %v. Diagnostics: %+v",
					analysis.HasErrors(), tt.wantErrors, analysis.Diagnostics)
			}
		})
	}
}

func TestAnalyzeConfigFileWithOptions(t *testing.T) {
	// Create temp directory with valid config
	tmpDir := t.TempDir()
	validConfig := `yapi: v1
url: http://example.com
method: GET`
	validPath := filepath.Join(tmpDir, "valid.yapi.yml")
	if err := os.WriteFile(validPath, []byte(validConfig), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		path       string
		opts       AnalyzeOptions
		wantErrors bool
	}{
		{
			name:       "valid config file",
			path:       validPath,
			opts:       AnalyzeOptions{},
			wantErrors: false,
		},
		{
			name:       "nonexistent file",
			path:       filepath.Join(tmpDir, "nonexistent.yapi.yml"),
			opts:       AnalyzeOptions{},
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := AnalyzeConfigFileWithOptions(tt.path, tt.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if analysis == nil {
				t.Fatal("expected analysis, got nil")
			}
			if analysis.HasErrors() != tt.wantErrors {
				t.Errorf("HasErrors() = %v, want %v", analysis.HasErrors(), tt.wantErrors)
			}
		})
	}
}

func TestAnalyzeConfigFile(t *testing.T) {
	// Create temp directory with valid config
	tmpDir := t.TempDir()
	validConfig := `yapi: v1
url: http://example.com
method: GET`
	validPath := filepath.Join(tmpDir, "valid.yapi.yml")
	if err := os.WriteFile(validPath, []byte(validConfig), 0600); err != nil {
		t.Fatal(err)
	}

	analysis, err := AnalyzeConfigFile(validPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if analysis == nil {
		t.Fatal("expected analysis, got nil")
	}
	if analysis.HasErrors() {
		t.Errorf("expected no errors, got: %+v", analysis.Diagnostics)
	}
}

func TestExtractLineFromError(t *testing.T) {
	tests := []struct {
		errMsg   string
		expected int
	}{
		{"line 22: cannot unmarshal", 21}, // 0-indexed
		{"line 1: error", 0},
		{"no line number here", -1},
		{"", -1},
	}

	for _, tt := range tests {
		t.Run(tt.errMsg, func(t *testing.T) {
			got := extractLineFromError(tt.errMsg)
			if got != tt.expected {
				t.Errorf("extractLineFromError(%q) = %d, want %d", tt.errMsg, got, tt.expected)
			}
		})
	}
}
